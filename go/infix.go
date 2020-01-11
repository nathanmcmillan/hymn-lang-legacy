package main

func (me *parser) infixConcat(left *node) *node {
	node := nodeInit("concat")
	node.copyDataOfNode(left)
	node.push(left)
	for me.token.is == "+" {
		me.eat("+")
		right := me.calc(getInfixPrecedence("+"))
		if !right.data().checkIsString() {
			err := me.fail() + "concatenation operation must be strings but found \"" + left.data().full + "\" and \"" + right.data().full + "\""
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
	if !isNumber(left.data().full) || !isNumber(right.data().full) {
		err := me.fail() + "operation expected numbers but was \"" + left.data().full + "\" and \"" + right.data().full + "\""
		err += "\n\nleft: " + left.string(0) + "\n\nright: " + right.string(0)
		panic(err)
	}
	if leftdata.notEqual(right.data()) {
		err := me.fail() + "number types do not match \"" + left.data().full + "\" and \"" + right.data().full + "\""
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
	if !isInteger(left.data().full) || !isInteger(right.data().full) {
		err := me.fail() + "operation requires discrete integers \"" + left.data().full + "\" and \"" + right.data().full + "\""
		err += "\nleft: " + left.string(0) + "\nright: " + right.string(0)
		panic(err)
	}
	if left.data().full != right.data().full {
		err := me.fail() + "operation types do not match \"" + left.data().full + "\" and \"" + right.data().full + "\""
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
	if one.data().full == "void" {
		panic(me.fail() + "left type cannot be void")
	}
	node.push(condition)
	node.push(one)
	me.eat(":")
	two := me.calc(getInfixPrecedence(op))
	if one.data().full != two.data().full {
		panic(me.fail() + "left \"" + one.data().full + "\" and right \"" + two.data().full + "\" types do not match")
	}
	node.push(two)
	node.copyDataOfNode(one)
	return node
}

func infixWalrus(me *parser, left *node, op string) *node {
	node := me.assign(left, true, false)
	return node
}
