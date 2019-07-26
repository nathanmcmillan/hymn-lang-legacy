package main

func (me *cfile) allocClass(n *node) *cnode {
	if _, ok := n.attributes["no-malloc"]; ok {
		return codeNode(n.is, n.value, n.typed, "")
	}

	data := me.hmfile.typeToVarData(n.typed)
	typed := data.module.classNameSpace(data.typed)

	code := "malloc(sizeof(" + typed + "))"
	return codeNode(n.is, n.value, n.typed, code)
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
