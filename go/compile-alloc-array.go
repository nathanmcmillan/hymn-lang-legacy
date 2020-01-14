package main

import (
	"strconv"
)

func (me *cfile) compileAllocArray(n *node) *codeblock {
	size := ""
	parenthesis := false
	has := len(n.has)
	var items *node
	if has > 0 {
		a := n.has[0]
		if a.is == "items" {
			items = a
		} else {
			e := me.eval(a)
			size = e.code()
			if e.getType() != TokenInt {
				parenthesis = true
			}
		}
	} else {
		size = n.data().sizeOfArray()
	}
	if items != nil {
		sizeint, er := strconv.Atoi(size)
		if er != nil || len(items.has) > sizeint {
			size = strconv.Itoa(len(items.has))
		}
	}
	if _, ok := n.attributes["global"]; ok {
		return codeBlockOne(n, "["+size+"]")
	}
	memberType := n.data().getmember().typeSig()
	code := "malloc("
	if parenthesis {
		code += "("
	}
	code += size
	if parenthesis {
		code += ")"
	}
	code += " * sizeof(" + memberType + "))"
	if items != nil {
		name := n.attributes["assign"]
		code += ";"
		for i, item := range items.has {
			code += "\n" + fmc(me.depth)
			code += name + "[" + strconv.Itoa(i) + "] = "
			code += me.eval(item).code() + ";"
		}
	}
	return codeBlockOne(n, code)
}

func (me *cfile) compileAllocSlice(n *node) *codeblock {
	size := "0"
	capacity := ""
	has := len(n.has)
	var items *node
	if has > 0 {
		a := n.has[0]
		if a.is == "items" {
			items = a
		} else {
			size = me.eval(a).code()
			if has > 1 {
				b := n.has[1]
				if b.is == "items" {
					items = b
				} else {
					capacity = me.eval(b).code()
				}
				if has > 2 {
					items = n.has[2]
				}
			}
		}
	}
	if items != nil {
		sizeint, er := strconv.Atoi(size)
		if er != nil || len(items.has) > sizeint {
			size = strconv.Itoa(len(items.has))
		}
	}
	code := ""
	if _, ok := n.attributes["global"]; ok {
		code = "[" + size + "]"
	} else {
		me.libReq.add(HmLibSlice)
		if capacity != "" {
			code = "hmlib_slice_init(sizeof(" + n.data().getmember().typeSig() + "), " + size + ", " + capacity + ")"
		} else {
			code = "hmlib_slice_simple_init(sizeof(" + n.data().getmember().typeSig() + "), " + size + ")"
		}
	}
	if items != nil {
		name := n.attributes["assign"]
		code += ";"
		for i, item := range items.has {
			code += "\n" + fmc(me.depth)
			code += name + "[" + strconv.Itoa(i) + "] = "
			code += me.eval(item).code() + ";"
		}
	}
	return codeBlockOne(n, code)
}

func (me *cfile) compileArrayToSlice(n *node) *codeblock {
	array := n.has[0]
	data := array.data()
	me.libReq.add(HmLibSlice)
	code := "hmlib_array_to_slice(" + array.idata.name + ", sizeof(" + data.getmember().typeSig() + "), " + data.sizeOfArray() + ")"
	return codeBlockOne(n, code)
}
