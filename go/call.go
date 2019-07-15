package main

func (me *parser) pushParam(call *node, arg *variable) {
	param := me.calc()
	if param.typed != arg.typed && arg.typed != "?" {
		panic(me.fail() + "argument " + arg.typed + " does not match parameter " + param.typed)
	}
	call.push(param)
}

func (me *parser) callClassFunction(module *hmfile, root *node, c *class, fn *function) *node {
	n := nodeInit("call")
	name := me.nameOfClassFunc(c.name, fn.name)
	if module == me.hmfile {
		n.value = name
	} else {
		n.value = module.name + "." + name
	}
	n.typed = fn.typed
	n.push(root)
	me.eat("(")
	for ix, arg := range fn.args {
		if ix == 0 {
			continue
		}
		if ix > 1 {
			me.eat("delim")
		}
		me.pushParam(n, arg)
	}
	me.eat(")")
	return n
}

func (me *parser) call(module *hmfile) *node {
	token := me.token
	name := token.value
	fn := module.functions[name]
	me.eat("id")
	n := nodeInit("call")
	if module == me.hmfile {
		n.value = name
	} else {
		n.value = module.name + "." + name
	}
	n.typed = fn.typed
	me.eat("(")
	for ix, arg := range fn.args {
		if ix > 0 {
			me.eat("delim")
		}
		me.pushParam(n, arg)
	}
	me.eat(")")
	return n
}
