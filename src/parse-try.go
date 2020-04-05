package main

func (me *parser) tryAutoCatch(n *node, fnret *datatype) (*node, *parseError) {
	catch := nodeInit("auto-catch")
	catch.copyData(fnret)
	n.push(catch)
	return n, nil
}

func (me *parser) tryCatch(n *node, fnret *datatype, calcdata *datatype) (*node, *parseError) {

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

	maybecalc := calcdata.isSomeOrNone()
	var data *datatype
	if maybecalc {
		data = calcdata.member.copy()
	} else {
		var er *parseError
		encalc, _, _ := calcdata.isEnum()
		data, er = encalc.getuniondata(me.hmfile, encalc.types[len(encalc.types)-1].name)
		if er != nil {
			return nil, er
		}
	}

	me.hmfile.pushScope()

	tempd := me.hmfile.varInitWithData(data, temp, false)
	me.hmfile.scope.variables[temp] = tempd

	catch := nodeInit("catch")
	if maybecalc {
		catch.copyData(fnret.member)
	} else {
		catch.copyData(fnret)
	}
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
		if me.isNewLine() {
			me.newLine()
		}
	}

	n.push(catch)

	me.hmfile.popScope()

	return n, nil
}

func (me *parser) trying() (*node, *parseError) {
	if er := me.eat("try"); er != nil {
		return nil, er
	}

	fn := me.hmfile.getFuncScope().fn
	maybefunc := fn.returns.isSomeOrNone()
	var enfunc *enum
	if !maybefunc {
		var ok bool
		enfunc, _, ok = fn.returns.isEnum()
		if !ok || len(enfunc.types) < 2 {
			er := err(me, ECodeEnumParameter, "`try` requires the current function to return the maybe type, or an enum with at least two types")
			return nil, er
		}
	}

	n := nodeInit("try")
	calc, er := me.calc(0, fn.returns)
	if er != nil {
		return nil, er
	}

	maybecalc := calc.data().isSomeOrNone()
	var encalc *enum
	if maybecalc {
		if maybefunc {
			if !fn.returns.equals(calc.data()) {
				er := err(me, ECodeEnumParameter, "`try` calculated type and function returning type do not match")
				return nil, er
			}
		}
	} else {
		var ok bool
		encalc, _, ok = calc.data().isEnum()
		if !ok || len(encalc.types) < 2 {
			er := err(me, ECodeEnumParameter, "`try` requires the following expression to return the maybe type, or an enum with at least two types")
			return nil, er
		}
	}

	if maybecalc {
		n.copyData(calc.data())
	} else {
		if encalc.types[0].types.size() == 1 {
			n.copyData(encalc.types[0].types.get(0))
		} else {
			data, er := encalc.getuniondata(me.hmfile, encalc.types[0].name)
			if er != nil {
				return nil, er
			}
			n.setData(data)
		}
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

	return me.tryCatch(n, fn.returns, calc.data())
}
