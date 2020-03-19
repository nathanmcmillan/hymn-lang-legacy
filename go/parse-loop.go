package main

func (me *parser) forloop() (*node, *parseError) {
	if er := me.eat("for"); er != nil {
		return nil, er
	}
	var no *node
	var templs []*variableNode
	no = nodeInit("for")
	v, er := me.eatvar(me.hmfile)
	if er != nil {
		return nil, er
	}
	f, er := me.forceassign(v, true, true)
	if er != nil {
		return nil, er
	}
	no.push(f)
	if er := me.eat(";"); er != nil {
		return nil, er
	}
	b, er := me.calcBool()
	if er != nil {
		return nil, er
	}
	no.push(b)
	if er := me.eat(";"); er != nil {
		return nil, er
	}
	v, er = me.eatvar(me.hmfile)
	if er != nil {
		return nil, er
	}
	a, er := me.forceassign(v, true, true)
	if er != nil {
		return nil, er
	}
	no.push(a)
	if er := me.eat("line"); er != nil {
		return nil, er
	}
	b, er = me.block()
	if er != nil {
		return nil, er
	}
	no.push(b)
	if templs != nil {
		me.enumstackclr(templs)
	}
	return no, nil
}

func (me *parser) whileloop() (*node, *parseError) {
	if er := me.eat("while"); er != nil {
		return nil, er
	}
	var no *node
	var templs []*variableNode
	if me.token.is == "line" {
		if er := me.eat("line"); er != nil {
			return nil, er
		}
		no = nodeInit("loop")
	} else {
		no = nodeInit("while")
		b, er := me.calcBool()
		if er != nil {
			return nil, er
		}
		no.push(b)
		templs = me.getenumstack(no)
		if er := me.eat("line"); er != nil {
			return nil, er
		}
	}
	b, er := me.block()
	if er != nil {
		return nil, er
	}
	no.push(b)
	if templs != nil {
		me.enumstackclr(templs)
	}
	return no, nil
}

func (me *parser) iterloop() (*node, *parseError) {
	if er := me.eat("iterate"); er != nil {
		return nil, er
	}
	var1 := me.token.value
	var2 := ""
	if er := me.eat("id"); er != nil {
		return nil, er
	}
	if me.token.is == "," {
		me.next()
		if me.token.is == "_" {
			var2 = "_"
			me.next()
		} else {
			var2 = me.token.value
			if er := me.eat("id"); er != nil {
				return nil, er
			}
		}
	}
	if er := me.eat("in"); er != nil {
		return nil, er
	}
	using, er := me.calc(0, nil)
	if er != nil {
		return nil, er
	}
	if !using.data().isArrayOrSlice() && !using.data().isString() {
		return nil, err(me, ECodeVariableNotIndexable, "expected array, slice, string but was \""+using.data().print()+"\"")
	}
	if er := me.eat("line"); er != nil {
		return nil, er
	}

	no := nodeInit("iterate")

	d := nodeInit("variable")
	d.idata = newidvariable(me.hmfile, var1)

	if var2 != "" {
		iterid, er := me.hmfile.varInit("int", var1, false)
		if er != nil {
			return nil, er
		}
		me.hmfile.scope.variables[iterid.name] = iterid
		e := nodeInit("variable")
		e.idata = newidvariable(me.hmfile, iterid.name)
		e.copyData(iterid.data())
		no.push(e)

		d.idata.name = var2
	}

	itermint := using.data().getmember().getnamedvariable(d.idata.name, false)
	if var2 != "_" {
		me.hmfile.scope.variables[itermint.name] = itermint
	}
	d.copyData(itermint.data())

	block, er := me.block()
	if er != nil {
		return nil, er
	}

	no.push(d)
	no.push(using)
	no.push(block)
	return no, nil
}
