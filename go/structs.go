package main

type scope struct {
	root      *scope
	tempID    int
	fn        *function
	variables map[string]*variable
	renaming  map[string]string
}

type hasGenerics interface {
	getGenerics() []string
}

func scopeInit(root *scope) *scope {
	s := &scope{}
	s.root = root
	s.variables = make(map[string]*variable)
	s.renaming = make(map[string]string)
	return s
}

func codeNode(node *node, code string) *cnode {
	c := &cnode{}
	c.is = node.is
	c.value = node.value
	c.copyData(node.data())
	c.code = code
	c.has = make([]*cnode, 0)
	return c
}

func (me *cnode) push(n *cnode) {
	me.has = append(me.has, n)
}

func isNumber(t string) bool {
	_, ok := numbers[t]
	return ok
}

func canCastToNumber(t string) bool {
	if t == TokenChar {
		return true
	}
	_, ok := numbers[t]
	return ok
}

func isAnyIntegerType(t string) bool {
	_, ok := integerTypes[t]
	return ok
}

type allocHint struct {
	stack bool
	array bool
	slice bool
	size  int
}

func checkIsPrimitive(t string) bool {
	_, ok := primitives[t]
	return ok
}

func getCName(primitive string) (string, bool) {
	if name, ok := typeToCName[primitive]; ok {
		return name, true
	}
	return primitive, false
}
