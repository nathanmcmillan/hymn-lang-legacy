package main

import (
	"fmt"
	"strconv"
)

func (me *parser) eatvar(from *hmfile) *node {
	head := nodeInit("variable")
	localvarname := me.token.value
	head.idata = &idData{}
	head.idata.module = from
	head.idata.name = localvarname
	me.eat("id")
	for {
		if me.token.is == "." {
			if head.is == "variable" {
				sv := me.hmfile.getvar(head.idata.name)
				if sv == nil {
					panic(me.fail() + "variable \"" + head.value + "\" out of scope")
				}
				head.vdata = sv.vdat
				head.is = "root-variable"
			}
			fmt.Println("eatvar root type :=", head.vdata.full)
			data := head.vdata
			if rootClass, ok := data.checkIsClass(); ok {
				me.eat(".")
				dotName := me.token.value
				me.eat("id")
				var member *node
				classOf, ok := rootClass.variables[dotName]
				if ok {
					fmt.Println("member variable \"" + dotName + "\" is type \"" + classOf.vdat.full + "\"")
					member = nodeInit("member-variable")
					member.vdata = classOf.vdat
					member.idata = &idData{}
					member.idata.module = from
					member.idata.name = dotName
					member.push(head)
				} else {
					nameOfFunc := me.nameOfClassFunc(rootClass.name, dotName)
					funcVar, ok := data.module.functions[nameOfFunc]
					if ok {
						fmt.Println("class function \"" + dotName + "\" returns \"" + funcVar.typed.full + "\"")
						member = me.callClassFunction(data.module, head, rootClass, funcVar)
					} else {
						panic(me.fail() + "class variable or function \"" + dotName + "\" does not exist")
					}
				}
				head = member

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
					member.vdata = typeInUnion
					member.value = dotIndexStr
					member.push(head)
					head = member
				} else {
					panic(me.fail() + "enum \"" + rootEnum.name + "\" must be union type does not exist")
				}
			} else {
				panic(me.fail() + "type \"" + head.vdata.full + "\" does not exist")
			}
		} else if me.token.is == "[" {
			if head.is == "variable" {
				sv := me.hmfile.getvar(head.idata.name)
				if sv == nil {
					panic(me.fail() + "variable out of scope")
				}
				head.copyTypeFromVar(sv)
				head.is = "root-variable"
			}
			if !head.asVar().array {
				panic(me.fail() + "root variable \"" + head.idata.name + "\" of type \"" + head.getType() + "\" is not array")
			}
			me.eat("[")
			member := nodeInit("array-member")
			index := me.calc(0)
			member.vdata = head.vdata.typeInArray
			member.push(index)
			member.push(head)
			head = member
			me.eat("]")
		} else {
			break
		}
	}
	if head.is == "variable" {
		if from == me.hmfile {
			sv := from.getvar(localvarname)
			if sv == nil {
				head.vdata = me.hmfile.typeToVarData("?")
			} else {
				head.vdata = sv.vdat
			}
		} else {
			sv := from.getStatic(localvarname)
			if sv == nil {
				panic(me.fail() + "static variable \"" + localvarname + "\" in module \"" + from.name + "\" not found")
			} else {
				head.vdata = sv.vdat
			}
		}
	}
	return head
}
