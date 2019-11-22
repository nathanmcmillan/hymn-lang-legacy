package main

func (me *cfile) compileLoop(op string, n *node) *codeblock {
	code := "while (true) {\n"
	code += me.eval(n.has[0]).code()
	code += me.maybeNewLine(code) + fmc(me.depth) + "}"
	return codeBlockOne(n, code)
}

func (me *cfile) compileWhile(op string, n *node) *codeblock {
	code := me.walrusLoop(n)
	whileval := me.eval(n.has[0])
	code += whileval.precode()
	code += "while (" + whileval.pop() + ") {\n"
	size := len(n.has)
	ix := 1
	for ix < size && n.has[ix].is == "variable" {
		temp := n.has[ix]
		tempname := temp.idata.name
		tempv := me.hmfile.varInitFromData(temp.data(), tempname, false)
		me.scope.variables[tempname] = tempv
		ix++
	}

	code += me.eval(n.has[ix]).code()
	code += me.maybeNewLine(code) + fmc(me.depth) + "}"
	return codeBlockOne(n, code)
}

func (me *cfile) compileFor(op string, n *node) *codeblock {
	code := ""
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
	vinit := me.compileAssign(vset).code()
	condition := me.eval(n.has[1]).code()
	inc := me.assignmentUpdate(n.has[2])
	code += "for (" + vinit + "; " + condition + "; " + inc + ") {\n"
	code += me.eval(n.has[3]).code()
	code += me.maybeNewLine(code) + fmc(me.depth) + "}"
	return codeBlockOne(n, code)
}

func (me *cfile) compileIterate(op string, n *node) *codeblock {
	code := ""
	item := ""
	var1 := ""
	var2 := ""
	index := "index_" + me.temp()
	ix := 0
	size := len(n.has)
	if size == 3 {
		a := n.has[0]
		me.scope.variables[a.idata.name] = me.hmfile.varInitFromData(a.data(), a.idata.name, false)
		item = a.idata.name
		var1 = a.idata.name
		ix = 1
	} else if size == 4 {
		a := n.has[0]
		me.scope.renaming[a.idata.name] = index
		me.scope.variables[index] = me.hmfile.varInitFromData(a.data(), index, false)
		b := n.has[1]
		me.scope.variables[b.idata.name] = me.hmfile.varInitFromData(b.data(), b.idata.name, false)
		item = b.idata.name
		var1 = a.idata.name
		var2 = b.idata.name
		ix = 2
	} else {
		panic("")
	}
	array := n.has[ix]
	arrayname := ""
	if array.is == "variable" {
		arrayname = array.idata.name
	} else {
		arrayname = "iterate_" + me.temp()
		array.attributes["assign"] = arrayname
		me.scope.variables[arrayname] = me.hmfile.varInitFromData(array.data(), arrayname, false)
		code += array.data().typeSig() + arrayname + " = "
		code += me.eval(array).code()
		code += "\n" + fmc(me.depth)
	}
	getlen := ""
	if array.data().checkIsArray() {
		getlen = array.data().sizeOfArray()
	} else {
		getlen = "size_" + me.temp()
		lennode := nodeInit("call")
		lennode.fn = me.hmfile.hmlib.functions[libLength]
		lennode.push(array)
		len := me.eval(lennode)
		code += "int " + getlen + " = " + len.code() + ";\n" + fmc(me.depth)
	}
	block := me.eval(n.has[ix+1])

	delete(me.scope.renaming, var1)
	if var2 != "" {
		delete(me.scope.renaming, var2)
	}

	code += "int " + index + ";\n" + fmc(me.depth)
	code += "for (" + index + " = 0; " + index + " < " + getlen + "; " + index + "++) {\n"
	code += fmc(me.depth + 1)
	code += array.data().memberType.typeSig() + " " + item + " = " + arrayname + "[" + index + "];\n"
	code += block.code()
	code += me.maybeNewLine(code) + fmc(me.depth) + "}"

	return codeBlockOne(n, code)
}
