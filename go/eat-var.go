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
			data := head.vdata
			if rootClass, ok := data.checkIsClass(); ok {
				me.eat(".")
				dotName := me.token.value
				me.eat("id")
				var member *node
				classOf, ok := rootClass.variables[dotName]
				if ok {
					member = nodeInit("member-variable")
					member.vdata = classOf.vdat
					member.idata = &idData{}
					member.idata.module = from
					member.idata.name = dotName
					member.push(head)
				} else {
					funcVar, ok := data.getFunction(nameOfClassFunc(rootClass.name, dotName))
					if ok {
						fmt.Println("class function \"" + dotName + "\" returns \"" + funcVar.typed.full + "\"")
						member = me.callClassFunction(data.module, head, rootClass, funcVar)
					} else {
						panic(me.fail() + "class \"" + rootClass.name + "\" does not have variable or function \"" + dotName + "\"")
					}
				}
				head = member

			} else if rootEnum, rootUnion, ok := data.checkIsEnum(); ok {
				if rootUnion == nil {
					peek := me.peek().value
					if peek == "index" {
						me.eat(".")
						me.eat("id")
						member := nodeInit("member-variable")
						member.vdata = me.hmfile.typeToVarData(TokenInt)
						member.idata = &idData{}
						member.idata.module = from
						member.idata.name = "type"
						member.push(head)
						head = member
					} else {
						panic(me.fail() + "enum \"" + rootEnum.name + "\" must be union type; missing root enum")
					}
				} else {
					me.eat(".")
					dotIndexStr := me.token.value
					me.eat(TokenIntLiteral)
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
				}
			} else {
				panic(me.fail() + "non primitive type \"" + head.vdata.full + "\" does not exist")
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
			if !head.asVar().checkIsArrayOrSlice() {
				panic(me.fail() + "root variable \"" + head.idata.name + "\" of type \"" + head.getType() + "\" is not an array")
			}
			me.eat("[")
			member := nodeInit("array-member")
			index := me.calc(0)
			member.vdata = head.vdata.memberType
			member.push(index)
			member.push(head)
			head = member
			me.eat("]")
		} else if me.token.is == "(" {
			var sig *fnSig
			if head.is == "variable" {
				sv := me.hmfile.getvar(head.idata.name)
				if sv == nil {
					panic(me.fail() + "variable \"" + head.value + "\" out of scope")
				}
				sig = sv.vdat.fn

			} else if head.is == "member-variable" {
				sig = head.vdata.fn
			}
			member := nodeInit("call")
			member.vdata = sig.typed
			member.push(head)
			me.pushSigParams(member, sig)
			head = member

		} else {
			break
		}
	}
	return head
}
