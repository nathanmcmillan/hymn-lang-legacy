package main

func (me *parser) infixConcat(left *node) (*node, *parseError) {
	node := nodeInit("concat")
	node.copyDataOfNode(left)
	node.push(left)
	for me.token.is == "+" {
		me.eat("+")
		right, er := me.calc(getInfixPrecedence("+"), nil)
		if er != nil {
			return nil, er
		}
		if !right.data().isString() {
			err := me.fail() + "concatenation operation must be strings but found \"" + left.data().print() + "\" and \"" + right.data().print() + "\""
			err += "\nleft: " + left.string(me.hmfile, 0) + "\nright: " + right.string(me.hmfile, 0)
			panic(err)
		}
		node.push(right)
	}
	return node, nil
}

func infixBinary(me *parser, left *node, op string) (*node, *parseError) {
	leftdata := left.data()
	if op == "+" {
		if leftdata.isString() {
			return me.infixConcat(left)
		}
	}
	node := nodeInit(op)
	node.value = me.token.value
	me.eat(op)
	right, er := me.calc(getInfixPrecedence(op), nil)
	if er != nil {
		return nil, er
	}
	if !left.data().isNumber() || !right.data().isNumber() {
		err := me.fail() + "operation expected numbers but was \"" + left.data().print() + "\" and \"" + right.data().print() + "\""
		err += "\n\nleft: " + left.string(me.hmfile, 0) + "\n\nright: " + right.string(me.hmfile, 0)
		panic(err)
	}
	if leftdata.notEquals(right.data()) {
		err := me.fail() + "number types do not match \"" + left.data().print() + "\" and \"" + right.data().print() + "\""
		panic(err)
	}
	node.push(left)
	node.push(right)
	node.copyDataOfNode(left)
	return node, nil
}

func infixBinaryInt(me *parser, left *node, op string) (*node, *parseError) {
	node := nodeInit(op)
	node.value = me.token.value
	me.eat(op)
	right, er := me.calc(getInfixPrecedence(op), nil)
	if er != nil {
		return nil, er
	}
	if !left.data().isAnyIntegerType() || !right.data().isAnyIntegerType() || left.data().notEquals(right.data()) {
		err := me.fail() + "operation requires discrete integers \"" + left.data().print() + "\" and \"" + right.data().print() + "\""
		err += "\nleft: " + left.string(me.hmfile, 0) + "\nright: " + right.string(me.hmfile, 0)
		panic(err)
	}
	node.push(left)
	node.push(right)
	node.copyDataOfNode(left)
	return node, nil
}

func infixCompare(me *parser, left *node, op string) (*node, *parseError) {
	node := nodeInit(getInfixName(op))
	node.value = me.token.value
	me.eat(op)
	right, er := me.calc(getInfixPrecedence(op), nil)
	if er != nil {
		return nil, er
	}
	node.push(left)
	node.push(right)
	d, er := getdatatype(me.hmfile, "bool")
	if er != nil {
		return nil, er
	}
	node.copyData(d)
	return node, nil
}

func infixCompareEnumIs(me *parser, left *node, op string) (*node, *parseError) {
	n := nodeInit(getInfixName(op))
	return me.parseIs(left, op, n)
}

func infixTernary(me *parser, condition *node, op string) (*node, *parseError) {
	node := nodeInit(getInfixName(op))
	node.value = me.token.value
	me.eat(op)
	one, er := me.calc(getInfixPrecedence(op), nil)
	if er != nil {
		return nil, er
	}
	if one.data().isVoid() {
		return nil, err(me, "left type cannot be void")
	}
	node.push(condition)
	node.push(one)
	me.eat(":")
	two, er := me.calc(getInfixPrecedence(op), nil)
	if er != nil {
		return nil, er
	}
	if one.data().notEquals(two.data()) {
		return nil, err(me, "left \""+one.data().print()+"\" and right \""+two.data().print()+"\" types do not match")
	}
	node.push(two)
	node.copyDataOfNode(one)
	return node, nil
}

func infixWalrus(me *parser, left *node, op string) (*node, *parseError) {
	node, er := me.assign(left, true, false)
	if er != nil {
		return nil, er
	}
	return node, nil
}
