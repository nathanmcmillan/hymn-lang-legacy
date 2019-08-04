package main

import (
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
			if me.hmfile.typeToVarData(param.typed).notEqual(arg.vdat) && arg.typed != "?" {
				err := "parameter \"" + param.typed
				err += "\" does not match argument \"" + argname + "\" typed \"" + arg.typed + "\" for function \"" + name + "\""
				panic(me.fail() + err)
			}
			params[aix] = param
			dict = true

		} else if dict {
			panic(me.fail() + "regular paramater found after mapped parameter")
		} else {
			param := me.calc(0)
			arg := fn.args[pix]
			if me.hmfile.typeToVarData(param.typed).notEqual(arg.vdat) && arg.typed != "?" {
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
			if arg.dfault == "" {
				panic(me.fail() + "argument[" + strconv.Itoa(pix) + "] is missing")
			}
			dfault := nodeInit(arg.typed)
			dfault.typed = arg.typed
			dfault.value = arg.dfault
			n.push(dfault)
		} else {
			n.push(param)
		}
	}
}

func (me *parser) callClassFunction(module *hmfile, root *node, c *class, fn *function) *node {
	n := nodeInit("call")
	name := me.nameOfClassFunc(c.name, fn.name)
	if module == me.hmfile {
		n.value = name
	} else {
		n.value = module.name + "." + name
	}
	n.typed = fn.typed.full
	params := make([]*node, len(fn.args))
	params[0] = root
	pix := 1
	me.pushParams(name, n, pix, 1, params, fn)
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
	n.typed = fn.typed.full
	params := make([]*node, len(fn.args))
	pix := 0
	me.pushParams(name, n, pix, 0, params, fn)
	return n
}
