package main

func (me *parser) some() *node {
	me.eat("some")
	n := nodeInit("some")
	n.typed = "some<?>"
	return n
}

func (me *parser) none() *node {
	me.eat("none")
	me.eat("<")
	option := me.declareType(true).typed
	me.eat(">")
	typed := "none<" + option + ">"
	me.defineMaybeImpl(typed)

	n := nodeInit("none")
	n.typed = typed
	return n
}

func (me *parser) maybe() *node {
	me.eat("maybe")
	me.eat("<")
	option := me.declareType(true).typed
	me.eat(">")
	typed := "maybe<" + option + ">"
	me.defineMaybeImpl(typed)

	n := nodeInit("maybe")
	n.typed = typed
	return n
}
