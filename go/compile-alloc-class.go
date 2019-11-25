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
	if _, ok := n.attributes["global"]; ok {
		return codeBlockOne(n, "")
	}

	data := n.data()
	typed := data.module.classNameSpace(data.dtype.cname())

	_, useStack := n.attributes["stack"]
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
			if !clv.data().isptr {
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
				d.copyDataOfNode(p)
				decl := me.declare(d)
				value := me.eval(p).code()
				code2 := ";\n" + fmc(me.depth) + decl
				if clv.data().isptr {
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
		d.copyDataOfNode(n)
		decl := me.declare(d)
		value := me.eval(n).code()
		code := decl + " = " + value + me.maybeColon(value) + "\n"
		cn := codeNode(n, code)
		cn.value = temp
		cb.prepend(codeNodeUpgrade(cn))
	}

	return cb
}
