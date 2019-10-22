package main

func (me *cfile) allocArray(n *node) *codeblock {
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
	memberType := n.data().typeSig()
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

func (me *cfile) allocSlice(n *node) *codeblock {
	size := "0"
	if len(n.has) > 0 {
		size = me.eval(n.has[0]).code()
	}
	code := ""
	if _, ok := n.attributes["global"]; ok {
		code = "[" + size + "]"
	} else {
		code = "hmlib_slice_init(" + size + ", sizeof(" + n.data().memberType.typeSig() + "))"
	}
	return codeBlockOne(n, code)
}
