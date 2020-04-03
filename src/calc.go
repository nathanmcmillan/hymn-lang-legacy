package main

func (me *parser) prefix() (*node, *parseError) {
	op := me.token.is
	if pre, ok := prefixes[op]; ok {
		return pre.fn(me, op)
	}
	return nil, err(me, ECodeUnknownCalcPrefix, "unknown calc prefix \""+op+"\"")
}

func (me *parser) infix(left *node) (*node, *parseError) {
	op, er := me.infixOp()
	if er != nil {
		return nil, er
	}
	if inf, ok := infixes[op]; ok {
		return inf.fn(me, left, op)
	}
	return nil, err(me, ECodeUnknownCalcInfix, "unknown calc infix \""+op+"\"")
}

func (me *parser) calc(precedence int, hint *datatype) (*node, *parseError) {
	me.hmfile.pushAssignStack(hint)
	node, er := me.prefix()
	if er != nil {
		return nil, er
	}
	for {
		op, er := me.infixOp()
		if er != nil {
			return nil, er
		}
		if precedence >= getInfixPrecedence(op) {
			break
		}
		node, er = me.infix(node)
		if er != nil {
			return nil, er
		}
	}
	me.hmfile.popAssignStack()
	return node, nil
}
