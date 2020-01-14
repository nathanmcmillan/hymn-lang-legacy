package main

import (
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
			head.copyData(typeToVarData(me.hmfile, "?"))
		} else {
			head.copyData(sv.data())
		}
	} else {
		sv := from.getStatic(localvarname)
		if sv == nil {
			panic(me.fail() + "static variable \"" + localvarname + "\" in module \"" + from.name + "\" not found")
		} else {
			head.copyData(sv.data())
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
				head.copyData(sv.data())
				head.is = "root-variable"
			}
			data := head.data()
			if rootClass, ok := data.checkIsClass(); ok {
				me.eat(".")
				dotName := me.token.value
				me.eat("id")
				var member *node
				classOf, ok := rootClass.variables[dotName]
				if ok {
					member = nodeInit("member-variable")
					member.copyData(classOf.data())
					member.idata = &idData{}
					member.idata.module = from
					member.idata.name = dotName
					member.push(head)
				} else {
					funcVar, ok := data.getFunction(nameOfClassFunc(rootClass.name, dotName))
					if ok {
						member = me.callClassFunction(data.getmodule(), head, rootClass, funcVar)
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
						member.copyData(typeToVarData(me.hmfile, TokenInt))
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
					member.copyData(typeInUnion)
					member.value = dotIndexStr
					member.push(head)
					head = member
				}
			} else if data.checkIsSomeOrNone() {
				panic(me.fail() + "unexpected maybe type \"" + head.data().print() + "\" do you need a match statement?")
			} else {
				panic(me.fail() + "unexpected type \"" + head.data().print() + "\"")
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
			me.eat("[")
			if me.token.is == ":" {
				if !head.data().isArray() {
					panic(me.fail() + "root variable \"" + head.idata.name + "\" of type \"" + head.data().print() + "\" is not an array")
				}
				me.eat(":")
				member := nodeInit("array-to-slice")
				member.copyData(head.data())
				member.data().arrayToSlice()
				member.push(head)
				head = member
			} else {
				if !head.data().checkIsIndexable() {
					panic(me.fail() + "root variable \"" + head.idata.name + "\" of type \"" + head.data().print() + "\" is not indexable")
				}
				member := nodeInit("array-member")
				index := me.calc(0)
				member.copyData(head.data().getmember())
				member.push(index)
				member.push(head)
				head = member
			}
			me.eat("]")
		} else if me.token.is == "(" {
			var sig *fnSig
			if head.is == "variable" {
				sv := me.hmfile.getvar(head.idata.name)
				if sv == nil {
					panic(me.fail() + "variable \"" + head.value + "\" out of scope")
				}
				sig = sv.data().functionSignature()

			} else if head.is == "member-variable" {
				sig = head.data().functionSignature()
			}
			member := nodeInit("call")
			member.copyData(sig.returns)
			member.push(head)
			me.pushSigParams(member, sig)
			head = member

		} else {
			break
		}
	}
	return head
}
