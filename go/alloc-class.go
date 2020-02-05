package main

import (
	"fmt"
	"strconv"
)

func (me *parser) pushAllDefaultClassParams(n *node) {
	base, ok := n.data().isClass()
	if !ok {
		panic(me.fail())
	}
	vars := base.variableOrder
	params := make([]*node, len(vars))
	me.pushClassParams(n, base, params, base.name)
}

func (me *parser) defaultValue(data *datatype, from string) *node {
	d := nodeInit(data.getRaw())
	d.copyData(data)
	if data.isString() {
		d.value = ""
	} else if data.isChar() {
		d.value = "'\\0'"
	} else if data.isNumber() {
		d.value = "0"
	} else if data.isBoolean() {
		d.value = "false"
	} else if data.isArray() {
		t := nodeInit("array")
		t.copyData(d.data())
		d = t
	} else if data.isSlice() {
		t := nodeInit("slice")
		t.copyData(d.data())
		s := nodeInit(TokenInt)
		s.copyData(getdatatype(me.hmfile, TokenInt))
		s.value = "0"
		t.push(s)
		d = t
	} else if _, ok := data.isClass(); ok {
		t := nodeInit("new")
		t.copyData(d.data())
		me.pushAllDefaultClassParams(t)
		d = t
	} else if data.isSomeOrNone() {
		t := nodeInit("none")
		t.copyData(d.data())
		t.value = "NULL"
		d = t
	} else {
		e := me.fail()
		if from != "" {
			e += "\nFrom: " + from
		}
		e += "\nType: " + d.is + "\nProblem: No default value available."
		panic(e)
	}
	return d
}

func (me *parser) pushClassParams(n *node, base *class, params []*node, typed string) {
	for i, param := range params {
		if param == nil {
			clsvar := base.variables[base.variableOrder[i]]
			d := me.defaultValue(clsvar.data(), typed)
			n.push(d)
		} else {
			n.push(param)
		}
	}
}

