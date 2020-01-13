package main

func (me *parser) infixConcat(left *node) *node {
	node := nodeInit("concat")
	node.copyDataOfNode(left)
	node.push(left)
	for me.token.is == "+" {
		me.eat("+")
		right := me.calc(getInfixPrecedence("+"))
		if !right.data().checkIsString() {
			err := me.fail() + "concatenation operation must be strings but found \"" + left.data().print() + "\" and \"" + right.data().print() + "\""
			err += "\nleft: " + left.string(0) + "\nright: " + right.string(0)
			panic(err)
		}
		node.push(right)
	}
	return node
}

func infixBinary(me *parser, left *node, op string) *node {
	leftdata := left.data()
	if op == "+" {
		if leftdata.checkIsString() {
			return me.infixConcat(left)
		}
	}
	node := nodeInit(op)
	node.value = me.token.value
	me.eat(op)
	right := me.calc(getInfixPrecedence(op))
	if !left.data().isNumber() || !right.data().isNumber() {
		err := me.fail() + "operation expected numbers but was \"" + left.data().print() + "\" and \"" + right.data().print() + "\""
		err += "\n\nleft: " + left.string(0) + "\n\nright: " + right.string(0)
		panic(err)
	}
	if leftdata.notEqual(right.data()) {
		err := me.fail() + "number types do not match \"" + left.data().print() + "\" and \"" + right.data().print() + "\""
		panic(err)
	}
	node.push(left)
	node.push(right)
	node.copyDataOfNode(left)
	return node
}

func infixBinaryInt(me *parser, left *node, op string) *node {
	node := nodeInit(op)
	node.value = me.token.value
	me.eat(op)
	right := me.calc(getInfixPrecedence(op))
	if !left.data().isAnyIntegerType() || !right.data().isAnyIntegerType() || left.data().notEqual(right.data()) {
		err := me.fail() + "operation requires discrete integers \"" + left.data().print() + "\" and \"" + right.data().print() + "\""
		err += "\nleft: " + left.string(0) + "\nright: " + right.string(0)
		panic(err)
	}
	node.push(left)
	node.push(right)
	node.copyDataOfNode(left)
	return node
}

func infixCompare(me *parser, left *node, op string) *node {
	node := nodeInit(getInfixName(op))
	node.value = me.token.value
	me.eat(op)
	right := me.calc(getInfixPrecedence(op))
	node.push(left)
	node.push(right)
	node.copyData(typeToVarData(me.hmfile, "bool"))
	return node
}

func infixCompareEnumIs(me *parser, left *node, op string) *node {
	n := nodeInit(getInfixName(op))
	return me.parseIs(left, op, n)
}

func infixTernary(me *parser, condition *node, op string) *node {
	node := nodeInit(getInfixName(op))
	node.value = me.token.value
	me.eat(op)
	one := me.calc(getInfixPrecedence(op))
	if one.data().isVoid() {
		panic(me.fail() + "left type cannot be void")
	}
	node.push(condition)
	node.push(one)
	me.eat(":")
	two := me.calc(getInfixPrecedence(op))
	if one.data().notEqual(two.data()) {
		panic(me.fail() + "left \"" + one.data().print() + "\" and right \"" + two.data().print() + "\" types do not match")
	}
	node.push(two)
	node.copyDataOfNode(one)
	return node
}

func infixWalrus(me *parser, left *node, op string) *node {
	node := me.assign(left, true, false)
	return node
}
