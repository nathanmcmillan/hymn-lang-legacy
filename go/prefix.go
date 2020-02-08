package main

import (
	"strconv"
)

func prefixSign(me *parser, op string) *node {
	node := nodeInit(getPrefixName(op))
	me.eat(op)
	right := me.calc(getPrefixPrecedence(op), nil)
	node.push(right)
	node.copyDataOfNode(right)
	return node
}

func prefixGroup(me *parser, op string) *node {
	me.eat("(")
	node := me.calc(0, nil)
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
	node.copyData(getdatatype(me.hmfile, t))
	node.value = me.token.value
	me.eat(op)
	return node
}

func prefixString(me *parser, op string) *node {
	node := nodeInit(TokenString)
	node.copyData(getdatatype(me.hmfile, TokenString))
	node.value = me.token.value
	me.eat(TokenStringLiteral)
	return node
}

func prefixChar(me *parser, op string) *node {
	node := nodeInit(TokenChar)
	node.copyData(getdatatype(me.hmfile, TokenChar))
	node.value = me.token.value
	me.eat(TokenCharLiteral)
	return node
}

func prefixNot(me *parser, op string) *node {
	if me.token.is == "!" {
		me.eat("!")
	} else {
		me.eat("not")
	}
	node := nodeInit("not")
	node.copyData(getdatatype(me.hmfile, "bool"))
	node.push(me.calcBool())
	return node
}

func prefixCast(me *parser, op string) *node {
	me.eat(op)
	node := nodeInit("cast")
	node.copyData(getdatatype(me.hmfile, op))
	calc := me.calc(getPrefixPrecedence(op), nil)
	data := calc.data()
	if data.canCastToNumber() {
		node.push(calc)
		return node
	}
	panic(me.fail() + "invalid cast \"" + data.print() + "\"")
}

func prefixIdent(me *parser, op string) *node {
	useStack := false
	if me.token.is == "$" {
		me.eat("$")
		useStack = true
	}

	name := me.token.value
	module := me.hmfile

	v := module.getvar(name)
	if v != nil {
		if me.peek().is == ":=" {
			if v.mutable == false {
				panic(me.fail() + "Variable \"" + v.name + "\" is not mutable.")
			}
		}
		return me.eatvar(module)
	}

	if _, ok := me.hmfile.imports[name]; ok && me.peek().is == "." {
		return me.extern()
	}

	if _, ok := module.getType(name); ok {
		if _, ok := module.getFunction(name); ok {
			return me.parseFn(module)
		}
		if _, ok := module.getClass(name); ok {
			data := &allocData{}
			data.stack = useStack
			return me.allocClass(module, data)
		}
		if _, ok := module.enums[name]; ok {
			alloc := &allocData{}
			alloc.stack = useStack
			no := me.allocEnum(module)
			no._vdata = no._vdata.merge(alloc)
			return no
		}
		if def, ok := module.defs[name]; ok {
			return me.exprDef(name, def)
		}
		panic(me.fail() + "Bad type \"" + name + "\" definition.")
	}

	panic(me.fail() + "Unknown value \"" + name + "\".")
}

func prefixArray(me *parser, op string) *node {
	me.eat("[")
	alloc := &allocData{}
	var no *node
	var size *node
	simple := false
	if me.token.is == "]" {
		alloc.slice = true
		no = nodeInit("slice")
		simple = true
	} else if me.token.is == ":" {
		me.eat(":")
		alloc.slice = true
		no = nodeInit("slice")
		if me.token.is != "]" {
			capacity := me.calc(0, nil)
			if !capacity.data().isInt() {
				panic(me.fail() + "slice capacity " + capacity.string(me.hmfile, 0) + " is not an integer")
			}
			defaultSize := nodeInit(TokenInt)
			defaultSize.value = "0"
			defaultSize._vdata = getdatatype(me.hmfile, TokenInt)
			no.push(defaultSize)
			no.push(capacity)
		}
	} else {
		size = me.calc(0, nil)
		if !size.data().isInt() {
			panic(me.fail() + "array or slice size " + size.string(me.hmfile, 0) + " is not an integer")
		}
		slice := false
		var capacity *node
		if me.token.is == ":" {
			me.eat(":")
			slice = true
			if me.token.is != "]" {
				capacity = me.calc(0, nil)
				if !capacity.data().isInt() {
					panic(me.fail() + "slice capacity " + capacity.string(me.hmfile, 0) + " is not an integer")
				}
			}
		}
		if slice || size.is != TokenInt {
			alloc.slice = true
			no = nodeInit("slice")
		} else {
			alloc.array = true
			alloc.size, _ = strconv.Atoi(size.value)
			no = nodeInit("array")
		}
		no.push(size)
		if capacity != nil {
			no.push(capacity)
		}
	}
	me.eat("]")
	data := me.declareType(true)
	if me.token.is == "(" {
		items := nodeInit("items")
		me.eat("(")
		for {
			item := me.calc(0, data)
			if item.data().notEquals(data) {
				panic(me.fail() + "array member type \"" + item.data().print() + "\" does not match array type \"" + no.data().getmember().print() + "\"")
			}
			items.push(item)
			if me.token.is == ")" {
				break
			}
			me.eat(",")
		}
		me.eat(")")

		if size != nil {
			sizeint, er := strconv.Atoi(size.value)
			if er != nil || sizeint < len(items.has) {
				panic(me.fail() + "defined array size is less than implied size")
			}
		}
		no.push(items)

		if simple {
			no.is = "array"
			alloc.array = true
			alloc.slice = false
			alloc.size = len(items.has)
		}
	}
	data = data.merge(alloc)
	no._vdata = data

	return no
}

func prefixNone(me *parser, op string) *node {
	me.verify("none")
	n := nodeInit("none")
	n._vdata = me.declareType(true)
	return n
}

func prefixMaybe(me *parser, op string) *node {
	me.verify("maybe")
	n := nodeInit("maybe")
	n._vdata = me.declareType(true)
	return n
}
