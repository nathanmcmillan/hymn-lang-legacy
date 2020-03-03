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
	vars := cl.variables
	params := make([]*node, len(vars))
	me.pushClassParams(n, cl, params, cl.name)
}

func (me *parser) pushClassParams(n *node, classDef *class, params []*node, typed string) *parseError {
	for i, param := range params {
		if param == nil {
			clsvar := classDef.variables[i]
			d, er := me.defaultValue(clsvar.data(), typed)
			if er != nil {
				return er
			}
			n.push(d)
		} else {
			n.push(param)
		}
	}
	return nil
}

func (me *parser) classParams(n *node, cl *class, depth int) (string, *parseError) {
	if er := me.eat("("); er != nil {
		return "", er
	}
	if me.token.is == "line" {
		me.eat("line")
	}
	params := make([]*node, len(cl.variables))
	pix := 0
	dict := false
	lazy := false
	gtypes := make(map[string]*datatype)
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
					return "", erc(me, ECodeLineIndentation)
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
			clsvar := cl.getVariable(vname)
			if clsvar == nil {
				return "", err(me, ECodeClassMemberNotFound, "Member variable: "+vname+" does not exist for class: "+cl.name)
			}

			me.eat(":")
			var param *node

			if me.token.is == "_" {
				me.eat("_")
			} else {
				var er *parseError
				param, er = me.calc(0, nil)
				if er != nil {
					return "", er
				}

				var update map[string]*datatype
				if len(cl.generics) > 0 {
					update = me.hintGeneric(param.data(), clsvar.data(), cl.generics)
				}

				if update != nil && len(update) > 0 {
					lazy = true
					good, newtypes := mergeMaps(update, gtypes)
					if !good {
						a := genericsmap(gtypes)
						b := genericsmap(update)
						f := fmt.Sprint("Lazy generic for class '"+cl.name+"' is ", a, " but found ", b)
						return "", err(me, ECodeClassLazyParameter, f)
					}
					gtypes = newtypes

				} else if param.data().notEquals(clsvar.data()) && !clsvar.data().isAnyType() {
					er := "parameter \"" + vname + "\" with type \"" + param.data().print()
					er += "\" does not match class variable \"" + cl.name + "."
					er += clsvar.name + "\" with type \"" + clsvar.data().print() + "\""
					return "", err(me, ECodeClassParameter, er)
				}
			}

			for i, v := range cl.variables {
				if vname == v.name {
					params[i] = param
					break
				}
			}

		} else if dict {
			return "", err(me, ECodeMixedParameters, "Regular paramater found after mapped parameter")
		} else {
			clsvar := cl.variables[pix]
			if me.token.is == "_" {
				me.eat("_")
				params[pix] = nil
			} else {
				param, er := me.calc(0, nil)
				if er != nil {
					return "", er
				}

				var update map[string]*datatype
				if len(cl.generics) > 0 {
					update = me.hintGeneric(param.data(), clsvar.data(), cl.generics)
				}

				if update != nil && len(update) > 0 {
					lazy = true
					good, newtypes := mergeMaps(update, gtypes)
					if !good {
						f := fmt.Sprint("Lazy generic for class '"+cl.name+"' is ", gtypes, " but found ", update)
						return "", err(me, ECodeClassLazyParameter, f)
					}
					gtypes = newtypes

				} else if param.data().notEquals(clsvar.data()) && !clsvar.data().isAnyType() {
					er := "parameter " + strconv.Itoa(pix) + " with type \"" + param.data().print()
					er += "\" does not match class variable \"" + cl.name + "."
					er += clsvar.name + "\" with type \"" + clsvar.data().print() + "\""
					return "", err(me, ECodeClassParameter, er)
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
			i := inList(cl.generics, k)
			glist[i] = v.copy()
		}
		if len(glist) != len(cl.generics) {
			f := fmt.Sprint("Missing generic for class \""+cl.name+"\"\nimplementation list was ", genericslist(glist))
			return "", err(me, ECodeClassMissingGeneric, f)
		}
		lazy := typed + genericslist(glist)
		if _, ok := module.classes[lazy]; !ok {
			me.defineClassImplGeneric(cl, glist)
		}
		cl = module.classes[lazy]
		typed = lazy
	}
	me.pushClassParams(n, cl, params, typed)
	return typed, nil
}

func (me *parser) buildClass(n *node, module *hmfile) (*datatype, *parseError) {
	typed := me.token.value
	depth := me.token.depth
	me.eat("id")
	cl, ok := module.classes[typed]
	if !ok {
		return nil, err(me, ECodeClassDoesNotExist, "Class: "+typed+" does not exist")
	}
	uid := module.reference(typed)
	gsize := len(cl.generics)
	if gsize > 0 {
		if me.token.is == "<" {
			gtypes, er := me.declareGeneric(len(cl.generics))
			if er != nil {
				return nil, er
			}
			if len(gtypes) != gsize {
				return nil, err(me, ECodeClassImplementationMismatch, "Class:"+cl.name+" with implementation "+fmt.Sprint(gtypes)+" does not match "+fmt.Sprint(cl.generics))
			}
			typed = uid + genericslist(gtypes)
			if _, ok := me.hmfile.classes[typed]; !ok {
				me.defineClassImplGeneric(cl, gtypes)
			}
			cl = module.classes[typed]
		} else {
			assign := me.hmfile.peekAssignStack()
			if assign != nil && !assign.isAnyType() {
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
		var er *parseError
		typed, er = me.classParams(n, cl, depth)
		if er != nil {
			return nil, er
		}
	}
	if me.hmfile != module {
		typed = module.reference(typed)
	}
	return getdatatype(me.hmfile, typed)
}

func (me *parser) allocClass(module *hmfile, hint *allocHint) (*node, *parseError) {
	n := nodeInit("new")
	data, er := me.buildClass(n, module)
	if er != nil {
		return nil, er
	}
	data = data.merge(hint)
	n.copyData(data)
	if hint != nil && hint.stack {
		n.attributes["stack"] = "true"
		n.data().setIsPointer(false)
		n.data().setIsOnStack(true)
	}
	return n, nil
}
