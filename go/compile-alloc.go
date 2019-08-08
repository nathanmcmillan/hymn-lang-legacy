package main

import "strconv"

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
	d.value = temp
	d.copyType(p)
	decl := me.declare(d)

	code := ""
	code += ";\n" + fmc(me.depth) + decl + temp + " = " + me.eval(p).code

	return codeNode(p.is, temp, p.typed, p.vdata, code)
}

func (me *cfile) allocClass(n *node) *cnode {
	if _, ok := n.attributes["no-malloc"]; ok {
		return codeNode(n.is, n.value, n.getType(), n.vdata, "")
	}

	_, useStack := n.attributes["use-stack"]
	useHeap := !useStack

	data := n.asVar(me.hmfile)
	typed := data.module.classNameSpace(data.typed)

	code := ""

	var ptrCh string
	if useHeap {
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

	return codeNode(n.is, n.value, n.typed, n.vdata, code)
}

func (me *cfile) allocEnum(module *hmfile, typed string, n *node) string {
	enumOf := module.enums[typed]
	if enumOf.simple {
		enumBase := module.enumNameSpace(typed)
		enumType := n.value
		globalName := module.enumTypeName(enumBase, enumType)
		return globalName
	}
	if _, ok := n.attributes["no-malloc"]; ok {
		return ""
	}
	enumType := n.value
	unionOf := enumOf.types[enumType]
	code := ""
	code += module.unionFnNameSpace(enumOf, unionOf) + "("
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
