package main

import (
	"strconv"
	"strings"
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
		if param.data().notEqual(arg.data()) && arg.data().full != "?" {
			err := "parameter \"" + param.getType()
			err += "\" does not match argument[" + strconv.Itoa(ix) + "] \"" + arg.data().full + "\" of function signature \"" + sig.print() + "\""
			panic(me.fail() + err)
		}
		params = append(params, param)
	}
	for _, param := range params {
		n.push(param)
	}
}

func (me *parser) pushParams(name string, n *node, pix int, params []*node, fn *function) []*node {
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
			if param.data().notEqual(arg.data()) && arg.data().full != "?" {
				err := "parameter \"" + param.getType()
				err += "\" does not match argument \"" + argname + "\" typed \"" + arg.data().full + "\" for function \"" + name + "\""
				panic(me.fail() + err)
			}
			params[aix] = param
			dict = true

		} else if dict {
			panic(me.fail() + "regular paramater found after mapped parameter")
		} else {
			var arg *funcArg
			size := len(fn.args)
			if pix >= size {
				if fn.argVariadic != nil {
					arg = fn.argVariadic
					params = append(params, nil)
				} else {
					panic(me.fail() + "function \"" + name + "\" argument count exceeds parameter count")
				}
			}
			if me.token.is == "_" {
				me.eat("_")
				var param *node
				if arg == nil {
					arg = fn.args[pix]
				}
				if arg.defaultNode != nil {
					param = arg.defaultNode
				} else {
					param = me.defaultValue(arg.data())
				}
				params[pix] = param
			} else {
				param := me.calc(0)
				if arg == nil {
					arg = fn.args[pix]
				}
				if param.data().notEqual(arg.data()) && arg.data().full != "?" {
					err := "parameter \"" + param.getType()
					err += "\" does not match argument[" + strconv.Itoa(pix) + "] \"" + arg.data().full + "\" for function \"" + name + "\""
					panic(me.fail() + err)
				}
				params[pix] = param
			}
			pix++
		}
	}
	for ix, param := range params {
		if param == nil {
			var arg *funcArg
			if ix < len(fn.args) {
				arg = fn.args[ix]
			} else {
				arg = fn.argVariadic
			}
			if arg.defaultNode == nil {
				panic(me.fail() + "argument[" + strconv.Itoa(pix) + "] is missing")
			}
			n.push(arg.defaultNode)
		} else {
			n.push(param)
		}
	}
	return params
}

func (me *parser) callClassFunction(module *hmfile, root *node, c *class, fn *function) *node {
	n := nodeInit("call")
	name := fn.nameOfClassFunc()
	n.fn = fn
	n.copyData(fn.returns)
	params := make([]*node, len(fn.args))
	params[0] = root
	me.pushParams(name, n, 1, params, fn)
	return n
}

func (me *parser) call(module *hmfile) *node {
	name := me.token.value
	me.eat("id")
	if me.token.is == "<" {
		order, _ := me.genericHeader()
		name += "<" + strings.Join(order, ",") + ">"
	}
	fn, _ := module.getFunction(name)
	n := nodeInit("call")
	n.fn = fn
	n.copyData(fn.returns)
	params := make([]*node, len(fn.args))
	me.pushParams(name, n, 0, params, fn)
	return n
}

func (me *parser) parseFn(module *hmfile) *node {
	if me.peek().is == "(" || me.peek().is == "<" {
		return me.call(module)
	}

	name := me.token.value
	fn := module.functions[name]
	me.eat("id")
	n := nodeInit("fn-ptr")
	n.copyData(fn.data())
	n.fn = fn

	return n
}
