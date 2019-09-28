package main

import (
	"strconv"
)

func (me *parser) pushSigParams(n *node, sig *fnSig) {
	params := make([]*node, 0)
	me.eat("(")
	ix := 0
	for {
		if me.token.is == ")" {
			me.eat(")")
			break
		} else if ix > 0 {
			me.eat(",")
		}
		param := me.calc(0)
		arg := sig.args[ix]
		if param.asVar().notEqual(arg.vdat) && arg.vdat.full != "?" {
			err := "parameter \"" + param.getType()
			err += "\" does not match argument[" + strconv.Itoa(ix) + "] \"" + arg.vdat.full + "\" for function"
			panic(me.fail() + err)
		}
		params = append(params, param)
	}
	for _, param := range params {
		n.push(param)
	}
}

func (me *parser) pushParams(name string, n *node, pix int, params []*node, fn *function) {
	me.eat("(")
	min := pix
	dict := false
	for {
		if me.token.is == ")" {
			me.eat(")")
			break
		} else if pix > min || dict {
			me.eat(",")
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
	me.pushParams(name, n, 1, params, fn)
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
	me.pushParams(name, n, 0, params, fn)
	return n
}

func (me *parser) parseFn(module *hmfile) *node {
	if me.peek().is == "(" {
		return me.call(module)
	}

	name := me.token.value
	fn := module.functions[name]
	me.eat("id")
	n := nodeInit("fn-ptr")
	n.vdata = fn.asVar()
	n.fn = fn

	return n
}
