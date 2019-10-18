package main

func (me *parser) infixConcat(left *node) *node {
	node := nodeInit("concat")
	node.copyType(left)
	node.push(left)
	for me.token.is == "+" {
		me.eat("+")
		right := me.calc(getInfixPrecedence("+"))
		if right.getType() != TokenString {
			err := me.fail() + "concatenation operation must be strings \"" + left.getType() + "\" and \"" + right.getType() + "\""
			err += "\nleft: " + left.string(0) + "\nright: " + right.string(0)
			panic(err)
		}
		node.push(right)
	}
	return node
}

func infixBinary(me *parser, left *node, op string) *node {
	if op == "+" && left.getType() == TokenString {
		return me.infixConcat(left)
	}
	node := nodeInit(op)
	node.value = me.token.value
	me.eat(op)
	right := me.calc(getInfixPrecedence(op))
	if !isNumber(left.getType()) || !isNumber(right.getType()) {
		err := me.fail() + "binary operation must be numbers \"" + left.getType() + "\" and \"" + right.getType() + "\""
		err += "\n\nleft: " + left.string(0) + "\n\nright: " + right.string(0)
		panic(err)
	}
	if left.data().notEqual(right.data()) {
		err := me.fail() + "number types do not match \"" + left.getType() + "\" and \"" + right.getType() + "\""
		panic(err)
	}
	node.push(left)
	node.push(right)
	node.copyType(left)
	return node
}

func infixBinaryInt(me *parser, left *node, op string) *node {
	node := nodeInit(op)
	node.value = me.token.value
	me.eat(op)
	right := me.calc(getInfixPrecedence(op))
	if !isInteger(left.getType()) || !isInteger(right.getType()) {
		err := me.fail() + "operation requires discrete integers \"" + left.getType() + "\" and \"" + right.getType() + "\""
		err += "\nleft: " + left.string(0) + "\nright: " + right.string(0)
		panic(err)
	}
	if left.getType() != right.getType() {
		err := me.fail() + "operation types do not match \"" + left.getType() + "\" and \"" + right.getType() + "\""
		err += "\nleft: " + left.string(0) + "\nright: " + right.string(0)
		panic(err)
	}
	node.push(left)
	node.push(right)
	node.copyType(left)
	return node
}

func infixCompare(me *parser, left *node, op string) *node {
	node := nodeInit(getInfixName(op))
	node.value = me.token.value
	me.eat(op)
	right := me.calc(getInfixPrecedence(op))
	node.push(left)
	node.push(right)
	node.copyData(me.hmfile.typeToVarData("bool"))
	return node
}

func infixCompareEnumIs(me *parser, left *node, op string) *node {
	n := nodeInit(getInfixName(op))
	n.copyData(me.hmfile.typeToVarData("bool"))
	me.eat(op)
	var right *node
	if left.data().none || left.data().maybe {
		if me.token.is == "some" {
			right = nodeInit("some")
			me.eat("some")
		} else if me.token.is == "none" {
			right = nodeInit("none")
			me.eat("none")
		} else {
			panic(me.fail() + "right side of \"is\" was \"" + me.token.is + "\"")
		}
	} else {
		if _, _, ok := left.data().checkIsEnum(); !ok {
			panic(me.fail() + "left side of \"is\" must be enum but was \"" + left.data().full + "\"")
		}
		if me.token.is == "id" {
			name := me.token.value
			baseEnum, _, _ := left.data().checkIsEnum()
			if un, ok := baseEnum.types[name]; ok {
				me.eat("id")
				right = nodeInit("match-enum")
				right.copyData(me.hmfile.typeToVarData(baseEnum.name + "." + un.name))
			} else {
				right = me.calc(getInfixPrecedence(op))
			}
		} else {
			panic(me.fail() + "unknown right side of \"is\"")
		}
	}
	n.push(left)
	n.push(right)
	return n
}

func infixTernary(me *parser, condition *node, op string) *node {
	node := nodeInit(getInfixName(op))
	node.value = me.token.value
	me.eat(op)
	one := me.calc(getInfixPrecedence(op))
	if one.getType() == "void" {
		panic(me.fail() + "left type cannot be void")
	}
	node.push(condition)
	node.push(one)
	me.eat(":")
	two := me.calc(getInfixPrecedence(op))
	if one.getType() != two.getType() {
		panic(me.fail() + "left \"" + one.getType() + "\" and right \"" + two.getType() + "\" types do not match")
	}
	node.push(two)
	node.copyType(one)
	return node
}

func infixWalrus(me *parser, left *node, op string) *node {
	node := me.assign(left, true, false)
	return node
}
