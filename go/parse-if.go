package main

func (me *parser) enumstackclr(stack []*variableNode) {
	if stack != nil {
		for _, tempd := range stack {
			delete(me.hmfile.scope.variables, tempd.v.name)
		}
	}
}

func (me *parser) getenumstack(n *node) []*variableNode {
	if len(me.hmfile.enumIsStack) > 0 {
		stack := me.hmfile.enumIsStack
		for _, temp := range stack {
			me.hmfile.scope.variables[temp.v.name] = temp.v
			n.push(temp.n)
		}
		me.hmfile.enumIsStack = make([]*variableNode, 0)
		return stack
	}
	return nil
}

func (me *parser) ifexpr() *node {
	me.eat("if")
	n := nodeInit("if")
	n.push(me.calcBool())
	templs := me.getenumstack(n)
	if me.token.is == ":" {
		me.eat(":")
		block := nodeInit("block")
		block.push(me.expression())
		n.push(block)
		me.eat("line")
	} else {
		me.eat("line")
		n.push(me.block())
	}
	me.enumstackclr(templs)
	for me.token.is == "elif" {
		me.eat("elif")
		other := nodeInit("elif")
		other.push(me.calcBool())
		templs := me.getenumstack(other)
		if me.token.is == ":" {
			me.eat(":")
			block := nodeInit("block")
			block.push(me.expression())
			n.push(block)
			me.eat("line")
		} else {
			me.eat("line")
			other.push(me.block())
		}
		me.enumstackclr(templs)
		n.push(other)
	}
	if me.token.is == "else" {
		me.eat("else")
		if me.token.is == ":" {
			me.eat(":")
			exp := me.expression()
			block := nodeInit("block")
			block.push(exp)
			n.push(block)
			me.eat("line")
		} else {
			me.eat("line")
			n.push(me.block())
		}
	}
	return n
}
