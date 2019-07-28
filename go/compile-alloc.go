package main

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

// HmConstructorVec *temp_0 = calloc(1, sizeof(HmConstructorVec));
// temp_0->x = 2 * 5;
// temp_0->y = 2 * 6;
// temp_0->z = 2 * 7;
// HmConstructorAttributeVec *temp_1 = calloc(1, sizeof(HmConstructorAttributeVec));
// temp_1->on = true;
// temp_1->has = temp_0;
// HmConstructorAttributeAttributeVec *const z = calloc(1, sizeof(HmConstructorAttributeAttributeVec));
// z->on = true;
// z->has = temp_1;

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
