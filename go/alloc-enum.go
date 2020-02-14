package main

import (
	"fmt"
	"strconv"
)

func (me *parser) pushEnumParams(n *node, un *union, params []*node, typed string) {
	for i, param := range params {
		if param == nil {
			v := un.types.get(i)
			d := me.defaultValue(v, typed)
			n.push(d)
		} else {
			n.push(param)
		}
	}
}

func (me *parser) enumParams(n *node, en *enum, un *union, depth int) string {
	if me.token.is != "(" {
		panic(me.fail() + "Enum: " + n.data().print() + " requires parameters")
	}
	me.eat("(")
	if me.token.is == "line" {
		me.eat("line")
	}
	vars := un.types.order
	params := make([]*node, len(vars))
	pix := 0
	dict := false
	lazy := false
	gtypes := make(map[string]*datatype)
	gindex := en.genericsDict
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
					panic(me.fail() + "Unexpected line indentation")
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
			unvar, ok := un.types.table[vname]
			if !ok {
				panic(me.fail() + "Member variable: " + vname + " does not exist for enum: " + en.join(un))
			}

			var update map[string]*datatype
			if len(gindex) > 0 {
				update = me.hintGeneric(param.data(), unvar, gindex)
			}

			if update != nil && len(update) > 0 {
				lazy = true
				good, newtypes := mergeMaps(update, gtypes)
				if !good {
					a := genericsmap(gtypes)
					b := genericsmap(update)
					f := fmt.Sprint("Lazy generic for enum: " + en.join(un) + " is " + a + " but found " + b)
					panic(me.fail() + f)
				}
				gtypes = newtypes

			} else if param.data().notEquals(unvar) && !unvar.isQuestion() {
				err := "Parameter: " + vname + " with type \"" + param.data().print()
				err += "\" does not match class variable \"" + en.join(un) + "."
				err += vname + "\" with type \"" + unvar.print() + "\""
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
			panic(me.fail() + "Regular paramater found after mapped parameter")
		} else {
			unvar := un.types.table[vars[pix]]
			if me.token.is == "_" {
				me.eat("_")
				params[pix] = nil
			} else {
				param := me.calc(0, nil)

				var update map[string]*datatype
				if len(gindex) > 0 {
					update = me.hintGeneric(param.data(), unvar, gindex)
				}

				if update != nil && len(update) > 0 {
					lazy = true
					good, newtypes := mergeMaps(update, gtypes)
					if !good {
						f := fmt.Sprint("Lazy generic for enum: "+en.join(un)+" is ", gtypes, " but found ", update)
						panic(me.fail() + f)
					}
					gtypes = newtypes

				} else if param.data().notEquals(unvar) && !unvar.isQuestion() {
					err := "Parameter " + strconv.Itoa(pix) + " with type: " + param.data().print()
					err += " does not match enum variable: " + en.join(un) + "."
					err += strconv.Itoa(pix) + "\" with type: " + unvar.print()
					panic(me.fail() + err)
				}
				params[pix] = param
			}
			pix++
		}
	}
	module := en.module
	if lazy {
		glist := make([]*datatype, len(gtypes))
		for k, v := range gtypes {
			i, _ := gindex[k]
			glist[i] = v.copy()
		}
		if len(glist) != len(en.generics) {
			f := fmt.Sprint("Missing generic for enum: " + en.join(un) + "\"\nImplementation list was " + genericslist(glist))
			panic(me.fail() + f)
		}
		typed := en.name + genericslist(glist)
		if _, ok := module.enums[typed]; !ok {
			me.defineEnumImplGeneric(en, glist)
		}
		en = module.enums[typed]
	}
	typed := en.join(un)
	me.pushEnumParams(n, un, params, typed)
	return typed

	// assign := me.hmfile.peekAssignStack()
	// var assignEn *enum
	// if assign != nil && !assign.isQuestion() {
	// 	assignEn = assign.enum
	// }

	// gdict := en.genericsDict

	// gimpl := make(map[string]string)

	// for ix, unionKey := range un.types.order {
	// 	unionType := un.types.table[unionKey]
	// 	if ix != 0 {
	// 		if me.token.is != "," {
	// 			panic(me.fail() + "Expecting \"" + unionType.print() + "\" for enum \"" + typed + "\".")
	// 		}
	// 		me.eat(",")
	// 	}
	// 	param := me.calc(0, nil)
	// 	if param.data().notEquals(unionType) {
	// 		if _, gok := gdict[unionType.getRaw()]; gok {
	// 			gimpl[unionType.getRaw()] = param.data().getRaw()
	// 		} else {
	// 			panic(me.fail() + "Enum: " + enumName + "." + unionName + " expects: " + unionType.print() + " but parameter was: " + param.data().print())
	// 		}
	// 	}
	// 	n.push(param)
	// }
	// me.eat(")")
	// if len(order) == 0 {
	// 	if len(gimpl) != len(gdict) {
	// 		if assignEn != nil {
	// 			for k, v := range assignEn.gmapper {
	// 				if nv, ok := gimpl[k]; ok {
	// 					if nv != v {
	// 						panic(me.fail() + "\"" + enumName + "\" with \"" + nv + "\" does not match \"" + v + "\"")
	// 					}
	// 				} else {
	// 					gimpl[k] = v
	// 				}
	// 			}
	// 		} else {
	// 			panic(me.fail() + "Enum: " + enumName + " with implementation: " + fmt.Sprint(gimpl) + " does not match: " + fmt.Sprint(gdict))
	// 		}
	// 	}
	// 	if len(gimpl) > 0 {
	// 		order := me.mapUnionGenerics(en, gimpl)
	// 		enumName += genericslist(order)
	// 		if _, ok := module.enums[enumName]; !ok {
	// 			me.defineEnumImplGeneric(enumDef, order)
	// 		}
	// 	}
	// }
}

