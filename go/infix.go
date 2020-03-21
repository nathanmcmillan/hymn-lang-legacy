package main

import "fmt"

func (me *parser) infixConcat(left *node) (*node, *parseError) {
	node := nodeInit("concat")
	node.copyDataOfNode(left)
	node.push(left)
	for me.token.is == "+" {
		if er := me.eat("+"); er != nil {
			return nil, er
		}
		right, er := me.calc(getInfixPrecedence("+"), nil)
		if er != nil {
			return nil, er
		}
		if !right.data().isString() {
			e := me.fail() + "concatenation operation must be strings but found \"" + left.data().print() + "\" and \"" + right.data().print() + "\""
			e += "\nleft: " + left.string(me.hmfile, 0) + "\nright: " + right.string(me.hmfile, 0)
			return nil, err(me, ECodeStringConcatenation, e)
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
	if er := me.eat(op); er != nil {
		return nil, er
	}
	right, er := me.calc(getInfixPrecedence(op), nil)
	if er != nil {
		return nil, er
	}
	if !left.data().isNumber() || !right.data().isNumber() {
		e := me.fail() + "operation expected numbers but was \"" + left.data().print() + "\" and \"" + right.data().print() + "\""
		e += "\n\nleft: " + left.string(me.hmfile, 0) + "\n\nright: " + right.string(me.hmfile, 0)
		return nil, err(me, ECodeOperationExpectedNumber, e)
	}
	if leftdata.notEquals(right.data()) {
		e := me.fail() + "number types do not match \"" + left.data().print() + "\" and \"" + right.data().print() + "\""
		return nil, err(me, ECodeNumberTypeMismatch, e)
	}
	node.push(left)
	node.push(right)
	node.copyDataOfNode(left)
	return node, nil
}

func infixBinaryInt(me *parser, left *node, op string) (*node, *parseError) {
	node := nodeInit(op)
	node.value = me.token.value
	if er := me.eat(op); er != nil {
		return nil, er
	}
	right, er := me.calc(getInfixPrecedence(op), nil)
	if er != nil {
		return nil, er
	}
	if !left.data().isAnyIntegerType() || !right.data().isAnyIntegerType() || left.data().notEquals(right.data()) {
		e := me.fail() + "operation requires discrete integers \"" + left.data().print() + "\" and \"" + right.data().print() + "\""
		e += "\nleft: " + left.string(me.hmfile, 0) + "\nright: " + right.string(me.hmfile, 0)
		return nil, err(me, ECodeOperationRequiresDiscreteNumber, e)
	}
	node.push(left)
	node.push(right)
	node.copyDataOfNode(left)
	return node, nil
}

func infixCompare(me *parser, left *node, op string) (*node, *parseError) {
	node := nodeInit(getInfixName(op))
	node.value = me.token.value
	if er := me.eat(op); er != nil {
		return nil, er
	}
	right, er := me.calc(getInfixPrecedence(op), nil)
	if er != nil {
		return nil, er
	}
	if left.data().notEquals(right.data()) {
		return nil, err(me, ECodeBadAssignment, fmt.Sprintf("Left %s and right `%s` types do not match.", left.data().print(), right.data().print()))
	}
	node.push(left)
	node.push(right)
	node._vdata = newdataprimitive("bool")
	return node, nil
}

func infixCompareEnumIs(me *parser, left *node, op string) (*node, *parseError) {
	n := nodeInit(getInfixName(op))
	return me.parseIs(left, op, n)
}

func infixTernary(me *parser, condition *node, op string) (*node, *parseError) {
	node := nodeInit(getInfixName(op))
	node.value = me.token.value
	if er := me.eat(op); er != nil {
		return nil, er
	}
	one, er := me.calc(getInfixPrecedence(op), nil)
	if er != nil {
		return nil, er
	}
	if one.data().isVoid() {
		return nil, err(me, ECodeTernaryVoid, "left type cannot be void")
	}
	node.push(condition)
	node.push(one)
	if er := me.eat(":"); er != nil {
		return nil, er
	}
	two, er := me.calc(getInfixPrecedence(op), nil)
	if er != nil {
		return nil, er
	}
	if one.data().notEquals(two.data()) {
		return nil, err(me, ECodeTernaryTypeMismatch, "left \""+one.data().print()+"\" and right \""+two.data().print()+"\" types do not match")
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
