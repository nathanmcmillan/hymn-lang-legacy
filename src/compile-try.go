package main

func (me *cfile) compileTry(n *node) *codeblock {

	calc := n.has[0]
	catch := n.has[1]

	maybecalc := calc.data().isSomeOrNone()
	var enCalc *enum
	var enCalcFirstType *union
	var enCalcLastType *union
	if !maybecalc {
		enCalc, _, _ = calc.data().isEnum()
		enCalcFirstType = enCalc.types[0]
		enCalcLastType = enCalc.types[len(enCalc.types)-1]
	}

	cb := &codeblock{}

	temp := "try_" + me.temp()

	d := nodeInit("variable")
	d.idata = newidvariable(me.hmfile, temp)
	d.copyData(calc.data())

	decl := me.compileDeclare(d) + " = " + me.eval(calc).code()

	var enCatch *enum
	var enCatchLastType *union
	var enCatchLast string

	code := "if (" + temp

	if maybecalc {
		code += " != NULL"
	} else {
		enCatch, _, _ = catch.data().isEnum()
		enCatchLastType = enCatch.types[len(enCatch.types)-1]
		enCatchLast = enumTypeName(enCatch.baseEnum().cname, enCatchLastType.name)
		code += "->type == " + enCatchLast
	}

	code += ") {\n"

	if catch.is == "auto-catch" {
		if maybecalc || enCalc.name == enCatch.name {
			code += fmc(me.depth+1) + "return " + temp + ";\n"
		} else {
			wrapper := "catch_" + me.temp()
			code += fmc(me.depth+1) + catch.data().typeSig(me) + wrapper + " = malloc(sizeof(" + enCatch.ucname + "));\n"
			code += fmc(me.depth+1) + wrapper + "->type = " + enCatchLast + ";\n"
			code += fmc(me.depth+1) + wrapper + "->" + enCatchLastType.name + " = " + temp + "->" + enCalcLastType.name + ";\n"
			code += fmc(me.depth+1) + "return " + wrapper + ";\n"
		}

	} else {
		id := catch.idata.name
		me.scope.renaming[id] = temp
		code += me.eval(catch.has[0]).code()
		delete(me.scope.renaming, id)
	}

	code += fmc(me.depth) + "}"

	cb.prepend(codeBlockOne(n, code))
	cb.prepend(codeBlockOne(n, decl))

	if maybecalc {
	} else {
		assign := temp
		assign += "->" + enCalcFirstType.name
		cb.current = codeNode(n, assign)
	}

	return cb
}
