package main

type node struct {
	is         string
	value      string
	idata      *idData
	fn         *function
	_vdata     *datatype
	attributes map[string]string
	parent     *node
	has        []*node
}

func nodeInit(is string) *node {
	n := &node{}
	n.is = is
	n.has = make([]*node, 0)
	n.attributes = make(map[string]string)
	return n
}

func (me *node) copy() *node {
	n := &node{}
	n.is = me.is
	n.value = me.value
	if me.idata != nil {
		n.idata = me.idata.copy()
	}
	if me.fn != nil {
		n.fn = me.fn.copy()
	}
	if me.data() != nil {
		n._vdata = me.data().copy()
	}
	n.attributes = make(map[string]string)
	for k, v := range me.attributes {
		n.attributes[k] = v
	}
	n.parent = me.parent
	n.has = make([]*node, len(me.has))
	for i, h := range me.has {
		n.has[i] = h.copy()
	}
	return n
}

func (me *node) prepend(leaf *node) {
	me.has = append([]*node{leaf}, me.has...)
}

func (me *node) push(leaf *node) {
	leaf.parent = me
	me.has = append(me.has, leaf)
}

func (me *node) copyDataOfNode(other *node) {
	me._vdata = other.data().copy()
}

func (me *node) copyTypeFromVar(other *variable) {
	me._vdata = other.data().copy()
}

func (me *node) data() *datatype {
	return me._vdata
}

func (me *node) copyData(data *datatype) {
	me._vdata = data.copy()
}

func (me *node) setData(data *datatype) {
	me._vdata = data
}
