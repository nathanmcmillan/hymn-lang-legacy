package main

import (
	"strconv"
)

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
	pix := 0
	dict := false
	params := make([]*node, len(fn.args))
	for {
		if me.token.is == ")" {
			me.eat(")")
			break
		}
		if pix > 0 || dict {
			me.eat("delim")
		}
		if me.token.is == "id" && me.peek().is == ":" {
			argname := me.token.value
			me.eat("id")
			me.eat(":")
			param := me.calc()
			aix := fn.argDict[argname]
			arg := fn.args[aix]
			if param.typed != arg.typed && arg.typed != "?" {
				err := "parameter \"" + param.typed
				err += "\" does not match argument \"" + argname + "\" typed \"" + arg.typed + "\" for function \"" + name + "\""
				panic(me.fail() + err)
			}
			params[aix] = param
			dict = true

		} else if dict {
			panic(me.fail() + "regular paramater found after mapped parameter")
		} else {
			param := me.calc()
			arg := fn.args[pix]
			if param.typed != arg.typed && arg.typed != "?" {
				err := "parameter \"" + param.typed
				err += "\" does not match argument[" + strconv.Itoa(pix) + "] \"" + arg.typed + "\" for function \"" + name + "\""
				panic(me.fail() + err)
			}
			params[pix] = param
			pix++
		}
	}

	for ix, param := range params {
		if param == nil {
			arg := fn.args[ix]
			if arg.dfault != "" {
				dfault := nodeInit(arg.typed)
				dfault.typed = arg.typed
				dfault.value = arg.dfault
				n.push(dfault)
			}
		} else {
			n.push(param)
		}
	}

	return n
}
