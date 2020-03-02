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

func (me *parser) ifexpr() (*node, *parseError) {
	var er *parseError
	depth := me.token.depth
	me.eat("if")
	n := nodeInit("if")
	b, er := me.calcBool()
	if er != nil {
		return nil, er
	}
	n.push(b)
	templs := me.getenumstack(n)
	if me.token.is == ":" {
		me.eat(":")
		block := nodeInit("block")
		e, er := me.expression()
		if er != nil {
			return nil, er
		}
		block.push(e)
		n.push(block)
	} else {
		me.eat("line")
		b, er = me.block()
		if er != nil {
			return nil, er
		}
		n.push(b)
	}
	if (me.peek().is == "elif" || me.peek().is == "else") && me.peek().depth == depth && me.token.is == "line" {
		me.eat("line")
	}
	me.enumstackclr(templs)
	for me.token.is == "elif" && me.token.depth == depth {
		me.eat("elif")
		elif := nodeInit("elif")
		b, er := me.calcBool()
		if er != nil {
			return nil, er
		}
		elif.push(b)
		templs := me.getenumstack(elif)
		if me.token.is == ":" {
			me.eat(":")
			block := nodeInit("block")
			e, er := me.expression()
			if er != nil {
				return nil, er
			}
			block.push(e)
			n.push(block)
		} else {
			me.eat("line")
			b, er := me.block()
			if er != nil {
				return nil, er
			}
			elif.push(b)
		}
		me.enumstackclr(templs)
		n.push(elif)
		if (me.peek().is == "elif" || me.peek().is == "else") && me.peek().depth == depth && me.token.is == "line" {
			me.eat("line")
		}
	}
	if me.token.is == "else" && me.token.depth == depth {
		me.eat("else")
		el := nodeInit("else")
		if me.token.is == ":" {
			me.eat(":")
			exp, er := me.expression()
			if er != nil {
				return nil, er
			}
			block := nodeInit("block")
			block.push(exp)
			el.push(block)
		} else {
			me.eat("line")
			b, er := me.block()
			if er != nil {
				return nil, er
			}
			el.push(b)
		}
		n.push(el)
		if me.token.is == "line" {
			me.eat("line")
		}
	}
	return n, nil
}
