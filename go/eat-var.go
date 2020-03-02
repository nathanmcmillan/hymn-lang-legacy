package main

func (me *parser) eatvar(from *hmfile) (*node, *parseError) {
	module := me.hmfile
	head := nodeInit("variable")
	localvarname := me.token.value
	head.idata = newidvariable(from, localvarname)
	if from == module {
		sv := from.getvar(localvarname)
		if sv == nil {
			head.copyData(newdataany())
		} else {
			head.copyData(sv.data())
		}
	} else {
		sv := from.getStatic(localvarname)
		if sv == nil {
			return nil, err(me, "static variable \""+localvarname+"\" in module \""+from.name+"\" not found")
		}
		head.copyData(sv.data())
	}
	me.eat("id")
	for {
		if me.token.is == "." {
			if head.is == "variable" {
				sv := module.getvar(head.idata.name)
				if sv == nil {
					return nil, err(me, "Variable '"+head.idata.name+"' out of scope")
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
				if classVar := rootClass.getVariable(dotName); classVar != nil {
					member = nodeInit("member-variable")
					member.copyData(classVar.data())
					member.idata = newidvariable(from, dotName)
					member.push(head)
				} else if classFunc := rootClass.getFunction(dotName); classFunc != nil {
					var er *parseError
					member, er = me.callClassFunction(data.getmodule(), head, rootClass, classFunc)
					if er != nil {
						return nil, er
					}
				} else {
					return nil, err(me, "Class: "+rootClass.name+" does not have a variable or function named: "+dotName)
				}
				head = member

			} else if data.isUnknown() && module.scope.fn.hasInterface(data) {
				me.eat(".")
				dotName := me.token.value
				me.eat("id")
				_, sig, ok := module.scope.fn.searchInterface(data, dotName)
				if !ok {
					return nil, err(me, "Generic '"+data.print()+" does not have an interface function called '"+dotName+"'")
				}
				member := nodeInit("call")
				member.copyData(sig.returns)
				member.push(head)
				me.pushSigParams(member, sig)
				head = member

			} else if rootEnum, rootUnion, ok := data.isEnum(); ok {
				if rootUnion == nil {
					peek := me.peek().value
					if peek == "index" {
						me.eat(".")
						me.eat("id")
						member := nodeInit("member-variable")
						newdata, er := getdatatype(module, TokenInt)
						if er != nil {
							return nil, er
						}
						member.copyData(newdata)
						member.idata = newidvariable(from, "class")
						member.push(head)
						head = member
					} else {
						return nil, err(me, "enum \""+rootEnum.name+"\" must be union type; missing root enum")
					}
				} else {
					me.eat(".")
					key := me.token.value
					me.eat("id")
					typeInUnion, ok := rootUnion.types.table[key]
					if !ok {
						return nil, err(me, "Union key: "+key+" does not exist for: "+rootUnion.name)
					}
					member := nodeInit("union-member-variable")
					member.copyData(typeInUnion)
					member.value = key
					member.push(head)
					head = member
				}
			} else if data.isSomeOrNone() {
				return nil, err(me, "Unexpected maybe type \""+head.data().print()+"\". Do you need a match statement?")
			} else {
				return nil, err(me, "Unknown type: "+head.data().error())
			}
		} else if me.token.is == "[" {
			if head.is == "variable" {
				sv := module.getvar(head.idata.name)
				if sv == nil {
					return nil, err(me, "variable out of scope")
				}
				head.copyTypeFromVar(sv)
				head.is = "root-variable"
			}
			me.eat("[")
			if me.token.is == ":" {
				if !head.data().isArray() {
					return nil, err(me, "root variable \""+head.idata.name+"\" of type \""+head.data().print()+"\" is not an array")
				}
				me.eat(":")
				member := nodeInit("array-to-slice")
				member.copyData(head.data())
				member.data().convertArrayToSlice()
				member.push(head)
				head = member
			} else {
				if !head.data().isIndexable() {
					return nil, err(me, "root variable \""+head.idata.name+"\" of type \""+head.data().print()+"\" is not indexable")
				}
				member := nodeInit("array-member")
				index, er := me.calc(0, nil)
				if er != nil {
					return nil, er
				}
				member.copyData(head.data().getmember())
				member.push(index)
				member.push(head)
				head = member
			}
			me.eat("]")
		} else if me.token.is == "(" {
			var sig *fnSig
			if head.is == "variable" {
				sv := module.getvar(head.idata.name)
				if sv == nil {
					return nil, err(me, "variable \""+head.idata.name+"\" not found in scope.")
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
	return head, nil
}
