package main

type function struct {
	name        string
	module      *hmfile
	forClass    *class
	args        []*funcArg
	argDict     map[string]int
	expressions []*node
	typed       *varData
}

func funcInit(module *hmfile, name string) *function {
	f := &function{}
	f.module = module
	f.name = name
	f.args = make([]*funcArg, 0)
	f.argDict = make(map[string]int)
	f.expressions = make([]*node, 0)
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
	f.typed = me.typed.copy()
	return f
}

func (me *function) canonical() string {
	name := me.name
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
	sig.typed = me.typed
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

func (me *parser) remapNodeRecursive(impl *class, n *node) {
	if n.data() != nil {
		n.data().genericReplace(impl.gmapper)
	}
	for _, h := range n.has {
		me.remapNodeRecursive(impl, h)
	}
}

func (me *parser) remapClassFunctionImpl(classImpl *class, original *function) {
	fn := original.copy()
	fn.forClass = classImpl
	for i, a := range fn.args {
		if i == 0 {
			fn.args[0] = me.hmfile.fnArgInit(classImpl.name, "self", false)
		} else {
			a.data().genericReplace(classImpl.gmapper)
		}
	}
	fn.typed.genericReplace(classImpl.gmapper)
	for _, e := range fn.expressions {
		me.remapNodeRecursive(classImpl, e)
	}
	classImpl.functions[fn.name] = fn
	classImpl.functionOrder = append(classImpl.functionOrder, fn)
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
	fn := me.defineFunction(funcName, class)
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
	if self != nil {
		ref := me.hmfile.fnArgInit(self.name, "self", false)
		fn.args = append(fn.args, ref)
	}
	parenthesis := false
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
					defaultTypeVarData := me.hmfile.typeToVarData(defaultType)
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
	}
	if me.token.is != "line" {
		if !parenthesis {
			panic(me.fail() + "functions that return must include parenthesis")
		}
		fn.typed = me.declareType(true)
	} else {
		fn.typed = me.hmfile.typeToVarData("void")
	}
	me.eat("line")
	me.hmfile.pushScope()
	me.hmfile.scope.fn = fn
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
