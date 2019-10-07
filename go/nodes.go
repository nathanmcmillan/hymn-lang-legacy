package main

import "fmt"

type node struct {
	is         string
	value      string
	idata      *idData
	fn         *function
	vdata      *varData
	attributes map[string]string
	has        []*node
	trunk      *node
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

func (me *node) push(leaf *node) {
	me.has = append(me.has, leaf)
	leaf.trunk = me
}

func (me *node) copyType(other *node) {
	me.vdata = other.vdata
}

func (me *node) copyTypeFromVar(other *variable) {
	me.vdata = other.vdat
}

func (me *node) getType() string {
	return me.vdata.full
}

func (me *node) asVar() *varData {
	return me.vdata
}
