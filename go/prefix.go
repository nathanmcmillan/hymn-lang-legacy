package main

func prefixSign(me *parser, op string) *node {
	node := nodeInit(getPrefixName(op))
	me.eat(op)
	right := me.calc(getPrefixPrecedence(op))
	node.push(right)
	node.typed = right.typed
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
	node := nodeInit(op)
	node.typed = op
	node.value = me.token.value
	me.eat(op)
	return node
}

func prefixIdent(me *parser, op string) *node {
	name := me.token.value
	if _, ok := me.hmfile.functions[name]; ok {
		return me.call(me.hmfile)
	}
	if _, ok := me.hmfile.types[name]; ok {
		if _, ok := me.hmfile.classes[name]; ok {
			return me.allocClass(me.hmfile)
		}
		if _, ok := me.hmfile.enums[name]; ok {
			return me.allocEnum(me.hmfile)
		}
		panic(me.fail() + "bad type \"" + name + "\" definition")
	}
	if _, ok := me.hmfile.imports[name]; ok {
		return me.extern()
	}
	if me.hmfile.getvar(name) == nil {
		panic(me.fail() + "variable out of scope")
	}
	return me.eatvar(me.hmfile)
}

func prefixArray(me *parser, op string) *node {
	me.eat("[")
	size := me.calc(0)
	if size.typed != "int" {
		panic(me.fail() + "array size must be integer")
	}
	me.eat("]")
	node := nodeInit("array")
	node.typed = "[]" + me.buildAnyType()
	node.push(size)
	return node
}

func prefixNot(me *parser, op string) *node {
	if me.token.is == "!" {
		me.eat("!")
	} else {
		me.eat("not")
	}
	node := nodeInit("not")
	node.typed = "bool"
	node.push(me.calcBool())
	return node
}

func prefixNone(me *parser, op string) *node {
	me.eat("none")
	me.eat("<")
	option := me.declareType(true).typed
	me.eat(">")
	typed := "none<" + option + ">"
	me.defineMaybeImpl(typed)

	node := nodeInit("none")
	node.typed = typed
	return node
}

func prefixMaybe(me *parser, op string) *node {
	me.eat("maybe")
	me.eat("<")
	option := me.declareType(true).typed
	me.eat(">")
	typed := "maybe<" + option + ">"
	me.defineMaybeImpl(typed)

	n := nodeInit("maybe")
	n.typed = typed
	return n
}
