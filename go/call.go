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
		arg := sig.args[ix]
		param := me.calc(0, arg.data())
		if param.data().notEquals(arg.data()) && !arg.data().isAnyType() {
			err := "parameter \"" + param.data().print()
			err += "\" does not match argument[" + strconv.Itoa(ix) + "] \"" + arg.data().print() + "\" of function signature \"" + sig.print() + "\""
			panic(me.fail() + err)
		}
		params = append(params, param)
	}
	for _, param := range params {
		n.push(param)
	}
}

func (me *parser) functionParams(name string, n *node, pix int, params []*node, fn *function, lazy bool) []*node {
	me.eat("(")
	if me.token.is == "line" {
		me.eat("line")
	}
	min := pix
	dict := false
	size := len(fn.args)
	for {
		if me.token.is == ")" {
			break
		} else if pix > min || dict {
			if me.token.is == "line" {
				me.eat("line")
				if me.token.is == ")" {
					break
				}
			} else {
				me.eat(",")
			}
		}
		if me.token.is == "id" && me.peek().is == ":" {
			argname := me.token.value
			me.eat("id")
			me.eat(":")
			param := me.calc(0, nil)
			aix := fn.argDict[argname]
			arg := fn.args[aix]
			if param.data().notEquals(arg.data()) && !arg.data().isAnyType() {
				err := "parameter \"" + param.data().print()
				err += "\" does not match argument \"" + argname + "\" typed \"" + arg.data().print() + "\" for function \"" + name + "\""
				panic(me.fail() + err)
			}
			params[aix] = param
			dict = true

		} else if dict {
			panic(me.fail() + "regular paramater found after mapped parameter")
		} else {
			var arg *funcArg
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
					param = me.defaultValue(arg.data(), "")
				}
				params[pix] = param
			} else {
				param := me.calc(0, nil)
				if arg == nil {
					arg = fn.args[pix]
				}
				if param.data().notEquals(arg.data()) && !arg.data().isAnyType() {
					err := "Parameter: " + param.data().print()
					err += " does not match expected: " + arg.data().print() + " for function: " + name
					panic(me.fail() + err)
				}
				params[pix] = param
			}
			pix++
		}
	}
	me.eat(")")
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
	n.fn = fn
	n.copyData(fn.returns)
	params := make([]*node, len(fn.args))
	params[0] = root
	me.functionParams(fn.getclsname(), n, 1, params, fn, false)
	return n
}

func (me *parser) call(module *hmfile) *node {
	name := me.token.value
	me.eat("id")
	var order []*datatype
	var fn *function
	bfn, ok := module.getFunction(name)
	if !ok {
		panic(me.fail() + "Missing function '" + name + "'")
	}
	lazy := false
	if bfn.generics != nil {
		if me.token.is == "<" {
			order, _, _ = me.genericHeader()
			name += genericslist(order)
			gfn, ok := module.getFunction(name)
			if ok {
				fn = gfn
			} else {
				mapping := make(map[string]*datatype)
				for i, g := range bfn.generics {
					mapping[g] = order[i]
				}
				fn = remapFunctionImpl(name, mapping, bfn)
			}
		} else {
			lazy = true
		}
	} else {
		fn = bfn
	}
	n := nodeInit("call")
	n.fn = fn
	n.copyData(fn.returns)
	params := make([]*node, len(fn.args))
	me.functionParams(name, n, 0, params, fn, lazy)
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
