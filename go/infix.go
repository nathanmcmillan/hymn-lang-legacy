package main

func (me *parser) infixConcat(left *node) *node {
	node := nodeInit("concat")
	node.typed = left.typed
	node.push(left)
	for me.token.is == "+" {
		me.eat("+")
		right := me.calc(getInfixPrecedence("+"))
		if right.typed != "string" {
			err := me.fail() + "concatenation operation must be strings \"" + left.typed + "\" and \"" + right.typed + "\""
			err += "\nleft: " + left.string(0) + "\nright: " + right.string(0)
			panic(err)
		}
		node.push(right)
	}
	return node
}

func infixBinary(me *parser, left *node, op string) *node {
	if op == "+" && left.typed == "string" {
		return me.infixConcat(left)
	}
	node := nodeInit(op)
	node.value = me.token.value
	me.eat(op)
	right := me.calc(getInfixPrecedence(op))
	if !isNumber(left.typed) || !isNumber(right.typed) {
		err := me.fail() + "binary operation must be numbers \"" + left.typed + "\" and \"" + right.typed + "\""
		err += "\nleft: " + left.string(0) + "\nright: " + right.string(0)
		panic(err)
	}
	if me.hmfile.typeToVarData(left.typed).notEqual(me.hmfile.typeToVarData(right.typed)) {
		err := me.fail() + "number types do not match \"" + left.typed + "\" and \"" + right.typed + "\""
		panic(err)
	}
	node.push(left)
	node.push(right)
	node.typed = left.typed
	return node
}

func infixBinaryInt(me *parser, left *node, op string) *node {
	node := nodeInit(op)
	node.value = me.token.value
	me.eat(op)
	right := me.calc(getInfixPrecedence(op))
	if left.typed != "int" || right.typed != "int" {
		err := me.fail() + "operation requires integers \"" + left.typed + "\" and \"" + right.typed + "\""
		err += "\nleft: " + left.string(0) + "\nright: " + right.string(0)
		panic(err)
	}
	node.push(left)
	node.push(right)
	node.typed = left.typed
	return node
}

func infixCompare(me *parser, left *node, op string) *node {
	node := nodeInit(getInfixName(op))
	node.value = me.token.value
	me.eat(op)
	right := me.calc(getInfixPrecedence(op))
	node.push(left)
	node.push(right)
	node.typed = "bool"
	return node
}

func infixWalrus(me *parser, left *node, op string) *node {
	node := me.assign(left, true, false)
	return node
}
