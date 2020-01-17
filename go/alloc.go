package main

import (
	"fmt"
	"strconv"
	"strings"
)

func (me *parser) allocEnum(module *hmfile) *node {
	enumName := me.token.value
	me.eat("id")
	enumDef, ok := module.enums[enumName]
	if !ok {
		panic(me.fail() + "enum \"" + enumName + "\" does not exist")
	}

	gdict := enumDef.genericsDict
	var order []string
	if me.token.is == "<" {
		order, _ = me.genericHeader()
		enumName += "<" + strings.Join(order, ",") + ">"
		if len(order) != len(gdict) {
			panic(me.fail() + "generic enum \"" + enumName + "\" with impl " + fmt.Sprint(order) + " does not match " + fmt.Sprint(gdict))
		}
		if _, ok := module.enums[enumName]; !ok {
			me.defineEnumImplGeneric(enumDef, enumName, order)
		}
	}

	me.eat(".")
	unionName := me.token.value
	me.eat("id")
	unionDef, ok := enumDef.types[unionName]
	if !ok {
		panic(me.fail() + "enum \"" + enumName + "\" does not have type \"" + unionName + "\"")
	}

	n := nodeInit("enum")

	typeSize := len(unionDef.types)
	if typeSize > 0 {
		if me.token.is != "(" {
			panic(me.fail() + "enum \"" + n.data().print() + "\" requires parameters")
		}
		me.eat("(")
		gimpl := make(map[string]string)
		for ix, unionType := range unionDef.types {
			if ix != 0 {
				me.eat(",")
			}
			param := me.calc(0)
			if param.data().notEquals(unionType) {
				if _, gok := gdict[unionType.getRaw()]; gok {
					gimpl[unionType.getRaw()] = param.data().getRaw()
				} else {
					panic(me.fail() + "enum \"" + enumName + "\" type \"" + unionName + "\" expects \"" + unionType.print() + "\" but parameter was \"" + param.data().print() + "\"")
				}
			}
			n.push(param)
		}
		me.eat(")")
		if len(order) == 0 {
			if len(gimpl) != len(gdict) {
				panic(me.fail() + "generic enum \"" + enumName + "\" with impl " + fmt.Sprint(gimpl) + " does not match " + fmt.Sprint(gdict))
			}
			if len(gimpl) > 0 {
				order = me.mapUnionGenerics(enumDef, gimpl)
				enumName += "<" + strings.Join(order, ",") + ">"
				if _, ok := module.enums[enumName]; !ok {
					me.defineEnumImplGeneric(enumDef, enumName, order)
				}
			}
		}
	} else if len(gdict) != 0 && len(order) == 0 {
		panic(me.fail() + "generic enum \"" + enumName + "\" has no impl for " + fmt.Sprint(enumDef.generics))
	}

	n.copyData(typeToVarData(module, enumName+"."+unionName))

	return n
}

func (me *parser) pushAllDefaultClassParams(n *node) {
	base, ok := n.data().isClass()
	if !ok {
		panic(me.fail())
	}
	vars := base.variableOrder
	params := make([]*node, len(vars))
	me.pushClassParams(n, base, params)
}

func (me *parser) defaultValue(in *datatype) *node {
	d := nodeInit(in.getRaw())
	d.copyData(in)
	data := in
	if data.isString() {
		d.value = ""
	} else if data.isChar() {
		d.value = "\\0"
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
		s.copyData(typeToVarData(me.hmfile, TokenInt))
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
		panic(me.fail() + "no default value for \"" + d.is + "\"")
	}
	return d
}

func (me *parser) pushClassParams(n *node, base *class, params []*node) {
	for i, param := range params {
		if param == nil {
			clsvar := base.variables[base.variableOrder[i]]
			d := me.defaultValue(clsvar.data())
			n.push(d)
		} else {
			n.push(param)
		}
	}
}

func (me *parser) classParams(n *node, module *hmfile, typed string, depth int) string {
	me.eat("(")
	if me.token.is == "line" {
		me.eat("line")
	}
	base := module.classes[typed]
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
			param := me.calc(0)
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
				param := me.calc(0)

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
	if lazyGenerics {
		glist := make([]string, len(gtypes))
		for k, v := range gtypes {
			i, _ := gindex[k]
			glist[i] = v.print()
		}
		if len(glist) != len(base.generics) {
			f := fmt.Sprint("missing generic for base class \""+base.name+"\"\nimplementation list was ", glist)
			panic(me.fail() + f)
		}
		lazy := typed + "<" + strings.Join(glist, ",") + ">"
		if _, ok := module.classes[lazy]; !ok {
			me.defineClassImplGeneric(base, lazy, glist)
		}
		base = module.classes[lazy]
		typed = lazy
	}
	me.pushClassParams(n, base, params)
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
	if gsize > 0 {
		if me.token.is == "<" {
			gtypes := me.declareGeneric(true, base)
			typed = name + "<" + strings.Join(gtypes, ",") + ">"
			if _, ok := me.hmfile.classes[typed]; !ok {
				me.defineClassImplGeneric(base, typed, gtypes)
			}
		} else {
			assign := me.hmfile.assignmentStack[len(me.hmfile.assignmentStack)-1].data()
			if !assign.isQuestion() {
				if assign.isSome() {
					typed = assign.getmember().getRaw()
				} else if assign.isArrayOrSlice() {
					typed = assign.getmember().getRaw()
				} else {
					typed = assign.getRaw()
				}
			}
		}
	}
	if n != nil {
		typed = me.classParams(n, module, typed, depth)
	}
	if me.hmfile != module {
		typed = module.name + "." + typed
	}
	return typeToVarData(me.hmfile, typed)
}

func (me *parser) allocClass(module *hmfile, alloc *allocData) *node {
	n := nodeInit("new")
	data := me.buildClass(n, module)
	data = data.merge(alloc)
	n.copyData(data)
	if alloc != nil && alloc.stack {
		n.attributes["stack"] = "true"
		n.data().setOnStackNotPointer()
	}
	return n
}
