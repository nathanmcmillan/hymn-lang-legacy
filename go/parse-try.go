package main

func (me *parser) tryAutoCatch(n *node, fnret *datatype) (*node, *parseError) {
	catch := nodeInit("auto-catch")
	catch.copyData(fnret)
	n.push(catch)
	return n, nil
}

func (me *parser) tryCatch(n *node, fnret *datatype, encalc *enum) (*node, *parseError) {

	if er := me.eat("("); er != nil {
		return nil, er
	}
	temp := me.token.value
	if er := me.eat("id"); er != nil {
		return nil, er
	}
	if er := me.eat(")"); er != nil {
		return nil, er

	}
	data, er := encalc.getuniondata(me.hmfile, encalc.types[len(encalc.types)-1].name)
	if er != nil {
		return nil, er
	}
	tempd := me.hmfile.varInitWithData(data, temp, false)
	me.hmfile.scope.variables[temp] = tempd

	catch := nodeInit("catch")
	catch.copyData(fnret)
	catch.idata = newidvariable(me.hmfile, tempd.name)

	if me.isNewLine() {
		me.newLine()
		b, er := me.block()
		if er != nil {
			return nil, er
		}
		catch.push(b)
	} else {
		e, er := me.expression()
		if er != nil {
			return nil, er
		}
		block := nodeInit("block")
		block.push(e)
		catch.push(block)
	}
	if me.isNewLine() {
		me.newLine()
	}

	n.push(catch)

	if temp != "" {
		delete(me.hmfile.scope.variables, temp)
	}

	return n, nil
}

func (me *parser) trying() (*node, *parseError) {
	if er := me.eat("try"); er != nil {
		return nil, er
	}

	fn := me.hmfile.scope.fn
	enfunc, _, ok := fn.returns.isEnum()
	if !ok || len(enfunc.types) < 2 {
		er := err(me, ECodeEnumParameter, "`try` requires the current function to return an enum with at least two types")
		return nil, er
	}

	n := nodeInit("try")
	calc, er := me.calc(0, fn.returns)
	if er != nil {
		return nil, er
	}

	encalc, _, ok := calc.data().isEnum()
	if !ok || len(encalc.types) < 2 {
		er := err(me, ECodeEnumParameter, "`try` requires the following expression to return an enum with at least two types")
		return nil, er
	}

	if encalc.types[0].types.size() == 1 {
		n.copyData(encalc.types[0].types.get(0))
	} else {
		data, er := encalc.getuniondata(me.hmfile, encalc.types[0].name)
		if er != nil {
			return nil, er
		}
		n.setData(data)
	}

	n.push(calc)

	if me.token.is == "comment" && me.peek().is == "line" && me.doublePeek().is == "catch" {
		me.next()
		me.next()
		me.next()
	} else if me.token.is == "line" && me.peek().is == "catch" {
		me.next()
		me.next()
	} else if me.token.is == "catch" {
		me.next()
	} else {
		return me.tryAutoCatch(n, fn.returns)
	}

	return me.tryCatch(n, fn.returns, encalc)
}
