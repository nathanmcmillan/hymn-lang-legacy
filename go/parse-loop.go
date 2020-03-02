package main

func (me *parser) forloop() (*node, *parseError) {
	me.eat("for")
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
	me.eat(";")
	b, er := me.calcBool()
	if er != nil {
		return nil, er
	}
	no.push(b)
	me.eat(";")
	v, er = me.eatvar(me.hmfile)
	if er != nil {
		return nil, er
	}
	a, er := me.forceassign(v, true, true)
	if er != nil {
		return nil, er
	}
	no.push(a)
	me.eat("line")
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
	me.eat("while")
	var no *node
	var templs []*variableNode
	if me.token.is == "line" {
		me.eat("line")
		no = nodeInit("loop")
	} else {
		no = nodeInit("while")
		b, er := me.calcBool()
		if er != nil {
			return nil, er
		}
		no.push(b)
		templs = me.getenumstack(no)
		me.eat("line")
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
	me.eat("iterate")
	var1 := me.token.value
	var2 := ""
	me.eat("id")
	if me.token.is == "," {
		me.eat(",")
		if me.token.is == "_" {
			var2 = me.token.is
			me.eat("_")
		} else {
			var2 = me.token.value
			me.eat("id")
		}
	}
	me.eat("in")
	using, er := me.calc(0, nil)
	if er != nil {
		return nil, er
	}
	if !using.data().isArrayOrSlice() && !using.data().isString() {
		return nil, err(me, "expected array, slice, string but was \""+using.data().print()+"\"")
	}
	me.eat("line")

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
	me.hmfile.scope.variables[itermint.name] = itermint
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
