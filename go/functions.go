package main

type function struct {
	name          string
	start         *parsepoint
	module        *hmfile
	forClass      *class
	args          []*funcArg
	argDict       map[string]int
	aliasing      map[string]string
	expressions   []*node
	returns       *varData
	generics      map[string]int
	genericsOrder []string
	genericsAlias map[string]string
	base          *function
	impls         []*function
}

func funcInit(module *hmfile, name string) *function {
	f := &function{}
	f.module = module
	f.name = name
	f.args = make([]*funcArg, 0)
	f.argDict = make(map[string]int)
	f.expressions = make([]*node, 0)
	f.aliasing = make(map[string]string)
	return f
}

func (me *function) copy() *function {
	f := &function{}
	f.module = me.module
	f.name = me.name
	f.args = make([]*funcArg, len(me.args))
	for i, a := range me.args {
		f.args[i] = a.copy()
	}
	f.argDict = me.argDict
	f.expressions = make([]*node, len(me.expressions))
	for i, e := range me.expressions {
		f.expressions[i] = e.copy()
	}
	f.returns = me.returns.copy()
	return f
}

func (me *function) canonical() string {
	name := ""
	if me.forClass != nil {
		name = me.nameOfClassFunc()
	} else {
		name = me.name
	}
	if me.module != nil {
		name = me.module.name + "." + name
	}
	return name
}

func (me *function) asSig() *fnSig {
	sig := fnSigInit(me.module)
	for _, arg := range me.args {
		sig.args = append(sig.args, arg)
	}
	sig.returns = me.returns
	return sig
}

func (me *function) data() *varData {
	return me.asSig().data()
}

func nameOfClassFunc(cl, fn string) string {
	return cl + "_" + fn
}

func (me *function) nameOfClassFunc() string {
	return nameOfClassFunc(me.forClass.name, me.name)
}

func (me *parser) pushFunction(name string, module *hmfile, fn *function) {
	module.functionOrder = append(module.functionOrder, name)
	module.functions[name] = fn
	module.types[name] = ""
	if me.file != nil {
		me.file.WriteString(fn.string(0))
	}
}

func (me *parser) remapClassFunctionImpl(class *class, original *function) {
	funcName := original.name
	pos := me.save()
	me.jump(original.start)
	fn := me.defineFunction(funcName, class)
	me.jump(pos)
	fn.start = original.start
	fn.forClass = class
	class.functions[fn.name] = fn
	class.functionOrder = append(class.functionOrder, fn)
	me.pushFunction(fn.nameOfClassFunc(), me.hmfile, fn)
}

func (me *parser) defineClassFunction() {
	module := me.hmfile
	className := me.token.value
	class := module.classes[className]
	me.eat("id")
	funcName := me.token.value
	globalFuncName := nameOfClassFunc(class.name, funcName)
	me.eat("id")
	if _, ok := module.functions[globalFuncName]; ok {
		panic(me.fail() + "class \"" + className + "\" with function \"" + funcName + "\" is already defined")
	}
	if _, ok := class.variables[funcName]; ok {
		panic(me.fail() + "class \"" + className + "\" with variable \"" + funcName + "\" is already defined")
	}
	start := me.save()
	fn := me.defineFunction(funcName, class)
	fn.start = start
	fn.forClass = class
	class.functions[funcName] = fn
	class.functionOrder = append(class.functionOrder, fn)
	me.pushFunction(globalFuncName, module, fn)
	for _, impl := range class.impls {
		me.remapClassFunctionImpl(impl, fn)
	}
}

func (me *parser) defineFileFunction() {
	module := me.hmfile
	token := me.token
	name := token.value
	if _, ok := module.functions[name]; ok {
		panic(me.fail() + "function \"" + name + "\" is already defined")
	}
	me.eat("id")
	fn := me.defineFunction(name, nil)
	me.pushFunction(name, module, fn)
}

func (me *parser) defineFunction(name string, self *class) *function {
	fn := funcInit(me.hmfile, name)
	me.hmfile.pushScope()
	me.hmfile.scope.fn = fn
	if self != nil {
		ref := me.hmfile.fnArgInit(self.name, "self", false)
		fn.args = append(fn.args, ref)
		fn.aliasing = self.gmapper
	}
	parenthesis := false
	if me.token.is == "<" {
		if self != nil {
			panic(me.fail() + "class functions cannot have additional generics")
		}
		order, dict := me.genericHeader()
		me.verify("(")
		fn.generics = dict
		fn.genericsOrder = order
	}
	if me.token.is == "(" {
		me.eat("(")
		parenthesis = true
		if me.token.is != ")" {
			for {
				if me.token.is != "id" {
					panic(me.fail() + "unexpected token in function definition")
				}
				argname := me.token.value
				me.eat("id")
				defaultValue := ""
				defaultType := ""
				if me.token.is == ":" {
					me.eat(":")
					op := me.token.is
					if literal, ok := literals[op]; ok {
						defaultValue = me.token.value
						defaultType = literal
						me.eat(op)
					} else {
						panic(me.fail() + "only primitive literals allowed for parameter defaults. was \"" + me.token.is + "\"")
					}
				}
				typed := me.declareType(true)
				fn.argDict[argname] = len(fn.args)
				fnArg := &funcArg{}
				fnArg.variable = me.hmfile.varInitFromData(typed, argname, false)
				if defaultValue != "" {
					defaultTypeVarData := typeToVarData(me.hmfile, defaultType)
					if typed.notEqual(defaultTypeVarData) {
						panic(me.fail() + "function parameter default type \"" + defaultType + "\" and signature \"" + typed.full + "\" do not match")
					}
					defaultNode := nodeInit(defaultTypeVarData.full)
					defaultNode.copyData(defaultTypeVarData)
					defaultNode.value = defaultValue
					fnArg.defaultNode = defaultNode
				}
				fn.args = append(fn.args, fnArg)
				if me.token.is == ")" {
					break
				} else {
					me.eat(",")
				}
			}
		}
		me.eat(")")
	} else {
		if self != nil {
			panic(me.fail() + "class function \"" + name + "\" must include parenthesis")
		}
	}
	if me.token.is != "line" {
		if !parenthesis {
			panic(me.fail() + "function \"" + name + "\" returns a value and must include parenthesis")
		}
		fn.returns = me.declareType(true)
	} else {
		fn.returns = typeToVarData(me.hmfile, "void")
	}
	me.eat("line")
	for _, arg := range fn.args {
		me.hmfile.scope.variables[arg.name] = arg.variable
	}
	for {
		for me.token.is == "line" {
			me.eat("line")
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
		if me.token.is == "comment" {
			me.eat("comment")
		}
		expr := me.expression()
		fn.expressions = append(fn.expressions, expr)
	}
fnEnd:
	me.hmfile.popScope()
	return fn
}
