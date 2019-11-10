package main

func (me *cfile) compileAllocArray(n *node) *codeblock {
	size := ""
	parenthesis := false
	if len(n.has) > 0 {
		e := me.eval(n.has[0])
		size = e.code()
		if e.getType() != TokenInt {
			parenthesis = true
		}
	} else {
		size = sizeOfArray(n.data().full)
	}
	if _, ok := n.attributes["global"]; ok {
		return codeBlockOne(n, "["+size+"]")
	}
	memberType := n.data().memberType.typeSig()
	code := "malloc("
	if parenthesis {
		code += "("
	}
	code += size
	if parenthesis {
		code += ")"
	}
	code += " * sizeof(" + memberType + "))"
	return codeBlockOne(n, code)
}

func (me *cfile) compileAllocSlice(n *node) *codeblock {
	size := "0"
	capacity := ""
	has := len(n.has)
	if has > 0 {
		size = me.eval(n.has[0]).code()
		if has > 1 {
			capacity = me.eval(n.has[1]).code()
		}
	}
	code := ""
	if _, ok := n.attributes["global"]; ok {
		code = "[" + size + "]"
	} else {
		if capacity != "" {
			code = "hmlib_slice_init(sizeof(" + n.data().memberType.typeSig() + "), " + size + ", " + capacity + ")"
		} else {
			code = "hmlib_slice_simple_init(sizeof(" + n.data().memberType.typeSig() + "), " + size + ")"
		}
	}
	return codeBlockOne(n, code)
}

func (me *cfile) compileArrayToSlice(n *node) *codeblock {
	array := n.has[0]
	data := array.data()
	code := "hmlib_array_to_slice(" + array.idata.name + ", sizeof(" + data.memberType.typeSig() + "), " + data.sizeOfArray() + ")"
	return codeBlockOne(n, code)
}
