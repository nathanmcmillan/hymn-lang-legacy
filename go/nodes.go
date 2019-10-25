package main

import "fmt"

type node struct {
	is         string
	value      string
	idata      *idData
	fn         *function
	_vdata     *varData
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

func (me *node) copy() *node {
	fmt.Println("TODO :: node copy")
	n := &node{}
	n.is = me.is
	n.value = me.value
	// n.idata = me.idata.copy()
	n.fn = me.fn.copy()
	return n
}

func (me *node) prepend(leaf *node) {
	me.has = append([]*node{leaf}, me.has...)
}

func (me *node) push(leaf *node) {
	me.has = append(me.has, leaf)
}

func (me *node) copyDataOfNode(other *node) {
	me._vdata = other.data().copy()
}

func (me *node) copyTypeFromVar(other *variable) {
	me._vdata = other.data().copy()
}

func (me *node) getType() string {
	return me.data().full
}

func (me *node) data() *varData {
	return me._vdata
}

func (me *node) copyData(data *varData) {
	me._vdata = data.copy()
}
