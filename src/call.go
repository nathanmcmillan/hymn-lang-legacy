package main

import (
	"fmt"
	"strconv"
)

func (me *parser) pushSigParams(n *node, sig *fnSig) *parseError {
	params := make([]*node, 0)
	if er := me.eat("("); er != nil {
		return er
	}
	ix := 0
	for {
		if me.token.is == ")" {
			if er := me.eat(")"); er != nil {
				return er
			}
			break
		} else if ix > 0 {
			if er := me.eat(","); er != nil {
				return er
			}
		}
		arg := sig.args[ix]
		param, er := me.calc(0, arg.data())
		if er != nil {
			return er
		}
		if param.data().notEquals(arg.data()) && !arg.data().isAnyType() {
			er := "parameter \"" + param.data().print()
			er += "\" does not match argument[" + strconv.Itoa(ix) + "] \"" + arg.data().print() + "\" of function signature \"" + sig.print() + "\""
			return err(me, ECodeFunctionParameter, er)
		}
		params = append(params, param)
	}
	for _, param := range params {
		n.push(param)
	}
	return nil
}

func (me *parser) pushFunctionParams(n *node, params []*node, fn *function) *parseError {
	for ix, param := range params {
		if param == nil {
			var arg *funcArg
			if ix < len(fn.args) {
				arg = fn.args[ix]
			} else {
				arg = fn.argVariadic
			}
			d := arg.defaultNode
			if d == nil {
				e := fmt.Sprintf("I did not find any parameter for index `%d` of `%s` function call", ix, fn.getname())
				return err(
					me, ECodeUnexpectedToken, e)
			}
			n.push(d)
		} else {
			n.push(param)
		}
	}
	return nil
}

func (me *parser) functionParams(name string, pix int, params []*node, fn *function, lazy bool) (*function, []*node, *parseError) {
	if er := me.eat("("); er != nil {
		return nil, nil, er
	}
	if me.isNewLine() {
		me.newLine()
	}
	min := pix
	dict := false
	size := len(fn.args)
	gtypes := make(map[string]*datatype)
	for {
		if me.token.is == ")" {
			break
		} else if pix > min || dict {
			if me.isNewLine() {
				me.newLine()
				if me.token.is == ")" {
					break
				}
			} else {
				if er := me.eat(","); er != nil {
					return nil, nil, er
				}
			}
		}
		if me.token.is == "id" && me.peek().is == ":" {
			argname := me.token.value
			if er := me.eat("id"); er != nil {
				return nil, nil, er
			}
			if er := me.eat(":"); er != nil {
				return nil, nil, er
			}
			aix, arg := getParameter(fn.args, argname)
			if me.token.is == "_" {
				if er := me.eat("_"); er != nil {
					return nil, nil, er
				}
				if arg.defaultNode == nil {
					d, er := me.defaultValue(arg.data(), fn.getname())
					if er != nil {
						return nil, nil, er
					}
					params[aix] = d
				} else {
					params[aix] = nil
				}
			} else {
				param, er := me.calc(0, nil)
				if er != nil {
					return nil, nil, er
				}

				var update map[string]*datatype
				if len(fn.generics) > 0 {
					update = me.hintGeneric(param.data(), arg.data(), fn.generics)
				}

				if update != nil && len(update) > 0 {
					lazy = true
					good, newtypes := mergeMaps(update, gtypes)
					if !good {
						a := genericsmap(gtypes)
						b := genericsmap(update)
						f := fmt.Sprint("Lazy generic for function '"+fn.getname()+"' is ", a, " but found ", b)
						return nil, nil, err(me, ECodeFunctionLazyParameter, f)
					}
					gtypes = newtypes

				} else if param.data().notEquals(arg.data()) && !arg.data().isAnyType() {
					er := "parameter \"" + param.data().print()
					er += "\" does not match argument \"" + argname + "\" typed \"" + arg.data().print() + "\" for function \"" + name + "\""
					return nil, nil, err(me, ECodeFunctionParameter, er)
				}
				params[aix] = param
			}
			dict = true

		} else if dict {
			return nil, nil, err(me, ECodeFunctionMixedParameters, "regular paramater found after mapped parameter")
		} else {
			var arg *funcArg
			if pix >= size {
				if fn.argVariadic != nil {
					arg = fn.argVariadic
					params = append(params, nil)
				} else {
					return nil, nil, err(me, ECodeFunctionTooManyParameters, "function \""+name+"\" argument count exceeds parameter count")
				}
			}
			if arg == nil {
				arg = fn.args[pix]
			}
			if me.token.is == "_" {
				if er := me.eat("_"); er != nil {
					return nil, nil, er
				}
				if arg.defaultNode == nil {
					d, er := me.defaultValue(arg.data(), fn.getname())
					if er != nil {
						return nil, nil, er
					}
					params[pix] = d
				} else {
					params[pix] = nil
				}
			} else {
				param, er := me.calc(0, nil)
				if er != nil {
					return nil, nil, er
				}
				var update map[string]*datatype
				if len(fn.generics) > 0 {
					update = me.hintGeneric(param.data(), arg.data(), fn.generics)
				}

				if update != nil && len(update) > 0 {
					lazy = true
					good, newtypes := mergeMaps(update, gtypes)
					if !good {
						a := genericsmap(gtypes)
						b := genericsmap(update)
						f := fmt.Sprint("Lazy generic for function '"+fn.getname()+"' is ", a, " but found ", b)
						return nil, nil, err(me, ECodeFunctionLazyParameter, f)
					}
					gtypes = newtypes

				} else if param.data().notEquals(arg.data()) && !arg.data().isAnyType() {
					er := fmt.Sprintf("Function `%s` expects `%s` but found `%s`", fn.canonical(me.hmfile), arg.data().print(), param.data().print())
					return nil, nil, err(me, ECodeFunctionParameter, er)
				}
				params[pix] = param
			}
			pix++
		}
	}
	if er := me.eat(")"); er != nil {
		return nil, nil, er
	}
	if lazy {
		module := me.hmfile
		glist := make([]*datatype, len(gtypes))
		for k, v := range gtypes {
			i := inList(fn.generics, k)
			glist[i] = v.copy()
		}
		if len(glist) != len(fn.generics) {
			f := fmt.Sprint("Missing generic for function '"+fn.getname()+"'\nImplementation list was ", genericslist(glist))
			return nil, nil, err(me, ECodeFunctionMissingGeneric, f)
		}
		lazy := name + genericslist(glist)
		if implementation, ok := module.functions[lazy]; ok {
			fn = implementation
		} else {
			var er *parseError
			fn, er = remapFunctionImpl(lazy, gtypes, fn)
			if er != nil {
				return nil, nil, er
			}
		}
	}
	return fn, params, nil
}

