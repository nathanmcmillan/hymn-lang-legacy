package main

type node struct {
	is    string
	value string
	// typed      string
	vdata      *varData
	attributes map[string]string
	has        []*node
}

func nodeInit(is string) *node {
	n := &node{}
	n.is = is
	n.has = make([]*node, 0)
	n.attributes = make(map[string]string)
	return n
}

func (me *node) push(n *node) {
	me.has = append(me.has, n)
}

func (me *node) copyType(other *node) {
	me.typed = other.typed
	me.vdata = other.vdata
}

func (me *node) copyTypeFromVar(other *variable) {
	me.typed = other.typed
	me.vdata = other.vdat
}

func (me *node) getType() string {
	if me.vdata != nil {
		return me.vdata.full
	}
	return me.typed
}

func (me *node) asVar(module *hmfile) *varData {
	if me.vdata != nil {
		return me.vdata
	}
	me.vdata = module.typeToVarData(me.typed)
	return me.vdata
}
