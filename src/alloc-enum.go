package main

import (
	"fmt"
	"strconv"
)

func (me *parser) pushEnumParams(n *node, un *union, params []*node, typed string) *parseError {
	for i, param := range params {
		if param == nil {
			v := un.types.get(i)
			d, er := me.defaultValue(v, typed)
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

func (me *parser) enumParams(n *node, en *enum, un *union, depth int) (string, *parseError) {
	if me.token.is != "(" {
		if me.token.is == "<" {
			return "", err(me, ECodeEnumBracketPosition, "Enum: "+en.join(un)+" was expects something like '"+en.name+"<>."+un.name+"()'")
		}
		return "", err(me, ECodeEnumMissingParenthesis, "Enum: "+en.join(un)+" must be instantiated with parenthesis\nExample: "+en.join(un)+"()")
	}
	if er := me.eat("("); er != nil {
		return "", er
	}
	if me.isNewLine() {
		me.newLine()
	}
	vars := un.types.order
	params := make([]*node, len(vars))
	pix := 0
	dict := false
	lazy := false
	gtypes := make(map[string]*datatype)
	for {
		if me.token.is == ")" {
			if er := me.eat(")"); er != nil {
				return "", er
			}
			break
		}
		if pix > 0 || dict {
			if me.token.is == "line" {
				ndepth := me.peek().depth
				if ndepth == depth && me.peek().is == ")" {
					if er := me.eat("line"); er != nil {
						return "", er
					}
					if er := me.eat(")"); er != nil {
						return "", er
					}
					break
				}
				if ndepth != depth+1 {
					return "", erc(me, ECodeLineIndentation)
				}
				if er := me.newLine(); er != nil {
					return "", er
				}
			} else {
				if er := me.eat(","); er != nil {
					return "", er
				}
			}
		}
		if me.token.is == "id" && me.peek().is == ":" {
			dict = true

			vname := me.token.value
			if er := me.eat("id"); er != nil {
				return "", er
			}
			unvar, ok := un.types.table[vname]
			if !ok {
				return "", err(me, ECodeEnumMemberNotFound, "Member variable: "+vname+" does not exist for enum: "+en.join(un))
			}

			if er := me.eat(":"); er != nil {
				return "", er
			}
			var param *node

			if me.token.is == "_" {
				if er := me.eat("_"); er != nil {
					return "", er
				}
			} else {
				var er *parseError
				param, er = me.calc(0, nil)
				if er != nil {
					return "", er
				}

				var update map[string]*datatype
				if len(en.generics) > 0 {
					update = me.hintGeneric(param.data(), unvar, en.generics)
				}

				if update != nil && len(update) > 0 {
					lazy = true
					good, newtypes := mergeMaps(update, gtypes)
					if !good {
						a := genericsmap(gtypes)
						b := genericsmap(update)
						f := fmt.Sprint("Lazy generic for enum: " + en.join(un) + " is " + a + " but found " + b)
						return "", err(me, ECodeEnumLazyParameter, f)
					}
					gtypes = newtypes

				} else if param.data().notEquals(unvar) && !unvar.isAnyType() {
					er := "Parameter: " + vname + " with type \"" + param.data().print()
					er += "\" does not match class variable \"" + en.join(un) + "."
					er += vname + "\" with type \"" + unvar.print() + "\""
					return "", err(me, ECodeEnumParameter, er)
				}
			}

			for i, v := range vars {
				if vname == v {
					params[i] = param
					break
				}
			}

		} else if dict {
			return "", err(me, ECodeMixedParameters, "Regular paramater found after mapped parameter")
		} else {
			unvar := un.types.table[vars[pix]]
			if me.token.is == "_" {
				if er := me.eat("_"); er != nil {
					return "", er
				}
				params[pix] = nil
			} else {
				param, er := me.calc(0, nil)
				if er != nil {
					return "", er
				}

				var update map[string]*datatype
				if len(en.generics) > 0 {
					update = me.hintGeneric(param.data(), unvar, en.generics)
				}

				if update != nil && len(update) > 0 {
					lazy = true
					good, newtypes := mergeMaps(update, gtypes)
					if !good {
						f := fmt.Sprint("Lazy generic for enum: "+en.join(un)+" is ", gtypes, " but found ", update)
						return "", err(me, ECodeEnumLazyParameter, f)
					}
					gtypes = newtypes

				} else if param.data().notEquals(unvar) && !unvar.isAnyType() {
					er := "Parameter " + strconv.Itoa(pix) + " with type: " + param.data().print()
					er += " does not match enum variable: " + en.join(un) + "."
					er += strconv.Itoa(pix) + "\" with type: " + unvar.print()
					return "", err(me, ECodeEnumParameter, er)
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
			i := inList(en.generics, k)
			if i >= len(glist) {
				return "", err(me, ECodeEnumIncompleteDeclaration, "Incomplete enum: "+en.join(un)+" declaration")
			}
			glist[i] = v.copy()
		}
		if len(glist) != len(en.generics) {
			f := fmt.Sprint("Missing generic for enum: " + en.join(un) + "\"\nImplementation list was " + genericslist(glist))
			return "", err(me, ECodeEnumMissingGeneric, f)
		}
		typed := en.name + genericslist(glist)
		if _, ok := module.enums[typed]; !ok {
			me.defineEnumImplGeneric(en, glist)
		}
		en = module.enums[typed]
	}
	typed := en.join(un)
	me.pushEnumParams(n, un, params, typed)
	return typed, nil
}

func (me *parser) buildEnum(n *node, module *hmfile) (*datatype, *parseError) {
	typed := me.token.value
	depth := me.token.depth
	if er := me.eat("id"); er != nil {
		return nil, er
	}
	en, ok := module.enums[typed]
	if !ok {
		return nil, err(me, ECodeEnumDoesNotExist, "Enum: "+typed+" does not exist")
	}
	uid := module.reference(typed)
	gsize := len(en.generics)
	if gsize > 0 {
		if me.token.is == "<" {
			gtypes, er := me.declareGeneric(len(en.generics))
			if er != nil {
				return nil, er
			}
			if len(gtypes) != len(en.generics) {
				return nil, err(me, ECodeEnumImplementationMismatch, "Enum \""+en.name+" with implementation "+fmt.Sprint(gtypes)+" does not match "+fmt.Sprint(en.generics))
			}
			typed = uid + genericslist(gtypes)
			if _, ok := module.enums[typed]; !ok {
				me.defineEnumImplGeneric(en, gtypes)
			}
			en = module.enums[typed]
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
				en = d.enum
			}
		}
	}

	if er := me.eat("."); er != nil {
		return nil, er
	}
	unvalue := me.token.value
	if er := me.eat("id"); er != nil {
		return nil, er
	}
	un := en.getType(unvalue)
	if un == nil {
		return nil, err(me, ECodeEnumDoesNotHaveType, "Enum: "+en.name+" does not have type: "+unvalue)
	}
	if n != nil && !en.simple {
		var er *parseError
		typed, er = me.enumParams(n, en, un, depth)
		if er != nil {
			return nil, er
		}
	} else {
		typed = en.join(un)
	}
	if me.hmfile != module {
		typed = module.reference(typed)
	}
	return getdatatype(module, typed)
}

func (me *parser) allocEnum(module *hmfile, hint *allocHint) (*node, *parseError) {
	n := nodeInit("enum")
	data, er := me.buildEnum(n, module)
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
