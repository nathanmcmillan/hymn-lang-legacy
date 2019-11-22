package main

func (me *parser) forloop() *node {
	me.eat("for")
	var no *node
	var templs []*variableNode
	no = nodeInit("for")
	no.push(me.forceassign(true, true))
	me.eat(",")
	no.push(me.calcBool())
	me.eat(",")
	no.push(me.forceassign(true, true))
	me.eat("line")
	no.push(me.block())
	if templs != nil {
		me.enumstackclr(templs)
	}
	return no
}

func (me *parser) whileloop() *node {
	me.eat("while")
	var no *node
	var templs []*variableNode
	if me.token.is == "line" {
		me.eat("line")
		no = nodeInit("loop")
	} else {
		no = nodeInit("while")
		no.push(me.calcBool())
		templs = me.getenumstack(no)
		me.eat("line")
	}
	no.push(me.block())
	if templs != nil {
		me.enumstackclr(templs)
	}
	return no
}

func (me *parser) iterloop() *node {
	me.eat("iterate")
	var1 := me.token.value
	var2 := ""
	me.eat("id")
	if me.token.is == "," {
		me.eat(",")
		var2 = me.token.value
		me.eat("id")
	}
	me.eat("in")
	using := me.calc(0)
	if !using.data().checkIsArrayOrSlice() {
		panic(me.fail() + "expected array or slice but was \"" + using.data().full + "\"")
	}
	me.eat("line")

	no := nodeInit("iterate")

	d := nodeInit("variable")
	d.idata = &idData{}
	d.idata.module = me.hmfile
	d.idata.name = var1

	if var2 != "" {
		iterid := me.hmfile.varInit("int", var1, false)
		me.hmfile.scope.variables[iterid.name] = iterid
		e := nodeInit("variable")
		e.idata = &idData{}
		e.idata.module = me.hmfile
		e.idata.name = iterid.name
		e.copyData(iterid.data())
		no.push(e)

		d.idata.name = var2
	}

	itermint := me.hmfile.varInitFromData(using.data().memberType, d.idata.name, false)
	me.hmfile.scope.variables[itermint.name] = itermint
	d.copyData(itermint.data())

	block := me.block()

	no.push(d)
	no.push(using)
	no.push(block)
	return no
}