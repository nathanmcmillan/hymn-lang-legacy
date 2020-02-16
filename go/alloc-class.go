package main

import (
	"fmt"
	"strconv"
)

func (me *parser) pushAllDefaultClassParams(n *node) {
	cl, ok := n.data().isClass()
	if !ok {
		panic(me.fail())
	}
	vars := cl.variableOrder
	params := make([]*node, len(vars))
	me.pushClassParams(n, cl, params, cl.name)
}

func (me *parser) pushClassParams(n *node, classDef *class, params []*node, typed string) {
	for i, param := range params {
		if param == nil {
			clsvar := classDef.variables[classDef.variableOrder[i]]
			d := me.defaultValue(clsvar.data(), typed)
			n.push(d)
		} else {
			n.push(param)
		}
	}
}

func (me *parser) classParams(n *node, cl *class, depth int) string {
	me.eat("(")
	if me.token.is == "line" {
		me.eat("line")
	}
	vars := cl.variableOrder
	params := make([]*node, len(vars))
	pix := 0
	dict := false
	lazy := false
	gtypes := make(map[string]*datatype)
	gindex := cl.genericsDict
	for {
		if me.token.is == ")" {
			me.eat(")")
			break
		}
		if pix > 0 || dict {
			if me.token.is == "line" {
				ndepth := me.peek().depth
				if ndepth == depth && me.peek().is == ")" {
					me.eat("line")
					me.eat(")")
					break
				}
				if ndepth != depth+1 {
					panic(me.fail() + "unexpected line indentation")
				}
				me.eat("line")
			} else {
				me.eat(",")
			}
		}
		if me.token.is == "id" && me.peek().is == ":" {
			dict = true

			vname := me.token.value
			me.eat("id")
			clsvar, ok := cl.variables[vname]
			if !ok {
				panic(me.fail() + "Member variable: " + vname + " does not exist for class: " + cl.name)
			}

			me.eat(":")

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
					lazy = true
					good, newtypes := mergeMaps(update, gtypes)
					if !good {
						a := genericsmap(gtypes)
						b := genericsmap(update)
						f := fmt.Sprint("Lazy generic for class \""+cl.name+"\" is ", a, " but found ", b)
						panic(me.fail() + f)
					}
					gtypes = newtypes

				} else if param.data().notEquals(clsvar.data()) && !clsvar.data().isQuestion() {
					err := "parameter \"" + vname + "\" with type \"" + param.data().print()
					err += "\" does not match class variable \"" + cl.name + "."
					err += clsvar.name + "\" with type \"" + clsvar.data().print() + "\""
					panic(me.fail() + err)
				}
				for i, v := range vars {
					if vname == v {
						params[i] = param
						break
					}
				}
			}

		} else if dict {
			panic(me.fail() + "Regular paramater found after mapped parameter")
		} else {
			clsvar := cl.variables[vars[pix]]
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
					lazy = true
					good, newtypes := mergeMaps(update, gtypes)
					if !good {
						f := fmt.Sprint("lazy generic for class \""+cl.name+"\" is ", gtypes, " but found ", update)
						panic(me.fail() + f)
					}
					gtypes = newtypes

				} else if param.data().notEquals(clsvar.data()) && !clsvar.data().isQuestion() {
					err := "parameter " + strconv.Itoa(pix) + " with type \"" + param.data().print()
					err += "\" does not match class variable \"" + cl.name + "."
					err += clsvar.name + "\" with type \"" + clsvar.data().print() + "\""
					panic(me.fail() + err)
				}
				params[pix] = param
			}
			pix++
		}
	}
	module := cl.module
	typed := cl.name
	if lazy {
		glist := make([]*datatype, len(gtypes))
		for k, v := range gtypes {
			i, _ := gindex[k]
			glist[i] = v.copy()
		}
		if len(glist) != len(cl.generics) {
			f := fmt.Sprint("Missing generic for class \""+cl.name+"\"\nimplementation list was ", genericslist(glist))
			panic(me.fail() + f)
		}
		lazy := typed + genericslist(glist)
		if _, ok := module.classes[lazy]; !ok {
			me.defineClassImplGeneric(cl, glist)
		}
		cl = module.classes[lazy]
		typed = lazy
	}
	me.pushClassParams(n, cl, params, typed)
	return typed
}

func (me *parser) buildClass(n *node, module *hmfile) *datatype {
	typed := me.token.value
	depth := me.token.depth
	me.eat("id")
	cl, ok := module.classes[typed]
	if !ok {
		panic(me.fail() + "Class: " + typed + " does not exist")
	}
	uid := module.reference(typed)
	gsize := len(cl.generics)
	if gsize > 0 {
		if me.token.is == "<" {
			gtypes := me.declareGeneric(cl)
			if len(gtypes) != gsize {
				panic(me.fail() + "Class:" + cl.name + " with implementation " + fmt.Sprint(gtypes) + " does not match " + fmt.Sprint(cl.generics))
			}
			typed = uid + genericslist(gtypes)
			if _, ok := me.hmfile.classes[typed]; !ok {
				me.defineClassImplGeneric(cl, gtypes)
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
	if me.hmfile != module {
		typed = module.reference(typed)
	}
	return getdatatype(me.hmfile, typed)
}

func (me *parser) allocClass(module *hmfile, hint *allocHint) *node {
	n := nodeInit("new")
	data := me.buildClass(n, module)
	data = data.merge(hint)
	n.copyData(data)
	if hint != nil && hint.stack {
		n.attributes["stack"] = "true"
		n.data().setIsPointer(false)
		n.data().setIsOnStack(true)
	}
	return n
}