func (me *parser) buildEnum(n *node, module *hmfile) *datatype {
	typed := me.token.value
	depth := me.token.depth
	me.eat("id")
	en, ok := module.enums[typed]
	if !ok {
		panic(me.fail() + "Enum: " + typed + " does not exist")
	}
	uid := module.reference(typed)
	gsize := len(en.generics)
	if gsize > 0 {
		if me.token.is == "<" {
			gtypes := me.declareGeneric(en)
			if len(gtypes) != len(en.generics) {
				panic(me.fail() + "Enum \"" + en.name + " with implementation " + fmt.Sprint(gtypes) + " does not match " + fmt.Sprint(en.generics))
			}
			typed = uid + genericslist(gtypes)
			if _, ok := module.enums[typed]; !ok {
				me.defineEnumImplGeneric(en, gtypes)
			}
			en = module.enums[typed]
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
				en = d.enum
			}
		}
	}

	me.eat(".")
	unvalue := me.token.value
	me.eat("id")
	un, ok := en.types[unvalue]
	if !ok {
		panic(me.fail() + "Enum: " + en.name + " does not have type: " + unvalue)
	}
	if n != nil && !en.simple {
		typed = me.enumParams(n, en, un, depth)
	} else {
		typed = en.join(un)
	}
	if me.hmfile != module {
		typed = module.reference(typed)
	}
	data := getdatatype(module, typed)
	fmt.Println("alloc enum || module:", module.name, "typed:", typed, "enum:", en.name, "union:", un.name, "data:", data.error())
	return data
}

func (me *parser) allocEnum(module *hmfile, hint *allocHint) *node {
	n := nodeInit("enum")
	data := me.buildEnum(n, module)
	data = data.merge(hint)
	n.copyData(data)
	if hint != nil && hint.stack {
		n.attributes["stack"] = "true"
		n.data().setIsPointer(false)
		n.data().setIsOnStack(true)
	}
	return n
}
