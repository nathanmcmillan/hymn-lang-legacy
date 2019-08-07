package main

import (
	"fmt"
	"strconv"
)

func (me *parser) eatvar(from *hmfile) *node {
	root := nodeInit("variable")
	localvarname := me.token.value
	if from == me.hmfile {
		root.value = localvarname
	} else {
		root.value = from.name + "." + localvarname
	}
	me.eat("id")
	for {
		if me.token.is == "." {
			if root.is == "variable" {
				sv := me.hmfile.getvar(root.value)
				if sv == nil {
					panic(me.fail() + "variable \"" + root.value + "\" out of scope")
				}
				root.typed = sv.typed
				root.is = "root-variable"
			}
			fmt.Println("eatvar root type :=", root.typed)
			data := me.hmfile.typeToVarData(root.typed)
			if rootClass, ok := data.checkIsClass(); ok {
				me.eat(".")
				dotName := me.token.value
				me.eat("id")
				var member *node
				classOf, ok := rootClass.variables[dotName]
				if ok {
					fmt.Println("member variable \"" + dotName + "\" is type \"" + classOf.typed + "\"")
					member = nodeInit("member-variable")
					member.typed = classOf.typed
					member.value = dotName
					member.push(root)
				} else {
					nameOfFunc := me.nameOfClassFunc(rootClass.name, dotName)
					funcVar, ok := data.module.functions[nameOfFunc]
					if ok {
						fmt.Println("class function \"" + dotName + "\" returns \"" + funcVar.typed.full + "\"")
						member = me.callClassFunction(data.module, root, rootClass, funcVar)
					} else {
						panic(me.fail() + "class variable or function \"" + dotName + "\" does not exist")
					}
				}
				root = member

			} else if rootEnum, rootUnion, ok := data.checkIsEnum(); ok {
				if rootUnion != nil {
					me.eat(".")
					dotIndexStr := me.token.value
					me.eat("int")
					dotIndex, _ := strconv.Atoi(dotIndexStr)
					if dotIndex > len(rootUnion.types) {
						panic(me.fail() + "index out of range for \"" + rootUnion.name + "\"")
					}
					typeInUnion := rootUnion.types[dotIndex]
					member := nodeInit("tuple-index")
					member.typed = typeInUnion.typed
					member.value = dotIndexStr
					member.push(root)
					root = member
				} else {
					panic(me.fail() + "enum \"" + rootEnum.name + "\" must be union type does not exist")
				}
				// } else if data.maybe {
				// 	member := nodeInit("maybe")
				// 	member.typed = data.maybeType
				// 	member.push(root)
				// 	root = member

				// } else if data.none {
				// 	member := nodeInit("none")
				// 	member.typed = data.typed
				// 	member.push(root)
				// 	root = member
			} else {
				panic(me.fail() + "type \"" + root.typed + "\" does not exist")
			}
		} else if me.token.is == "[" {
			if root.is == "variable" {
				sv := me.hmfile.getvar(root.value)
				if sv == nil {
					panic(me.fail() + "variable out of scope")
				}
				root.typed = sv.typed
				root.is = "root-variable"
			}
			if !checkIsArray(root.typed) {
				panic(me.fail() + "root variable \"" + root.value + "\" of type \"" + root.typed + "\" is not array")
			}
			atype := typeOfArray(root.typed)
			me.eat("[")
			member := nodeInit("array-member")
			index := me.calc(0)
			member.typed = atype
			member.push(index)
			member.push(root)
			root = member
			me.eat("]")
		} else {
			break
		}
	}
	if root.is == "variable" {
		if from == me.hmfile {
			sv := from.getvar(localvarname)
			if sv == nil {
				root.typed = "?"
			} else {
				root.typed = sv.typed
			}
		} else {
			sv := from.getStatic(localvarname)
			if sv == nil {
				panic(me.fail() + "static variable \"" + localvarname + "\" in module \"" + from.name + "\" not found")
			} else {
				root.typed = sv.typed
			}
		}
	}
	return root
}
