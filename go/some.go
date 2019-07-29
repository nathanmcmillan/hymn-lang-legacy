package main

func (me *parser) some() *node {
	me.eat("some")
	n := nodeInit("some")
	n.typed = "some<?>"
	return n
}

func (me *parser) none() *node {
	me.eat("none")
	n := nodeInit("none")
	n.typed = "none<?>"
	return n
}

func (me *parser) maybe() *node {
	me.eat("maybe")
	n := nodeInit("maybe")
	n.typed = "maybe<?>"
	return n
}
