package main

func (me *cfile) compileIf(n *node) *codeblock {
	code := me.walrusIf(n)
	ifval := me.eval(n.has[0])
	code += "if (" + ifval.pop() + ") {\n"

	cblock := &codeblock{}
	cblock.prepend(ifval.pre)

	hsize := len(n.has)
	ix := 1
	for ix < hsize && n.has[ix].is == "variable" {
		temp := n.has[ix]
		tempname := temp.idata.name
		tempv := temp.data().getnamedvariable(tempname, false)
		me.scope.variables[tempname] = tempv
		ix++
	}
	thenEval := me.eval(n.has[ix])
	thenCode := thenEval.code()
	code += me.maybeFmc(thenCode, me.depth+1) + thenCode + me.maybeColon(thenCode)
	code += me.maybeNewLine(code) + fmc(me.depth) + "}"
	ix++
	for ix < hsize && n.has[ix].is == "elif" {
		elif := n.has[ix]
		elifval := me.eval(elif.has[0])
		cblock.prepend(elifval.pre)
		code += " else if (" + elifval.pop() + ") {\n"
		elsize := len(elif.has)
		ixo := 1
		for ixo < elsize && elif.has[ixo].is == "variable" {
			temp := elif.has[ixo]
			tempname := temp.idata.name
			tempv := temp.data().getnamedvariable(tempname, false)
			me.scope.variables[tempname] = tempv
			ixo++
		}
		thenBlock := me.eval(elif.has[ixo]).code()
		code += me.maybeFmc(thenBlock, me.depth+1) + thenBlock + me.maybeColon(thenBlock)
		code += me.maybeNewLine(code) + fmc(me.depth) + "}"
		ix++
	}
	if ix < hsize && n.has[ix].is == "else" {
		el := n.has[ix].has[0]
		code += " else {\n"
		c := me.eval(el).code()
		code += me.maybeFmc(c, me.depth+1) + c + me.maybeColon(c)
		code += me.maybeNewLine(code) + fmc(me.depth) + "}"
	}

	cblock.current = codeNode(n, code)
	return cblock
}
