package main

import (
	"strconv"
)

func (me *cfile) temp() string {
	temp := "temp" + strconv.Itoa(me.scope.tempID)
	me.scope.tempID++
	return temp
}

func (me *cfile) compileAllocClass(n *node) *codeblock {

	data := n.data()
	typed := data.cname()

	_, useStack := n.attributes["stack"]
	assignName, local := n.attributes["assign"]
	var assign *variable
	if local {
		assign = me.getvar(assignName)
	}

	cb := &codeblock{}

	var memberRef string
	if useStack {
		memberRef = "."
	} else {
		memberRef = "->"
	}

	if local {
		cl, _ := data.isClass()
		code := ""
		if !useStack {
			me.stdReq.add(CStdLib)
			code += "malloc(sizeof(" + typed + "))"
		}
		params := n.has
		for i, p := range params {
			clv := cl.variables[i]
			if !clv.data().isPointer() {
				p.attributes["stack"] = "true"
			}
			cassign := ";\n" + fmc(me.depth) + assign.cname + memberRef + clv.name + " = "
			if p.is == "new" || (p.is == "enum" && p.data().union != nil) {
				temp := me.temp()
				p.attributes["assign"] = temp
				d := nodeInit("variable")
				d.idata = newidvariable(me.hmfile, temp)
				d.copyDataOfNode(p)
				decl := me.compileDeclare(d)
				value := me.eval(p).code()
				code2 := ";\n" + fmc(me.depth) + decl
				if clv.data().isPointer() {
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
		nodeCopy := n.copy()
		nodeCopy.attributes["assign"] = temp
		d := nodeInit("variable")
		d.idata = newidvariable(me.hmfile, temp)
		d.copyDataOfNode(nodeCopy)
		decl := me.compileDeclare(d)
		value := me.eval(nodeCopy).code()
		code := decl + " = " + value + me.maybeColon(value) + "\n"
		cn := codeNode(nodeCopy, code)
		cn.value = temp
		cb.prepend(codeNodeUpgrade(cn))
	}

	return cb
}
