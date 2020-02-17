package main

import "fmt"

func (me *parser) parseIs(left *node, op string, n *node) *node {
	n.copyData(newdataprimitive("bool"))
	me.eat(op)

	negate := false
	if me.token.is == "not" {
		me.eat("not")
		negate = true
	}
	fmt.Println("negate!", negate)

	var right *node
	data := left.data()
	if data.isSomeOrNone() {
		invert := false
		if me.token.is == "not" {
			invert = true
			me.eat("not")
		}
		is := ""
		if me.token.is == "some" {
			me.eat("some")
			if invert {
				is = "none"
			} else {
				is = "some"
			}
		} else if me.token.is == "none" {
			me.eat("none")
			if invert {
				is = "some"
			} else {
				is = "none"
			}
		} else {
			panic(me.fail() + "right side of \"is\" was \"" + me.token.is + "\"")
		}
		if is == "some" {
			right = nodeInit("some")
			if me.token.is == "(" {
				if invert {
					panic(me.fail() + "inversion not allowed with value here.")
				}
				me.eat("(")
				temp := me.token.value
				me.eat("id")
				me.eat(")")
				tempd := data.getmember().getnamedvariable(temp, false)
				tempv := nodeInit("variable")
				tempv.idata = newidvariable(me.hmfile, tempd.name)
				tempv.copyData(tempd.data())
				tempv.push(left)
				varnode := &variableNode{tempv, tempd}
				me.hmfile.enumIsStack = append(me.hmfile.enumIsStack, varnode)

				// TODO :: cleanup the above enumIsStack
				tempvv := nodeInit("variable")
				tempvv.idata = newidvariable(me.hmfile, tempd.name)
				tempvv.copyData(tempd.data())
				right.push(tempvv)
				//
			}
		} else if is == "none" {
			right = nodeInit("none")
			if me.token.is == "(" {
				if invert {
					panic(me.fail() + "inversion not allowed with value here.")
				}
				panic(me.fail() + "none type can't have a value here.")
			}
		}
	} else {
		baseEnum, _, ok := data.isEnum()
		if !ok {
			panic(me.fail() + "Left side of \"is\" must be enum but was: " + data.error())
		}
		if me.token.is == "id" {
			name := me.token.value
			if un, ok := baseEnum.types[name]; ok {
				me.eat("id")
				newenum := newdataenum(me.hmfile, baseEnum, un, copydatalist(data.generics))
				right = nodeInit("match-enum")
				right.setData(newenum)
				if me.token.is == "(" {
					me.eat("(")
					temp := me.token.value
					me.eat("id")
					me.eat(")")
					tempd := right.data().getnamedvariable(temp, false)
					tempv := nodeInit("variable")
					tempv.idata = newidvariable(me.hmfile, tempd.name)
					tempv.copyData(tempd.data())
					tempv.push(left)
					varnode := &variableNode{tempv, tempd}
					me.hmfile.enumIsStack = append(me.hmfile.enumIsStack, varnode)

					// TODO :: cleanup the above enumIsStack
					tempvv := nodeInit("variable")
					tempvv.idata = newidvariable(me.hmfile, tempd.name)
					tempvv.copyData(tempd.data())
					right.push(tempvv)
					//
				}
			} else {
				right = me.calc(getInfixPrecedence(op), nil)
			}
		} else if checkIsPrimitive(me.token.is) {
			panic(me.fail() + "can't match on a primitive. did you mean to use an enum implementation?")
		} else {
			panic(me.fail() + "Unknown right side of is: " + me.token.is)
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

	matching := me.calc(0, nil)
	matchType := matching.data()

	_, un, ok := matchType.isEnum()
	if ok && un != nil {
		panic(me.fail() + "enum \"" + matchType.print() + "\" does not need a match expression.")
	}

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
				en, _, ok := matchType.isEnum()
				if !ok {
					panic(me.fail() + "only enums supported for matching")
				}
				tempd := me.hmfile.varInit(en.name+"."+name, temp, false)
				me.hmfile.scope.variables[temp] = tempd
				tempv := nodeInit("variable")
				tempv.idata = newidvariable(me.hmfile, temp)
				tempv.copyData(tempd.data())
				caseNode.push(tempv)
			}
			if me.token.is == "line" {
				me.eat("line")
				n.push(me.block())
			} else {
				n.push(me.expression())
			}
			if me.token.is == "line" {
				me.eat("line")
			}
			if temp != "" {
				delete(me.hmfile.scope.variables, temp)
			}
		} else if me.token.is == "some" {
			me.eat("some")
			if matchVar != nil {
				if !matchType.isSomeOrNone() {
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
				tempd := matchType.getmember().getnamedvariable(temp, false)
				me.hmfile.scope.variables[temp] = tempd
				tempv := nodeInit("variable")
				tempv.idata = newidvariable(me.hmfile, temp)
				tempv.copyData(tempd.data())
				some.push(tempv)
			}
			if me.token.is == "line" {
				me.eat("line")
				n.push(me.block())
			} else {
				n.push(me.expression())
			}
			if me.token.is == "line" {
				me.eat("line")
			}
			if temp != "" {
				delete(me.hmfile.scope.variables, temp)
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
			}
			if me.token.is == "line" {
				me.eat("line")
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
			}
			if me.token.is == "line" {
				me.eat("line")
			}
		} else if literal, ok := literals[me.token.is]; ok {
			if literal != matchType.print() {
				panic(me.fail() + "Literal does not match type.")
			}
			value := me.token.value
			caseNodes := nodeInit(me.token.is)
			me.eat(me.token.is)
			caseNodes.push(nodeInit(value))
			for me.token.is == "|" {
				me.eat("|")
				if me.token.is == "line" {
					me.eat("line")
				}
				literal, ok := literals[me.token.is]
				if !ok {
					panic(me.fail() + "Expecting matching type.")
				}
				if literal != matchType.print() {
					panic(me.fail() + "Literal does not match type.")
				}
				value := me.token.value
				me.eat(me.token.is)
				caseNodes.push(nodeInit(value))
			}
			n.push(caseNodes)
			me.eat("=>")
			if me.token.is == "line" {
				me.eat("line")
				n.push(me.block())
			} else {
				n.push(me.expression())
			}
			if me.token.is == "line" {
				me.eat("line")
			}
		} else {
			panic(me.fail() + "Unknown match expression.")
		}
	}

	return n
}
