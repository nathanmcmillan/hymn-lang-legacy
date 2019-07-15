package main

func (me *parser) allocClass(module *hmfile) *node {
	n := nodeInit("new")
	// TODO deprecated
	n.typed = me.buildClass(module)
	// n.typed = me.declareType(true)
	return n
}

func (me *parser) allocEnum(module *hmfile) *node {
	enumName := me.token.value
	me.eat("id")
	enumOf, ok := module.enums[enumName]
	if !ok {
		panic(me.fail() + "enum \"" + enumName + "\" does not exist")
	}

	me.eat(".")
	typeName := me.token.value
	me.eat("id")
	enumType, ok := enumOf.types[typeName]
	if !ok {
		panic(me.fail() + "enum \"" + enumName + "\" does not have type \"" + typeName + "\"")
	}

	n := nodeInit("enum")

	typeSize := len(enumType.types)
	if typeSize > 0 {
		me.eat("(")
		for ix, unionType := range enumType.types {
			if ix != 0 {
				me.eat("delim")
			}
			param := me.calc()
			if param.typed != unionType {
				panic(me.fail() + "enum \"" + enumName + "\" type \"" + typeName + "\" expects \"" + unionType + "\" but parameter was \"" + param.typed + "\"")
			}
			n.push(param)
		}
		me.eat(")")
	}

	if me.hmfile == module {
		n.typed = enumName
		n.value = typeName
	} else {
		n.typed = module.name + "." + enumName
		n.value = typeName
	}
	return n
}