func (me *parser) callClassFunction(module *hmfile, root *node, c *class, fn *function) (*node, *parseError) {
	params := make([]*node, len(fn.args))
	params[0] = root
	var er *parseError
	_, params, er = me.functionParams(fn.getclsname(), 1, params, fn, false)
	if er != nil {
		return nil, er
	}
	n := nodeInit("call")
	n.fn = fn
	n.copyData(fn.returns)
	if er = me.pushFunctionParams(n, params, fn); er != nil {
		return nil, er
	}
	return n, nil
}

func (me *parser) call(module *hmfile) (*node, *parseError) {
	name := me.token.value
	if er := me.eat("id"); er != nil {
		return nil, er
	}
	var order []*datatype
	bfn, ok := module.getFunction(name)
	if !ok {
		return nil, err(me, ECodeFunctionNotFound, "Missing function '"+name+"'")
	}
	fn := bfn
	lazy := false
	if bfn.generics != nil {
		if me.token.is == "<" {
			var er *parseError
			order, _, er = me.genericHeader()
			if er != nil {
				return nil, er
			}
			name += genericslist(order)
			gfn, ok := module.getFunction(name)
			if ok {
				fn = gfn
			} else {
				mapping := make(map[string]*datatype)
				for i, g := range bfn.generics {
					mapping[g] = order[i]
				}
				var er *parseError
				fn, er = remapFunctionImpl(name, mapping, bfn)
				if er != nil {
					return nil, er
				}
			}
		} else {
			lazy = true
		}
	}
	params := make([]*node, len(fn.args))
	var er *parseError
	fn, params, er = me.functionParams(name, 0, params, fn, lazy)
	if er != nil {
		return nil, er
	}
	n := nodeInit("call")
	n.fn = fn
	n.copyData(fn.returns)
	if er = me.pushFunctionParams(n, params, fn); er != nil {
		return nil, er
	}
	return n, er
}

func (me *parser) parseFn(module *hmfile) (*node, *parseError) {
	if me.peek().is == "(" || me.peek().is == "<" {
		return me.call(module)
	}

	name := me.token.value
	fn := module.functions[name]
	if er := me.eat("id"); er != nil {
		return nil, er
	}
	n := nodeInit("fn-ptr")
	newdata, er := fn.data()
	if er != nil {
		return nil, er
	}
	n.copyData(newdata)
	n.fn = fn
	return n, nil
}
