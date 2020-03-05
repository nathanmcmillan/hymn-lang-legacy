package main

import "fmt"

type function struct {
	_name         string
	_clsname      string
	_cname        string
	start         *parsepoint
	module        *hmfile
	forClass      *class
	args          []*funcArg
	argVariadic   *funcArg
	aliasing      map[string]string
	expressions   []*node
	returns       *datatype
	generics      []string
	mapping       map[string]*datatype
	interfaces    map[string][]*classInterface
	genericsAlias map[string]string
	base          *function
	impls         []*function
	comments      []string
}

func funcInit(module *hmfile, name string, class *class) *function {
	f := &function{}
	f._name = name
	if class != nil {
		f._clsname = nameOfClassFunc(class.name, name)
	}
	if module != nil {
		f.module = module
		if class == nil {
			f._cname = module.funcNameSpace(name)
		} else {
			f._cname = module.funcNameSpace(f._clsname)
		}
	}
	f.args = make([]*funcArg, 0)
	f.expressions = make([]*node, 0)
	f.aliasing = make(map[string]string)
	f.forClass = class
	return f
}

func (me *function) getname() string {
	if me.forClass == nil {
		return me._name
	}
	return me._clsname
}

func (me *function) getcname() string {
	return me._cname
}

func (me *function) getclsname() string {
	return me._clsname
}

func (me *function) canonical(current *hmfile) string {
	name := me.getname()
	if me.module != nil {
		return me.module.reference(name)
	}
	return name
}

func (me *function) asSig() *fnSig {
	sig := fnSigInit(me.module)
	for _, arg := range me.args {
		sig.args = append(sig.args, arg)
	}
	sig.argVariadic = me.argVariadic
	sig.returns = me.returns
	return sig
}

func (me *function) data() (*datatype, *parseError) {
	return me.asSig().newdatatype()
}

func nameOfClassFunc(cl, fn string) string {
	return cl + "_" + fn
}

func (me *parser) pushFunction(name string, module *hmfile, fn *function) {
	module.functionOrder = append(module.functionOrder, name)
	module.functions[name] = fn
	module.types[name] = "function"
	if me.file != nil {
		me.file.WriteString(fn.string(me.hmfile, 0))
	}
}

func remapFunctionImpl(name string, mapping map[string]*datatype, original *function) (*function, *parseError) {
	module := original.module
	if fn, ok := module.functions[name]; ok {
		return fn, nil
	}
	parsing := module.parser
	module.program.pushRemapStack(module.reference(original.getname()))
	pos := parsing.save()
	parsing.jump(original.start)

	fn, er := parsing.defineFunction(name, mapping, original, nil)
	if er != nil {
		return nil, er
	}

	if len(original.interfaces) > 0 {
		for _, g := range original.generics {
			i, ok := original.interfaces[g]
			if !ok {
				continue
			}
			m := fn.mapping[g]
			if cl, ok := m.isClass(); ok {
				for _, t := range i {
					if _, ok := cl.selfInterfaces[t.uid()]; !ok {
						panic(parsing.fail() + "Class '" + cl.name + "' for function '" + name + "' requires interface '" + t.name + "'")
					}
				}
			} else {
				panic(parsing.fail() + "Function '" + name + "' requires interface implementation but type was " + m.error())
			}
		}
	}

	parsing.jump(pos)
	fn.start = original.start
	parsing.pushFunction(fn.getname(), module, fn)
	module.program.popRemapStack()
	return fn, nil
}

func remapClassFunctionImpl(class *class, original *function) *parseError {
	module := class.module
	plain := original._name
	registered := nameOfClassFunc(class.name, plain)
	if _, ok := module.functions[registered]; ok {
		return nil
	}
	parsing := module.parser
	module.program.pushRemapStack(module.reference(registered))
	pos := parsing.save()
	parsing.jump(original.start)
	fn, er := parsing.defineFunction(plain, nil, nil, class)
	if er != nil {
		return er
	}
	parsing.jump(pos)
	fn.start = original.start
	class.functions = append(class.functions, fn)
	parsing.pushFunction(fn.getname(), module, fn)
	module.program.popRemapStack()
	return er
}

