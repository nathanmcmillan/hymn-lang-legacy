package main

import "fmt"

func (me *parser) prefix() *node {
	op := me.token.is
	fmt.Println("prefix", op)
	if pre, ok := prefixes[op]; ok {
		return pre.fn(me, op)
	}
	panic(me.fail() + "unknown calc prefix \"" + op + "\"")
}

func (me *parser) infix(left *node) *node {
	op := me.infixOp()
	fmt.Println("infix", op)
	if inf, ok := infixes[op]; ok {
		return inf.fn(me, left, op)
	}
	panic(me.fail() + "unknown calc infix \"" + op + "\"")
}

func (me *parser) calc(precedence int) *node {
	fmt.Println("calc", precedence)
	node := me.prefix()
	for {
		op := me.infixOp()
		fmt.Println("precedence :=", precedence)
		fmt.Println(op, "infix precedence :=", getInfixPrecedence(op))
		if precedence >= getInfixPrecedence(op) {
			break
		}
		node = me.infix(node)
	}
	fmt.Println("calc return", node.string(0))
	return node
}
