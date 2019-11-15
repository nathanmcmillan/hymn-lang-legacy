package main

func (me *cfile) compileIf(n *node) *codeblock {
	hsize := len(n.has)
	code := ""
	code += me.walrusIf(n)
	ifval := me.eval(n.has[0])
	code += ifval.precode()
	code += "if (" + ifval.pop() + ") {\n"
	ix := 1
	for ix < hsize && n.has[ix].is == "variable" {
		temp := n.has[ix]
		tempname := temp.idata.name
		tempv := me.hmfile.varInitFromData(temp.data(), tempname, false)
		me.scope.variables[tempname] = tempv
		ix++
	}
	thenCode := me.eval(n.has[ix]).code()
	code += me.maybeFmc(thenCode, me.depth+1) + thenCode + me.maybeColon(thenCode)
	code += "\n" + fmc(me.depth) + "}"
	ix++
	for ix < hsize && n.has[ix].is == "elif" {
		elif := n.has[ix]
		elifval := me.eval(elif.has[0])
		code += elifval.precode()
		code += " else if (" + elifval.pop() + ") {\n"
		elsize := len(elif.has)
		ixo := 1
		for ixo < elsize && elif.has[ixo].is == "variable" {
			temp := elif.has[ixo]
			tempname := temp.idata.name
			tempv := me.hmfile.varInitFromData(temp.data(), tempname, false)
			me.scope.variables[tempname] = tempv
			ixo++
		}
		thenBlock := me.eval(elif.has[ixo]).code()
		code += me.maybeFmc(thenBlock, me.depth+1) + thenBlock + me.maybeColon(thenBlock)
		code += "\n" + fmc(me.depth) + "}"
		ix++
	}
	if ix >= 2 && ix < hsize {
		code += " else {\n"
		c := me.eval(n.has[ix]).code()
		code += me.maybeFmc(c, me.depth+1) + c + me.maybeColon(c)
		code += "\n" + fmc(me.depth) + "}"
	}
	return codeBlockOne(n, code)
}

func (me *cfile) compileLoop(op string, n *node) *codeblock {
	ix := 0
	code := ""
	if op == "loop" {
		code += "while (true) {\n"
	} else if op == "while" {
		ix++
		code += me.walrusLoop(n)
		whileval := me.eval(n.has[0])
		code += whileval.precode()
		code += "while (" + whileval.pop() + ") {\n"
		size := len(n.has)
		for ix < size && n.has[ix].is == "variable" {
			temp := n.has[ix]
			tempname := temp.idata.name
			tempv := me.hmfile.varInitFromData(temp.data(), tempname, false)
			me.scope.variables[tempname] = tempv
			ix++
		}
	} else {
		ix += 3
		vset := n.has[0]
		if vset.is != "=" {
			panic("for loop must start with assign")
		}
		vobj := vset.has[0]
		if vobj.is != "variable" {
			panic("for loop must assign a regular variable")
		}
		vexist := me.getvar(vobj.idata.name)
		if vexist == nil {
			code += me.declare(vobj) + ";\n" + fmc(me.depth)
		}
		vinit := me.compileAssign(vset)
		condition := me.eval(n.has[1]).code()
		inc := me.assignmentUpdate(n.has[2])
		code += "for (" + vinit + "; " + condition + "; " + inc + ") {\n"
	}
	code += me.eval(n.has[ix]).code()
	code += "\n" + fmc(me.depth) + "}"
	return codeBlockOne(n, code)
}
