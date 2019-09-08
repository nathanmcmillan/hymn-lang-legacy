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

func (me *function) asSig() *fnSig {
	sig := fnSigInit(me.module)
	for _, arg := range me.args {
		sig.args = append(sig.args, arg)
	}
	sig.typed = me.typed
	return sig
}

func (me *function) asVar() *varData {
	return me.asSig().asVar()
}

func (me *parser) pushFunction(name string, module *hmfile, fn *function) {
	module.functionOrder = append(module.functionOrder, name)
	module.functions[name] = fn
	if me.file != nil {
		me.file.WriteString(fn.dump(0))
	}
}

func (me *parser) defineClassFunction() {
	module := me.hmfile
	className := me.token.value
	me.eat("id")
	me.eat(".")
	funcName := me.token.value
	globalFuncName := nameOfClassFunc(className, funcName)
	me.eat("id")
	class := module.classes[className]
	if _, ok := module.functions[globalFuncName]; ok {
		panic(me.fail() + "class \"" + className + "\" with function \"" + funcName + "\" is already defined")
	}
	if _, ok := class.variables[funcName]; ok {
		panic(me.fail() + "class \"" + className + "\" with variable \"" + funcName + "\" is already defined")
	}
	fn := me.defineFunction(funcName, class)
	fn.forClass = class
	me.pushFunction(globalFuncName, module, fn)
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
		ref := me.hmfile.fnArgInit(self.name, "self", false, true)
		fn.args = append(fn.args, ref)
	}
	if me.token.is == "(" {
		me.eat("(")
		if me.token.is != ")" {
			for {
				if me.token.is == "id" {
					argname := me.token.value
					me.eat("id")
					defaultValue := ""
					defaultType := ""
					if me.token.is == ":" {
						me.eat(":")
						op := me.token.is
						if _, ok := primitives[op]; ok {
							defaultValue = me.token.value
							defaultType = op
							me.eat(op)
						} else {
							panic(me.fail() + "only primitives allowed for parameter defaults. was \"" + me.token.is + "\"")
						}
					}
					typed := me.declareType(true)
					fn.argDict[argname] = len(fn.args)
					fnArg := &funcArg{}
					fnArg.variable = me.hmfile.varInitFromData(typed, argname, false, true)
					if defaultValue != "" {
						defaultTypeVarData := me.hmfile.typeToVarData(defaultType)
						if typed.notEqual(defaultTypeVarData) {
							panic(me.fail() + "function parameter default type \"" + defaultType + "\" and signature \"" + typed.full + "\" do not match")
						}
						defaultNode := nodeInit(defaultTypeVarData.full)
						defaultNode.vdata = defaultTypeVarData
						defaultNode.value = defaultValue
						fnArg.defaultNode = defaultNode
					}
					fn.args = append(fn.args, fnArg)
					if me.token.is == ")" {
						break
					} else if me.token.is == "delim" {
						me.eat("delim")
						continue
					}
				}
				panic(me.fail() + "unexpected token in function definition")
			}
		}
		me.eat(")")
	}
	if me.token.is != "line" {
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
		if expr.is == "return" {
			if fn.typed.notEqual(expr.asVar()) {
				panic(me.fail() + "function " + name + " returns " + fn.typed.full + " but found " + expr.getType())
			}
			goto fnEnd
		}
	}
fnEnd:
	me.hmfile.popScope()
	return fn
}
