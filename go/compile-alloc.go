package main

import (
	"fmt"
	"strconv"
)

type blockNode struct {
	pre     *blockNode
	current []*node
}

func (me *blockNode) flatten() []*node {
	flat := make([]*node, 0)
	for _, p := range me.pre.flatten() {
		flat = append(flat, p)
	}
	for _, c := range me.current {
		flat = append(flat, c)
	}
	return flat
}

func (me *cfile) tempClass(p *node) *cnode {
	temp := "temp_" + strconv.Itoa(me.scope.temp)
	me.scope.temp++
	p.attributes["assign"] = temp

	d := nodeInit("variable")
	d.idata = &idData{}
	d.idata.module = me.hmfile
	d.idata.name = temp
	d.copyType(p)
	decl := me.declare(d)

	code := ""
	code += ";\n" + fmc(me.depth) + decl + " = " + me.eval(p).code

	cn := codeNode(p, code)
	cn.value = temp
	return cn
}

func (me *cfile) allocClass(n *node) *cnode {
	if _, ok := n.attributes["no-malloc"]; ok {
		return codeNode(n, "")
	}

	_, useStack := n.attributes["use-stack"]
	useHeap := !useStack

	data := n.asVar()
	typed := data.module.classNameSpace(data.typed)

	code := ""

	var ptrCh string
	if useHeap {
		fmt.Println("ALLOC CLASS USE HEAP FOR MEMBER ::", typed)
		code += "malloc(sizeof(" + typed + "))"
		ptrCh = "->"
	} else {
		ptrCh = "."
	}

	assign, _ := n.attributes["assign"]
	cl, _ := data.checkIsClass()
	params := n.has
	for ix, p := range params {
		clv := cl.variables[cl.variableOrder[ix]]
		cassign := ";\n" + fmc(me.depth) + assign + ptrCh + clv.name + " = "

		if p.is == "new" {
			temp := me.tempClass(p)
			code += temp.code
			code += cassign + temp.value
		} else {
			code += cassign + me.eval(p).code
		}
	}

	return codeNode(n, code)
}

func (me *cfile) allocEnum(n *node) string {
	if _, ok := n.attributes["no-malloc"]; ok {
		return ""
	}
	data := n.vdata
	module := data.module
	en, un, _ := data.checkIsEnum()
	enumType := un.name
	if en.simple {
		enumBase := module.enumNameSpace(en.name)
		globalName := module.enumTypeName(enumBase, enumType)
		return globalName
	}
	unionOf := en.types[enumType]
	code := ""
	code += module.unionFnNameSpace(en, unionOf) + "("
	if len(unionOf.types) == 1 {
		unionHas := n.has[0]
		code += me.eval(unionHas).code
	} else {
		for ix := range unionOf.types {
			if ix > 0 {
				code += ", "
			}
			unionHas := n.has[ix]
			code += me.eval(unionHas).code
		}
	}
	code += ")"
	return code
}

func (me *cfile) allocArray(n *node) string {
	size := ""
	parenthesis := false
	if len(n.has) > 0 {
		e := me.eval(n.has[0])
		size = e.code
		if e.getType() != TokenInt {
			parenthesis = true
		}
	} else {
		size = sizeOfArray(n.asVar().full)
	}
	if _, ok := n.attributes["no-malloc"]; ok {
		return "[" + size + "]"
	}
	memberType := n.asVar().typeSig()
	code := "malloc("
	if parenthesis {
		code += "("
	}
	code += size
	if parenthesis {
		code += ")"
	}
	code += " * sizeof(" + memberType + "))"
	return code
}

func (me *cfile) allocSlice(n *node) string {
	size := "0"
	if len(n.has) > 0 {
		size = me.eval(n.has[0]).code
	}
	if _, ok := n.attributes["no-malloc"]; ok {
		return "[" + size + "]"
	}
	return "hmlib_slice_init(" + size + ", sizeof(" + n.asVar().memberType.typeSig() + "))"
}
