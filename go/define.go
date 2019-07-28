package main

func (me *parser) defineClassFunction() {
	module := me.hmfile
	className := me.token.value
	me.eat("id")
	me.eat(".")
	funcName := me.token.value
	globalFuncName := me.nameOfClassFunc(className, funcName)
	me.eat("id")
	class := module.classes[className]
	if _, ok := module.functions[globalFuncName]; ok {
		panic(me.fail() + "class \"" + className + "\" with function \"" + funcName + "\" is already defined")
	}
	if _, ok := class.variables[funcName]; ok {
		panic(me.fail() + "class \"" + className + "\" with variable \"" + funcName + "\" is already defined")
	}
	fn := me.defineFunction(funcName, class)
	module.functionOrder = append(module.functionOrder, globalFuncName)
	module.functions[globalFuncName] = fn
}

func (me *parser) defineFileFunction() {
	program := me.hmfile
	token := me.token
	name := token.value
	if _, ok := program.functions[name]; ok {
		panic(me.fail() + "function \"" + name + "\" is already defined")
	}
	me.eat("id")
	fn := me.defineFunction(name, nil)
	program.functionOrder = append(program.functionOrder, name)
	program.functions[name] = fn
}

func (me *parser) defineFunction(name string, self *class) *function {
	fn := funcInit()
	fn.name = name
	fn.typed = me.hmfile.typeToVarData("void")
	if self != nil {
		ref := me.hmfile.varInit(self.name, "self", false, true)
		fn.args = append(fn.args, ref)
	}
	if me.token.is == "(" {
		me.eat("(")
		if me.token.is != ")" {
			for {
				if me.token.is == "id" {
					argname := me.token.value
					me.eat("id")
					dval := ""
					dtype := ""
					if me.token.is == ":" {
						me.eat(":")
						op := me.token.is
						if _, ok := primitives[op]; ok {
							dval = me.token.value
							dtype = op
							me.eat(op)
						} else {
							panic(me.fail() + "only primitives allowed for parameter defaults. was \"" + me.token.is + "\"")
						}
					}
					typed := me.declareType(true)
					if dval != "" {
						if typed.notEqual(me.hmfile.typeToVarData(dtype)) {
							panic(me.fail() + "function parameter default type \"" + dtype + "\" and signature \"" + typed.full + "\" do not match")
						}
					}
					fn.argDict[argname] = len(fn.args)
					fn.args = append(fn.args, me.hmfile.varWithDefaultInit(typed.full, argname, false, true, dval))
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
	}
	me.eat("line")
	me.hmfile.pushScope()
	me.hmfile.scope.fn = fn
	for _, arg := range fn.args {
		me.hmfile.scope.variables[arg.name] = arg
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
		if me.token.is == "#" {
			me.eat("#")
		}
		expr := me.expression()
		fn.expressions = append(fn.expressions, expr)
		if expr.is == "return" {
			if fn.typed.notEqual(me.hmfile.typeToVarData(expr.typed)) {
				panic(me.fail() + "function " + name + " returns " + fn.typed.full + " but found " + expr.typed)
			}
			goto fnEnd
		}
	}
fnEnd:
	me.hmfile.popScope()
	return fn
}

func (me *parser) genericHeader() ([]string, map[string]bool) {
	order := make([]string, 0)
	dict := make(map[string]bool, 0)
	if me.token.is == "<" {
		me.eat("<")
		for {
			gname := me.token.value
			me.eat("id")
			dict[gname] = true
			order = append(order, gname)
			if me.token.is == "delim" {
				me.eat("delim")
				continue
			}
			if me.token.is == ">" {
				break
			}
			panic(me.fail() + "bad token \"" + me.token.is + "\" in class generic")
		}
		me.eat(">")
	}
	return order, dict
}

func (me *parser) defineClass() {
	me.eat("type")
	token := me.token
	name := token.value
	if _, ok := me.hmfile.namespace[name]; ok {
		panic(me.fail() + "name \"" + name + "\" already defined")
	}
	me.eat("id")

	genericsOrder, genericsDict := me.genericHeader()
	me.eat("line")

	me.hmfile.namespace[name] = "type"
	me.hmfile.types[name] = true
	me.hmfile.defineOrder = append(me.hmfile.defineOrder, name+"_type")

	classDef := classInit(name, genericsOrder, genericsDict)
	me.hmfile.classes[name] = classDef

	memberOrder := make([]string, 0)
	memberMap := make(map[string]*variable)

	for {
		if me.token.is == "line" {
			break
		}
		if me.token.is == "eof" || me.token.is == "#" {
			break
		}
		if me.token.is == "id" {
			mname := me.token.value
			me.eat("id")
			if _, ok := memberMap[mname]; ok {
				panic(me.fail() + "member name \"" + mname + "\" already used")
			}
			if _, ok := genericsDict[mname]; ok {
				panic(me.fail() + "cannot use \"" + mname + "\" as member name")
			}
			mtype := me.declareType(false)
			me.eat("line")
			memberOrder = append(memberOrder, mname)
			memberMap[mname] = me.hmfile.varInit(mtype.full, mname, true, true)
			continue
		}
		panic(me.fail() + "bad token \"" + token.is + "\" in class")
	}

	classDef.initMembers(memberOrder, memberMap)
}

func (me *parser) defineEnum() {
	me.eat("enum")
	token := me.token
	name := token.value
	if _, ok := me.hmfile.namespace[name]; ok {
		panic(me.fail() + "name \"" + name + "\" already defined")
	}
	me.eat("id")
	genericsOrder, genericsDict := me.genericHeader()
	me.eat("line")

	me.hmfile.namespace[name] = "enum"
	me.hmfile.types[name] = true
	me.hmfile.defineOrder = append(me.hmfile.defineOrder, name+"_enum")

	typesOrder := make([]*union, 0)
	typesMap := make(map[string]*union)
	isSimple := true
	for {
		token := me.token
		if token.is == "line" {
			me.eat("line")
			break
		}
		if token.is == "eof" || token.is == "#" {
			break
		}
		if token.is == "id" {
			typeName := token.value
			me.eat("id")
			if _, ok := typesMap[typeName]; ok {
				panic(me.fail() + "type name \"" + typeName + "\" already used")
			}
			unionList := make([]string, 0)
			unionGOrder := make([]string, 0)
			if me.token.is == "(" {
				isSimple = false
				me.eat("(")
				for {
					if me.token.is == ")" {
						break
					}
					if me.token.is == "delim" {
						me.eat("delim")
						continue
					}
					unionArgType := me.token.value
					me.eat("id")
					if _, ok := me.hmfile.types[unionArgType]; !ok {
						if _, ok2 := genericsDict[unionArgType]; ok2 {
							unionGOrder = append(unionGOrder, unionArgType)
						} else {
							panic(me.fail() + "union type name \"" + unionArgType + "\" does not exist")
						}
					}
					unionList = append(unionList, unionArgType)
				}
				me.eat(")")
			}
			me.eat("line")
			un := me.hmfile.unionInit(typeName, unionList, unionGOrder)
			typesOrder = append(typesOrder, un)
			typesMap[typeName] = un
			continue
		}
		panic(me.fail() + "bad token \"" + token.is + "\" in enum")
	}
	me.hmfile.enums[name] = enumInit(me.hmfile, name, isSimple, typesOrder, typesMap, genericsOrder, genericsDict)
}
