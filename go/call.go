package main

import (
	"fmt"
	"strconv"
)

func (me *parser) pushParams(name string, n *node, pix, min int, params []*node, fn *function) {
	me.eat("(")
	dict := false
	for {
		if me.token.is == ")" {
			me.eat(")")
			break
		}
		if pix > min || dict {
			me.eat("delim")
		}
		if me.token.is == "id" && me.peek().is == ":" {
			argname := me.token.value
			me.eat("id")
			me.eat(":")
			param := me.calc(0)
			aix := fn.argDict[argname]
			arg := fn.args[aix]
			if param.asVar().notEqual(arg.vdat) && arg.vdat.full != "?" {
				err := "parameter \"" + param.getType()
				err += "\" does not match argument \"" + argname + "\" typed \"" + arg.vdat.full + "\" for function \"" + name + "\""
				panic(me.fail() + err)
			}
			params[aix] = param
			dict = true

		} else if dict {
			panic(me.fail() + "regular paramater found after mapped parameter")
		} else {
			param := me.calc(0)
			arg := fn.args[pix]
			if param.asVar().notEqual(arg.vdat) && arg.vdat.full != "?" {
				err := "parameter \"" + param.getType()
				err += "\" does not match argument[" + strconv.Itoa(pix) + "] \"" + arg.vdat.full + "\" for function \"" + name + "\""
				panic(me.fail() + err)
			}
			params[pix] = param
			pix++
		}
	}
	for ix, param := range params {
		if param == nil {
			arg := fn.args[ix]
			if arg.defaultNode == nil {
				panic(me.fail() + "argument[" + strconv.Itoa(pix) + "] is missing")
			}
			n.push(arg.defaultNode)
		} else {
			n.push(param)
		}
	}
}

func (me *parser) callClassFunction(module *hmfile, root *node, c *class, fn *function) *node {
	n := nodeInit("call")
	name := nameOfClassFunc(c.name, fn.name)
	n.fn = fn
	n.vdata = fn.typed
	params := make([]*node, len(fn.args))
	params[0] = root
	pix := 1
	me.pushParams(name, n, pix, 1, params, fn)
	return n
}

func (me *parser) call(module *hmfile) *node {
	name := me.token.value
	fn := module.functions[name]
	me.eat("id")
	n := nodeInit("call")
	n.fn = fn
	n.vdata = fn.typed
	params := make([]*node, len(fn.args))
	pix := 0
	me.pushParams(name, n, pix, 0, params, fn)
	return n
}

func (me *parser) parseFn(module *hmfile) *node {
	if me.peek().is == "(" {
		return me.call(module)
	}

	name := me.token.value

	fmt.Println("FUNCTION PTR ::", name)

	fn := module.functions[name]
	me.eat("id")
	n := nodeInit("function-ptr")
	n.vdata = fn.asVar()
	n.fn = fn

	return n
}
