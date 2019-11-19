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
	depth := me.token.depth
	me.eat("if")
	n := nodeInit("if")
	n.push(me.calcBool())
	templs := me.getenumstack(n)
	if me.token.is == ":" {
		me.eat(":")
		block := nodeInit("block")
		block.push(me.expression())
		n.push(block)
	} else {
		me.eat("line")
		n.push(me.block())
	}
	if me.peek().is == "elif" && me.peek().depth == depth && me.token.is == "line" {
		me.eat("line")
	}
	me.enumstackclr(templs)
	for me.token.is == "elif" && me.token.depth == depth {
		me.eat("elif")
		elif := nodeInit("elif")
		elif.push(me.calcBool())
		templs := me.getenumstack(elif)
		if me.token.is == ":" {
			me.eat(":")
			block := nodeInit("block")
			block.push(me.expression())
			n.push(block)
		} else {
			me.eat("line")
			elif.push(me.block())
		}
		me.enumstackclr(templs)
		n.push(elif)
		if (me.peek().is == "elif" || me.peek().is == "else") && me.token.is == "line" {
			me.eat("line")
		}
	}
	if me.token.is == "else" && me.token.depth == depth {
		me.eat("else")
		el := nodeInit("else")
		if me.token.is == ":" {
			me.eat(":")
			exp := me.expression()
			block := nodeInit("block")
			block.push(exp)
			el.push(block)
		} else {
			me.eat("line")
			el.push(me.block())
		}
		n.push(el)
		if me.token.is == "line" {
			me.eat("line")
		}
	}
	return n
}

func (me *parser) iswhile() bool {
	pos := me.pos
	token := me.tokens.get(pos)
	for token.is != "line" && token.is != "eof" {
		if token.is == "," {
			return false
		}
		pos++
		token = me.tokens.get(pos)
	}
	return true
}

func (me *parser) forexpr() *node {
	me.eat("for")
	var no *node
	var templs []*variableNode
	if me.token.is == "line" {
		me.eat("line")
		no = nodeInit("loop")
	} else {
		if me.iswhile() {
			no = nodeInit("while")
			no.push(me.calcBool())
			templs = me.getenumstack(no)
		} else {
			no = nodeInit("for")
			no.push(me.forceassign(true, true))
			me.eat(",")
			no.push(me.calcBool())
			me.eat(",")
			no.push(me.forceassign(true, true))
		}
		me.eat("line")
	}
	no.push(me.block())
	if templs != nil {
		me.enumstackclr(templs)
	}
	return no
}