func (me *parser) classParams(n *node, base *class, depth int) string {
	me.eat("(")
	if me.token.is == "line" {
		me.eat("line")
	}
	vars := base.variableOrder
	params := make([]*node, len(vars))
	pix := 0
	dict := false
	lazyGenerics := false
	gtypes := make(map[string]*datatype)
	gindex := base.genericsDict
	for {
		if me.token.is == ")" {
			me.eat(")")
			break
		}
		if pix > 0 || dict {
			if me.token.is == "line" {
				ndepth := me.peek().depth
				if ndepth != depth+1 {
					panic(me.fail() + "unexpected line indentation")
				}
				me.eat("line")
			} else {
				me.eat(",")
			}
		}
		if me.token.is == "id" && me.peek().is == ":" {
			vname := me.token.value
			me.eat("id")
			me.eat(":")
			param := me.calc(0, nil)
			clsvar, ok := base.variables[vname]
			if !ok {
				panic(me.fail() + "member variable \"" + vname + "\" does not exist for class \"" + base.name + "\"")
			}

			var update map[string]*datatype
			if len(gindex) > 0 {
				update = me.hintGeneric(param.data(), clsvar.data(), gindex)
			}

			if update != nil && len(update) > 0 {
				lazyGenerics = true
				good, newtypes := mergeMaps(update, gtypes)
				if !good {
					a := ""
					for k, v := range gtypes {
						if a != "" {
							a += ", "
						}
						a += k + ":" + v.print()
					}
					a = "map[" + a + "]"
					b := ""
					for k, v := range update {
						if b != "" {
							b += ", "
						}
						b += k + ":" + v.print()
					}
					b = "map[" + b + "]"
					f := fmt.Sprint("lazy generic for class \""+base.name+"\" is ", a, " but found ", b)
					panic(me.fail() + f)
				}
				gtypes = newtypes

			} else if param.data().notEquals(clsvar.data()) && !clsvar.data().isQuestion() {
				err := "parameter \"" + vname + "\" with type \"" + param.data().print()
				err += "\" does not match class variable \"" + base.name + "."
				err += clsvar.name + "\" with type \"" + clsvar.data().print() + "\""
				panic(me.fail() + err)
			}
			for i, v := range vars {
				if vname == v {
					params[i] = param
					break
				}
			}
			dict = true

		} else if dict {
			panic(me.fail() + "regular paramater found after mapped parameter")
		} else {
			clsvar := base.variables[vars[pix]]
			if me.token.is == "_" {
				me.eat("_")
				params[pix] = nil
			} else {
				param := me.calc(0, nil)

				var update map[string]*datatype
				if len(gindex) > 0 {
					update = me.hintGeneric(param.data(), clsvar.data(), gindex)
				}

				if update != nil && len(update) > 0 {
					lazyGenerics = true
					good, newtypes := mergeMaps(update, gtypes)
					if !good {
						f := fmt.Sprint("lazy generic for class \""+base.name+"\" is ", gtypes, " but found ", update)
						panic(me.fail() + f)
					}
					gtypes = newtypes

				} else if param.data().notEquals(clsvar.data()) && !clsvar.data().isQuestion() {
					err := "parameter " + strconv.Itoa(pix) + " with type \"" + param.data().print()
					err += "\" does not match class variable \"" + base.name + "."
					err += clsvar.name + "\" with type \"" + clsvar.data().print() + "\""
					panic(me.fail() + err)
				}
				params[pix] = param
			}
			pix++
		}
	}
	module := base.module
	typed := base.name
	if lazyGenerics {
		glist := make([]*datatype, len(gtypes))
		gprint := genericslist(glist)
		if len(glist) != len(base.generics) {
			f := fmt.Sprint("missing generic for base class \""+base.name+"\"\nimplementation list was ", gprint)
			panic(me.fail() + f)
		}
		lazy := typed + gprint
		if _, ok := module.classes[lazy]; !ok {
			me.defineClassImplGeneric(base, lazy, glist)
		}
		base = module.classes[lazy]
		typed = lazy
	}
	me.pushClassParams(n, base, params, typed)
	return typed
}

func (me *parser) buildClass(n *node, module *hmfile) *datatype {
	name := me.token.value
	depth := me.token.depth
	me.eat("id")
	base, ok := module.classes[name]
	if !ok {
		panic(me.fail() + "class \"" + name + "\" does not exist")
	}
	typed := name
	gsize := len(base.generics)
	cl := module.classes[typed]
	if gsize > 0 {
		if me.token.is == "<" {
			gtypes := me.declareGeneric(true, base)
			typed = name + genericslist(gtypes)
			if _, ok := me.hmfile.classes[typed]; !ok {
				me.defineClassImplGeneric(base, typed, gtypes)
			}
			cl = module.classes[typed]
		} else {
			assign := me.hmfile.peekAssignStack()
			if assign != nil && !assign.isQuestion() {
				var d *datatype
				if assign.isSome() || assign.isArrayOrSlice() {
					d = assign.getmember()
				} else {
					d = assign
				}
				typed = d.getRaw()
				module = d.getmodule()
				cl = d.class
			}
		}
	}
	if n != nil {
		typed = me.classParams(n, cl, depth)
	}
	// TODO: UID
	if me.hmfile != module {
		typed = module.cross(me.hmfile) + "." + typed
	}
	// typed = module.uid + "." + typed
	return getdatatype(me.hmfile, typed)
}

func (me *parser) allocClass(module *hmfile, alloc *allocData) *node {
	n := nodeInit("new")
	data := me.buildClass(n, module)
	data = data.merge(alloc)
	n.copyData(data)
	if alloc != nil && alloc.stack {
		n.attributes["stack"] = "true"
		n.data().setIsPointer(false)
		n.data().setIsOnStack(true)
	}
	return n
}
