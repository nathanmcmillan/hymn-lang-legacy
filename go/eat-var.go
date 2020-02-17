package main

func (me *parser) eatvar(from *hmfile) *node {
	head := nodeInit("variable")
	localvarname := me.token.value
	head.idata = newidvariable(from, localvarname)
	if from == me.hmfile {
		sv := from.getvar(localvarname)
		if sv == nil {
			head.copyData(getdatatype(me.hmfile, "?"))
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
					panic(me.fail() + "Variable '" + head.idata.name + "' out of scope")
				}
				head.copyData(sv.data())
				head.is = "root-variable"
			}
			data := head.data()
			if rootClass, ok := data.isClass(); ok {
				me.eat(".")
				dotName := me.token.value
				me.eat("id")
				var member *node
				if classVar, ok := rootClass.variables[dotName]; ok {
					member = nodeInit("member-variable")
					member.copyData(classVar.data())
					member.idata = newidvariable(from, dotName)
					member.push(head)
				} else if classFunc, ok := rootClass.functions[dotName]; ok {
					member = me.callClassFunction(data.getmodule(), head, rootClass, classFunc)
				} else {
					panic(me.fail() + "Class: " + rootClass.name + " does not have a variable or function named: " + dotName)
				}
				head = member

			} else if rootEnum, rootUnion, ok := data.isEnum(); ok {
				if rootUnion == nil {
					peek := me.peek().value
					if peek == "index" {
						me.eat(".")
						me.eat("id")
						member := nodeInit("member-variable")
						member.copyData(getdatatype(me.hmfile, TokenInt))
						member.idata = newidvariable(from, "class")
						member.push(head)
						head = member
					} else {
						panic(me.fail() + "enum \"" + rootEnum.name + "\" must be union type; missing root enum")
					}
				} else {
					me.eat(".")
					key := me.token.value
					me.eat("id")
					typeInUnion, ok := rootUnion.types.table[key]
					if !ok {
						panic(me.fail() + "Union key: " + key + " does not exist for: " + rootUnion.name)
					}
					member := nodeInit("union-member-variable")
					member.copyData(typeInUnion)
					member.value = key
					member.push(head)
					head = member
				}
			} else if data.isSomeOrNone() {
				panic(me.fail() + "Unexpected maybe type \"" + head.data().print() + "\". Do you need a match statement?")
			} else {
				panic(me.fail() + "Unknown type: " + head.data().error())
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
				member.data().convertArrayToSlice()
				member.push(head)
				head = member
			} else {
				if !head.data().isIndexable() {
					panic(me.fail() + "root variable \"" + head.idata.name + "\" of type \"" + head.data().print() + "\" is not indexable")
				}
				member := nodeInit("array-member")
				index := me.calc(0, nil)
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
					panic(me.fail() + "variable \"" + head.idata.name + "\" not found in scope.")
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
