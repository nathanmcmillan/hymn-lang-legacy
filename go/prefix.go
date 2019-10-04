package main

func prefixSign(me *parser, op string) *node {
	node := nodeInit(getPrefixName(op))
	me.eat(op)
	right := me.calc(getPrefixPrecedence(op))
	node.push(right)
	node.copyType(right)
	return node
}

func prefixGroup(me *parser, op string) *node {
	me.eat("(")
	node := me.calc(0)
	node.attributes["parenthesis"] = "true"
	me.eat(")")
	return node
}

func prefixPrimitive(me *parser, op string) *node {
	t, ok := literals[op]
	if !ok {
		panic(me.fail() + "unknown primitive \"" + op + "\"")
	}
	node := nodeInit(t)
	node.vdata = me.hmfile.typeToVarData(t)
	node.value = me.token.value
	me.eat(op)
	return node
}

func prefixString(me *parser, op string) *node {
	node := nodeInit(TokenString)
	node.vdata = me.hmfile.typeToVarData(TokenString)
	node.value = me.token.value
	me.eat(TokenStringLiteral)
	return node
}

func prefixChar(me *parser, op string) *node {
	node := nodeInit(TokenChar)
	node.vdata = me.hmfile.typeToVarData(TokenChar)
	node.value = me.token.value
	me.eat(TokenCharLiteral)
	return node
}

func prefixIdent(me *parser, op string) *node {
	useStack := false
	if me.token.is == "$" {
		me.eat("$")
		useStack = true
	}

	name := me.token.value
	module := me.hmfile

	if _, ok := module.getType(name); ok {
		if _, ok := module.getFunction(name); ok {
			return me.parseFn(module)
		}
		if _, ok := module.getClass(name); ok {
			data := &allocData{}
			data.useStack = useStack
			return me.allocClass(module, data)
		}
		if _, ok := module.enums[name]; ok {
			data := &allocData{}
			data.useStack = useStack
			return me.allocEnum(module, data)
		}
		if def, ok := module.defs[name]; ok {
			return me.exprDef(name, def)
		}
		panic(me.fail() + "bad type \"" + name + "\" definition")
	}
	if _, ok := module.imports[name]; ok {
		return me.extern()
	}
	v := module.getvar(name)
	if me.peek().is == ":=" {
		if v != nil && v.mutable == false {
			panic(me.fail() + "variable not mutable")
		}
	} else if v == nil {
		panic(me.fail() + "variable out of scope")
	}
	return me.eatvar(module)
}

func prefixArray(me *parser, op string) *node {
	me.eat("[")
	alloc := &allocData{}
	alloc.isArray = true
	var node *node
	if me.token.is == "]" {
		node = nodeInit("slice")
	} else {
		node = nodeInit("array")
		size := me.calc(0)
		if size.getType() != TokenInt {
			panic(me.fail() + "array size must be integer")
		}
		node.push(size)
	}
	me.eat("]")
	node.vdata = me.buildAnyType(alloc)
	return node
}

func prefixNot(me *parser, op string) *node {
	if me.token.is == "!" {
		me.eat("!")
	} else {
		me.eat("not")
	}
	node := nodeInit("not")
	node.vdata = me.hmfile.typeToVarData("bool")
	node.push(me.calcBool())
	return node
}

func prefixNone(me *parser, op string) *node {
	me.eat("none")
	node := nodeInit("none")
	if me.token.is == "<" {
		me.eat("<")
		option := me.declareType(true)
		me.eat(">")
		typed := "none<" + option.typed + ">"
		node.vdata = me.hmfile.typeToVarData(typed)
	} else {
		node.vdata = me.hmfile.typeToVarData("none")
	}
	return node
}

func prefixMaybe(me *parser, op string) *node {
	me.eat("maybe")
	me.eat("<")
	option := me.declareType(true)
	me.eat(">")
	typed := "maybe<" + option.typed + ">"
	data := me.hmfile.typeToVarData(typed)

	n := nodeInit("maybe")
	n.vdata = data
	return n
}

func prefixCast(me *parser, op string) *node {
	me.eat(op)
	node := nodeInit("cast")
	node.vdata = me.hmfile.typeToVarData(op)
	calc := me.calc(getPrefixPrecedence(op))
	value := calc.vdata.full
	if canCastToNumber(value) {
		node.push(calc)
		return node
	}
	panic(me.fail() + "invalid cast \"" + value + "\"")
}
