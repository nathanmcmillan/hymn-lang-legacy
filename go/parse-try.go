package main

import "fmt"

func (me *parser) tryableType(data *datatype) {

}

func (me *parser) tryAutoCatch(n *node, fnret *datatype, enfunc, encalc *enum) (*node, *parseError) {

	if encalc.types[0].types.size() == 1 {
		n.copyData(encalc.types[0].types.get(0))
		fmt.Println("[DATA [1]")
	} else {
		data, er := encalc.getuniondata(me.hmfile, encalc.types[0].name)
		if er != nil {
			return nil, er
		}
		n.setData(data)
	}

	catch := nodeInit("auto-catch")
	catch.copyData(fnret)
	n.push(catch)

	// enCalcLastType := encalc.types[len(encalc.types)-1]

	// boo := nodeInit(getInfixName("=="))
	// boo.setData(newdataprimitive("bool"))
	// boo.push(left)
	// boo.push(right)

	// ret := nodeInit("return")
	// if enCalcLastType.types.size() == 1 {
	// 	ret.copyData(enCalcLastType.types.get(0))
	// } else {
	// 	data, er := encalc.getuniondata(me.hmfile, enCalcLastType.name)
	// 	if er != nil {
	// 		return nil, er
	// 	}
	// 	ret.setData(data)
	// }

	// catch := nodeInit("if")
	// catch.push(boo)
	// catch.push(ret)

	// n.push(catch)

	// ---------------------------

	// er := err(me, ECodeEnumParameter, "`try` without `catch` requires the current function and expression to return the same enum")
	// return nil, er

	fmt.Println("TRY NODE [1]::", n.string(me.hmfile, 0))
	return n, nil

	// fmt.Println("TRY NODE [2]::", n.string(me.hmfile, 0))
	// return n, nil
}

func (me *parser) tryCatch(fn *function, n *node, enfunc, encalc *enum) (*node, *parseError) {

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
	fmt.Println("[RETURNS]:", fn.returns.print())
	data, er := encalc.getuniondata(me.hmfile, encalc.types[len(encalc.types)-1].name)
	if er != nil {
		return nil, er
	}

	// TODO: need to check enum / union equality BUT only if no catch(e) is used, otherwise must be exact?

	fmt.Println("[RIGHT]:", data.print())
	tempd := me.hmfile.varInitWithData(data, temp, false)
	me.hmfile.scope.variables[temp] = tempd
	b, er := me.block()
	if er != nil {
		return nil, er
	}
	if temp != "" {
		delete(me.hmfile.scope.variables, temp)
	}
	n.push(b)
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
		return me.tryAutoCatch(n, fn.returns, enfunc, encalc)
	}

	return me.tryCatch(fn, n, enfunc, encalc)
}
