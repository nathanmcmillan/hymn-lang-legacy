package main

func (me *cfile) compileLoop(op string, n *node) *codeblock {
	code := "while (true) {\n"
	code += me.eval(n.has[0]).code()
	code += me.maybeNewLine(code) + fmc(me.depth) + "}"
	return codeBlockOne(n, code)
}

func (me *cfile) compileWhile(op string, n *node) *codeblock {
	code := ""
	whileval := me.eval(n.has[0])
	code += whileval.precode()
	code += "while ("
	code += whileval.pop()
	code += ") {\n"
	size := len(n.has)
	ix := 1
	for ix < size && n.has[ix].is == "variable" {
		temp := n.has[ix]
		tempname := temp.idata.name
		tempv := temp.data().getnamedvariable(tempname, false)
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
		panic(me.fail(n) + "for loop must start with assign")
	}
	vobj := vset.has[0]
	if vobj.is != "variable" {
		panic(me.fail(n) + "for loop must assign a regular variable")
	}
	vexist := me.getvar(vobj.idata.name)
	if vexist == nil {
		code += me.compileDeclare(vobj) + ";\n" + fmc(me.depth)
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
		me.scope.variables[a.idata.name] = a.data().getnamedvariable(a.idata.name, false)
		item = a.idata.name
		var1 = a.idata.name
		ix = 1
	} else if size == 4 {
		a := n.has[0]
		me.scope.renaming[a.idata.name] = index
		me.scope.variables[index] = a.data().getnamedvariable(index, false)
		var1 = a.idata.name
		b := n.has[1]
		if b.idata.name == "_" {
			item = ""
		} else {
			me.scope.variables[b.idata.name] = b.data().getnamedvariable(b.idata.name, false)
			item = b.idata.name
			var2 = b.idata.name
		}
		ix = 2
	} else {
		panic(me.fail(n) + "Unexpected node size for iterator")
	}
	array := n.has[ix]
	arrayname := ""
	if item != "" {
		if array.is == "variable" {
			arrayname = array.idata.name
		} else {
			arrayname = "iterate_" + me.temp()
			array.attributes["assign"] = arrayname
			me.scope.variables[arrayname] = array.data().getnamedvariable(arrayname, false)
			code += array.data().typeSig(me) + arrayname + " = "
			code += me.eval(array).code()
			code += ";\n" + fmc(me.depth)
		}
	}
	getlen := ""
	if array.data().isArray() {
		getlen = array.data().arraySize()
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
	if item != "" {
		code += fmc(me.depth + 1)
		code += fmtassignspace(array.data().getmember().typeSig(me)) + item + " = " + arrayname + "[" + index + "];\n"
	}
	code += block.code()
	code += me.maybeNewLine(code) + fmc(me.depth) + "}"

	return codeBlockOne(n, code)
}