func (me *parser) defineClassFunction() *parseError {
	module := me.hmfile
	className := me.token.value
	class := module.classes[className]
	if er := me.eat("id"); er != nil {
		return er
	}
	funcName := me.token.value
	globalFuncName := nameOfClassFunc(class.name, funcName)
	if er := me.eat("id"); er != nil {
		return er
	}
	if _, ok := module.functions[globalFuncName]; ok {
		return err(me, ECodeNameAlreadyDefined, "Class '"+className+"' with function '"+funcName+"' is already defined")
	} else if class.getVariable(funcName) != nil {
		return err(me, ECodeNameAlreadyDefined, "Class '"+className+"' with variable '"+funcName+"' is already defined")
	}
	fn, er := me.defineFunction(funcName, nil, nil, class)
	if er != nil {
		return er
	}
	class.functions = append(class.functions, fn)
	me.pushFunction(globalFuncName, module, fn)
	for _, implementation := range class.implementations {
		remapClassFunctionImpl(implementation, fn)
	}
	return nil
}

func (me *parser) defineStaticFunction() *parseError {
	module := me.hmfile
	token := me.token
	name := token.value
	if _, ok := module.functions[name]; ok {
		return err(me, ECodeNameConflict, "Function \""+name+"\" is already defined.")
	}
	if name == "static" {
		return err(me, ECodeReservedName, "Function \""+name+"\" is reserved.")
	}
	if er := me.eat("id"); er != nil {
		return er
	}
	fn, er := me.defineFunction(name, nil, nil, nil)
	if er != nil {
		return er
	}
	me.pushFunction(name, module, fn)
	return nil
}

