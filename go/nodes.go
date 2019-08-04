package main

type node struct {
	is         string
	value      string
	typed      string
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
