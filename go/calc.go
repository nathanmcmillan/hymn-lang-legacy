package main

func (me *parser) prefix() *node {
	op := me.token.is
	if pre, ok := prefixes[op]; ok {
		return pre.fn(me, op)
	}
	panic(me.fail() + "unknown calc prefix \"" + op + "\"")
}

func (me *parser) infix(left *node) *node {
	op := me.infixOp()
	if inf, ok := infixes[op]; ok {
		return inf.fn(me, left, op)
	}
	panic(me.fail() + "unknown calc infix \"" + op + "\"")
}

func (me *parser) calc(precedence int, hint *datatype) *node {
	me.hmfile.pushAssignStack(hint)
	node := me.prefix()
	for {
		op := me.infixOp()
		if precedence >= getInfixPrecedence(op) {
			break
		}
		node = me.infix(node)
	}
	me.hmfile.popAssignStack()
	return node
}
