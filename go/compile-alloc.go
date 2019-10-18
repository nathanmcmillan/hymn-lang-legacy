package main

import (
	"strconv"
)

func (me *cfile) temp() string {
	temp := "temp_" + strconv.Itoa(me.scope.temp)
	me.scope.temp++
	return temp
}

func (me *cfile) compileAllocClass(n *node) *codeblock {
	if _, ok := n.attributes["no-malloc"]; ok {
		return codeBlockOne(n, "")
	}

	_, useStack := n.attributes["stack"]

	data := n.data()
	typed := data.module.classNameSpace(data.typed)

	assign, local := n.attributes["assign"]

	cb := &codeblock{}

	var memberRef string
	if useStack {
		memberRef = "."
	} else {
		memberRef = "->"
	}

	if local {
		cl, _ := data.checkIsClass()
		code := ""
		if !useStack {
			code += "malloc(sizeof(" + typed + "))"
		}
		params := n.has
		for i, p := range params {
			clv := cl.variables[cl.variableOrder[i]]
			if !clv.isptr {
				p.attributes["stack"] = "true"
			}
			cassign := ";\n" + fmc(me.depth) + assign + memberRef + clv.name + " = "
			if p.is == "new" {
				temp := me.temp()
				p.attributes["assign"] = temp
				d := nodeInit("variable")
				d.idata = &idData{}
				d.idata.module = me.hmfile
				d.idata.name = temp
				d.copyType(p)
				decl := me.declare(d)
				value := me.eval(p).code()
				code2 := ";\n" + fmc(me.depth) + decl
				if clv.isptr {
					code2 += " = "
				}
				code2 += value
				cn := codeNode(p, code2)
				cn.value = temp
				code += cn.code
				code += cassign + cn.value
			} else {
				code += cassign + me.eval(p).code()
			}
		}
		cb.current = codeNode(n, code)

	} else {
		temp := me.temp()
		cb.current = codeNode(n, temp)
		n.attributes["assign"] = temp
		d := nodeInit("variable")
		d.idata = &idData{}
		d.idata.module = me.hmfile
		d.idata.name = temp
		d.copyType(n)
		decl := me.declare(d)
		value := me.eval(n).code()
		code := decl + " = " + value + me.maybeColon(value) + "\n"
		cn := codeNode(n, code)
		cn.value = temp
		cb.pre = codeNodeUpgrade(cn)
	}

	return cb
}

func (me *cfile) compileAllocEnum(n *node) *codeblock {
	if _, ok := n.attributes["no-malloc"]; ok {
		return codeBlockOne(n, "")
	}
	data := n.data()
	module := data.module
	en, un, _ := data.checkIsEnum()
	enumType := un.name
	if en.simple {
		enumBase := module.enumNameSpace(en.name)
		globalName := module.enumTypeName(enumBase, enumType)
		return codeBlockOne(n, globalName)
	}
	unionOf := en.types[enumType]
	code := ""
	code += module.unionFnNameSpace(en, unionOf) + "("
	if len(unionOf.types) == 1 {
		unionHas := n.has[0]
		code += me.eval(unionHas).code()
	} else {
		for ix := range unionOf.types {
			if ix > 0 {
				code += ", "
			}
			unionHas := n.has[ix]
			code += me.eval(unionHas).code()
		}
	}
	code += ")"
	return codeBlockOne(n, code)
}

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
	if _, ok := n.attributes["no-malloc"]; ok {
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
	if _, ok := n.attributes["no-malloc"]; ok {
		code = "[" + size + "]"
	} else {
		code = "hmlib_slice_init(" + size + ", sizeof(" + n.data().memberType.typeSig() + "))"
	}
	return codeBlockOne(n, code)
}