func (me *parser) defineFunction(name string, mapping map[string]*datatype, base *function, self *class) (*function, *parseError) {
	module := me.hmfile
	if base != nil {
		module = base.module
	} else if self != nil {
		module = self.module
	}
	fn := funcInit(module, name, self)
	if len(me.hmfile.comments) > 0 {
		fn.comments = me.hmfile.comments
		me.hmfile.comments = make([]string, 0)
	}
	module.pushScope()
	module.scope.fn = fn
	fname := name
	if base != nil {
		fn.base = base
		base.impls = append(base.impls, fn)
		alias := make(map[string]string)
		for k, v := range mapping {
			alias[k] = v.print()
		}
		fn.mapping = mapping
		fn.aliasing = alias
	} else if self != nil {
		ref, er := module.fnArgInit(self.uid(), "self", false)
		if er != nil {
			return nil, er
		}
		fn.args = append(fn.args, ref)
		alias := make(map[string]string)
		for k, v := range self.mapping {
			alias[k] = v.print()
		}
		fn.mapping = mapping
		fn.aliasing = alias
		fn.interfaces = self.genericsInterfaces
		fname = self.name + "." + name
	}
	if me.token.is == "<" {
		if self != nil {
			return nil, err(me, ECodeNoAdditionalGenerics, "class functions cannot have additional generics")
		}
		if base != nil {
			return nil, err(me, ECodeNoAdditionalGenerics, "implementation of static function cannot have additional generics")
		}
		order, interfaces, er := me.genericHeader()
		if er != nil {
			return nil, er
		}
		if er := me.verify("("); er != nil {
			return nil, er
		}
		fn.generics = datatypels(order)
		fn.interfaces = interfaces
	}
	fn.start = me.save()
	parenthesis := false
	if me.token.is == "(" {
		if er := me.eat("("); er != nil {
			return nil, er
		}
		if me.token.is == "line" {
			if er := me.eat("line"); er != nil {
				return nil, er
			}
		}
		parenthesis = true
		if me.token.is != ")" {
			for {
				if me.token.is != "id" {
					return nil, err(me, ECodeUnexpectedToken, "Unexpected token in function definition")
				}
				argname := me.token.value
				if er := me.eat("id"); er != nil {
					return nil, er
				}
				defaultValue := ""
				defaultType := ""
				if me.token.is == ":" {
					if er := me.eat(":"); er != nil {
						return nil, er
					}
					op := me.token.is
					if literal, ok := literals[op]; ok {
						defaultValue = me.token.value
						defaultType = literal
						if er := me.eat(op); er != nil {
							return nil, er
						}
					} else {
						return nil, err(me, ECodeOnlyPrimitiveLitreralsAllowed, "only primitive literals allowed for parameter defaults. was \""+me.token.is+"\"")
					}
				}
				typed, er := me.declareType()
				if er != nil {
					return nil, er
				}
				fnArg := &funcArg{}
				fnArg.variable = typed.getnamedvariable(argname, false)
				if defaultValue != "" {
					defaultTypeVarData, er := getdatatype(module, defaultType)
					if er != nil {
						return nil, er
					}
					if typed.notEquals(defaultTypeVarData) {
						return nil, err(me, ECodeFunctionAndSignatureMismatch, "function parameter default type \""+defaultType+"\" and signature \""+typed.print()+"\" do not match")
					}
					defaultNode := nodeInit(defaultTypeVarData.getRaw())
					defaultNode.copyData(defaultTypeVarData)
					defaultNode.value = defaultValue
					fnArg.defaultNode = defaultNode
				}
				fn.args = append(fn.args, fnArg)
				if me.token.is == ")" {
					break
				} else if me.token.is == "line" {
					if er := me.eat("line"); er != nil {
						return nil, er
					}
				} else {
					if er := me.eat(","); er != nil {
						return nil, er
					}
				}
			}
		}
		if er := me.eat(")"); er != nil {
			return nil, er
		}
	} else {
		if self != nil {
			return nil, err(me, ECodeFunctionMissingParenthesis, "class function \""+fname+"\" must include parenthesis")
		}
	}
	// TODO ::
	// if fn.generics == nil && fn.interfaces == nil {
	// 	order, interfaces := me.withGenericsHeader()
	// 	fn.generics = datatypels(order)
	// 	fn.interfaces = interfaces
	// 	fmt.Println(name, fn.generics, fn.interfaces)
	// }
	if me.token.is != "line" {
		if !parenthesis {
			return nil, err(me, ECodeFunctionMissingParenthesis, "function \""+name+"\" returns a value and must include parenthesis")
		}
		var er *parseError
		fn.returns, er = me.declareType()
		if er != nil {
			return nil, er
		}
	} else {
		fn.returns = newdatavoid()
	}
	if er := me.eat("line"); er != nil {
		return nil, er
	}

	if self == nil && name == "main" {
		if len(fn.args) != 0 {
			if len(fn.args) == 1 {
				if !fn.args[0].data().isSlice() || !fn.args[0].data().member.isString() {
					return nil, err(me, ECodeFunctionMainSignature, "Function main argument must be []string")
				}
			} else {
				return nil, err(me, ECodeFunctionMainSignature, "Function main cannot have more than one argument.")
			}
		}
		if fn.returns != nil && !fn.returns.isInt() && !fn.returns.isVoid() {
			return nil, err(me, ECodeFunctionMainSignature, "Function main must return an integer but was: "+fn.returns.error())
		}
	}

	for _, arg := range fn.args {
		module.scope.variables[arg.name] = arg.variable
	}

	for {
		for me.token.is == "line" {
			if er := me.eat("line"); er != nil {
				return nil, er
			}
			if me.token.is != "line" {
				if me.token.depth != 1 {
					goto fnEnd
				}
				break
			}
		}
		if me.token.depth == 0 {
			goto fnEnd
		}
		expr, er := me.expression()
		if er != nil {
			return nil, er
		}
		fn.expressions = append(fn.expressions, expr)
	}
fnEnd:
	for _, arg := range fn.args {
		if !arg.used {
			e := fmt.Sprintf("I found the variable `%s` for function `%s` is unused.", arg.name, fname)
			h := ""
			if arg.name == "self" {
				h += " This is a class function. Can it be a static?"
			}
			return nil, errh(me, ECodeUnusedVariable, e, h)
		}
	}
	module.popScope()
	return fn, nil
}

func (me *function) hasInterface(data *datatype) bool {
	_, ok := me.interfaces[data.print()]
	return ok
}

func (me *function) searchInterface(data *datatype, name string) (*classInterface, *fnSig, bool) {
	return searchInterface(me.interfaces[data.print()], name)
}

func searchInterface(interfaces []*classInterface, name string) (*classInterface, *fnSig, bool) {
	for _, def := range interfaces {
		for fname, fn := range def.functions {
			if name == fname {
				return def, fn, true
			}
		}
	}
	return nil, nil, false
}

func (me *function) getParameter(name string) *funcArg {
	_, p := getParameter(me.args, name)
	return p
}

func getParameter(parameters []*funcArg, name string) (int, *funcArg) {
	for i, p := range parameters {
		if name == p.name {
			return i, p
		}
	}
	return -1, nil
}
