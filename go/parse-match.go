package main

func (me *parser) parseIs(left *node, op string, n *node) *node {
	n.copyData(me.hmfile.typeToVarData("bool"))
	me.eat(op)
	var right *node
	if left.data().none || left.data().maybe {
		if me.token.is == "some" {
			right = nodeInit("some")
			me.eat("some")
			if me.token.is == "(" {
				me.eat("(")
				temp := me.token.value
				me.eat("id")
				me.eat(")")
				tempd := me.hmfile.varInitFromData(left.data().someType, temp, false)
				tempv := nodeInit("variable")
				tempv.idata = &idData{}
				tempv.idata.module = me.hmfile
				tempv.idata.name = tempd.name
				tempv.copyData(tempd.data())
				tempv.push(left)
				varnode := &variableNode{tempv, tempd}
				me.hmfile.enumIsStack = append(me.hmfile.enumIsStack, varnode)
			}
		} else if me.token.is == "none" {
			right = nodeInit("none")
			me.eat("none")
		} else {
			panic(me.fail() + "right side of \"is\" was \"" + me.token.is + "\"")
		}
	} else {
		if _, _, ok := left.data().checkIsEnum(); !ok {
			panic(me.fail() + "left side of \"is\" must be enum but was \"" + left.data().full + "\"")
		}
		if me.token.is == "id" {
			name := me.token.value
			baseEnum, _, _ := left.data().checkIsEnum()
			if un, ok := baseEnum.types[name]; ok {
				me.eat("id")
				right = nodeInit("match-enum")
				right.copyData(me.hmfile.typeToVarData(baseEnum.name + "." + un.name))
				if me.token.is == "(" {
					me.eat("(")
					temp := me.token.value
					me.eat("id")
					me.eat(")")
					tempd := me.hmfile.varInitFromData(right.data(), temp, false)
					tempv := nodeInit("variable")
					tempv.idata = &idData{}
					tempv.idata.module = me.hmfile
					tempv.idata.name = tempd.name
					tempv.copyData(tempd.data())
					tempv.push(left)
					varnode := &variableNode{tempv, tempd}
					me.hmfile.enumIsStack = append(me.hmfile.enumIsStack, varnode)
				}
			} else {
				right = me.calc(getInfixPrecedence(op))
			}
		} else {
			panic(me.fail() + "unknown right side of \"is\"")
		}
	}
	n.push(left)
	n.push(right)
	return n
}

func (me *parser) parseMatch() *node {
	depth := me.token.depth
	me.eat("match")
	n := nodeInit("match")

	matching := me.calc(0)
	matchType := matching.data()
	var matchVar *variable
	if matching.is == "variable" {
		matchVar = me.hmfile.getvar(matching.idata.name)
	} else if matching.is == ":=" {
		matchVar = me.hmfile.getvar(matching.has[0].idata.name)
	}

	n.push(matching)

	me.eat("line")
	for {
		if me.token.depth <= depth {
			break
		} else if me.token.is == "id" {
			name := me.token.value
			me.eat("id")
			caseNode := nodeInit(name)
			temp := ""
			if me.token.is == "(" {
				me.eat("(")
				temp = me.token.value
				me.eat("id")
				me.eat(")")
			}
			me.eat("=>")
			n.push(caseNode)
			if temp != "" {
				en, _, ok := matchType.checkIsEnum()
				if !ok {
					panic(me.fail() + "only enums supported for matching")
				}
				tempd := me.hmfile.varInit(en.name+"."+name, temp, false)
				me.hmfile.scope.variables[temp] = tempd
				tempv := nodeInit("variable")
				tempv.idata = &idData{}
				tempv.idata.module = me.hmfile
				tempv.idata.name = temp
				tempv.copyData(tempd.data())
				caseNode.push(tempv)
			}
			if me.token.is == "line" {
				me.eat("line")
				n.push(me.block())
			} else {
				n.push(me.expression())
				me.eat("line")
			}
			if temp != "" {
				delete(me.hmfile.scope.variables, temp)
			}
		} else if me.token.is == "some" {
			me.eat("some")
			if matchVar != nil {
				if !matchType.maybe {
					panic("type \"" + matchVar.name + "\" is not \"maybe\"")
				}
			}
			temp := ""
			if me.token.is == "(" {
				me.eat("(")
				temp = me.token.value
				me.eat("id")
				me.eat(")")
			}
			me.eat("=>")
			some := nodeInit("some")
			n.push(some)
			if temp != "" {
				tempd := me.hmfile.varInitFromData(matchType.someType, temp, false)
				me.hmfile.scope.variables[temp] = tempd
				tempv := nodeInit("variable")
				tempv.idata = &idData{}
				tempv.idata.module = me.hmfile
				tempv.idata.name = temp
				tempv.copyData(tempd.data())
				some.push(tempv)
			}
			if me.token.is == "line" {
				me.eat("line")
				n.push(me.block())
			} else {
				n.push(me.expression())
				me.eat("line")
			}
			if temp != "" {
				delete(me.hmfile.scope.variables, temp)
			}
		} else if me.token.is == "_" {
			me.eat("_")
			me.eat("=>")
			n.push(nodeInit("_"))
			if me.token.is == "line" {
				me.eat("line")
				n.push(me.block())
			} else {
				n.push(me.expression())
				me.eat("line")
			}
		} else if me.token.is == "none" {
			me.eat("none")
			me.eat("=>")
			n.push(nodeInit("none"))
			if me.token.is == "line" {
				me.eat("line")
				n.push(me.block())
			} else {
				n.push(me.expression())
				me.eat("line")
			}
		} else {
			panic(me.fail() + "unknown match expression")
		}
	}
	return n
}
