package main

func (me *cfile) compileIf(n *node) *codeblock {
	hsize := len(n.has)
	code := ""
	code += me.walrusIf(n)
	code += "if (" + me.eval(n.has[0]).code() + ") {\n"
	ix := 1
	for ix < hsize && n.has[ix].is == "variable" {
		temp := n.has[ix]
		tempname := temp.idata.name
		tempv := me.hmfile.varInitFromData(temp.data(), tempname, false)
		me.scope.variables[tempname] = tempv
		ref := me.eval(n.has[ix].has[0]).code()
		code += fmc(me.depth + 1)
		code += fmtassignspace(temp.data().typeSig()) + tempname + " = " + ref + ";\n"
		ix++
	}
	thenCode := me.eval(n.has[ix]).code()
	code += me.maybeFmc(thenCode, me.depth+1) + thenCode + me.maybeColon(thenCode)
	code += "\n" + fmc(me.depth) + "}"
	ix++
	for ix < hsize && n.has[ix].is == "elif" {
		elif := n.has[ix]
		code += " else if (" + me.eval(elif.has[0]).code() + ") {\n"
		elsize := len(elif.has)
		ixo := 1
		for ixo < elsize && elif.has[ixo].is == "variable" {
			temp := elif.has[ixo]
			tempname := temp.idata.name
			tempv := me.hmfile.varInitFromData(temp.data(), tempname, false)
			me.scope.variables[tempname] = tempv
			ref := me.eval(elif.has[ixo].has[0]).code()
			code += fmc(me.depth + 1)
			code += fmtassignspace(temp.data().typeSig()) + tempname + " = " + ref + ";\n"
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

func (me *cfile) compileFor(n *node) *codeblock {
	size := len(n.has)
	ix := 0
	code := ""
	if size > 2 {
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
	} else if size > 1 {
		ix++
		code += me.walrusLoop(n)
		code += "while (" + me.eval(n.has[0]).code() + ") {\n"
	} else {
		code += "while (true) {\n"
	}
	code += me.eval(n.has[ix]).code()
	code += "\n" + fmc(me.depth) + "}"
	return codeBlockOne(n, code)
}
