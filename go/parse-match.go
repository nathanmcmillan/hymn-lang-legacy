package main

func (me *parser) parseIs(left *node, op string, n *node) (*node, *parseError) {
	n.copyData(newdataprimitive("bool"))
	me.eat(op)

	negate := false
	if me.token.is == "not" {
		me.eat("not")
		negate = true
	}

	var right *node
	data := left.data()
	if data.isSomeOrNone() {
		is := ""
		if me.token.is == "some" {
			me.eat("some")
			if negate {
				is = "none"
			} else {
				is = "some"
			}
		} else if me.token.is == "none" {
			me.eat("none")
			if negate {
				is = "some"
			} else {
				is = "none"
			}
		} else {
			return nil, err(me, ECodeBadIsStatement, "right side of \"is\" was \""+me.token.is+"\"")
		}
		if is == "some" {
			right = nodeInit("some")
			if me.token.is == "(" {
				if negate {
					return nil, err(me, ECodeNegationProhibited, "Negation not allowed when declaring a variable here.")
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
				if negate {
					return nil, err(me, ECodeNegationProhibited, "Negation not allowed when declaring a variable here.")
				}
				return nil, err(me, ECodeNoneTypeValueProhibited, "none type can't have a value here.")
			}
		}
	} else {
		baseEnum, _, ok := data.isEnum()
		if !ok {
			return nil, err(me, ECodeIsStatementExpectedEnum, "Left side of \"is\" must be enum but was: "+data.error())
		}
		if me.token.is == "id" {
			name := me.token.value
			if un := baseEnum.getType(name); un != nil {
				me.eat("id")
				newenum := newdataenum(me.hmfile, baseEnum, un, copydatalist(data.generics))
				if negate {
					right = nodeInit("negate-match-enum")
				} else {
					right = nodeInit("match-enum")
				}
				right.setData(newenum)
				if me.token.is == "(" {
					if negate {
						return nil, err(me, ECodeNegationProhibited, "Negation not allowed when declaring a variable here.")
					}
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
				var er *parseError
				right, er = me.calc(getInfixPrecedence(op), nil)
				if er != nil {
					return nil, er
				}
			}
		} else if checkIsPrimitive(me.token.is) {
			return nil, err(me, ECodeCannotMatchOnPrimitive, "can't match on a primitive. did you mean to use an enum implementation?")
		} else {
			return nil, err(me, ECodeUnknownType, "Unknown right side of is: "+me.token.is)
		}
	}
	n.push(left)
	n.push(right)
	return n, nil
}

func (me *parser) parseMatch() (*node, *parseError) {
	depth := me.token.depth
	me.eat("match")
	n := nodeInit("match")

	matching, er := me.calc(0, nil)
	if er != nil {
		return nil, er
	}

	matchType := matching.data()

	if _, ok := matchType.isClass(); ok {
		return nil, err(me, ECodeCannotMatchOnClass, "Can't match on class.\nFound: "+matchType.error())
	}

	_, un, ok := matchType.isEnum()
	if ok && un != nil {
		return nil, err(me, ECodeEnumMatchNotNeeded, "enum \""+matchType.print()+"\" does not need a match expression.")
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
					return nil, err(me, ECodeEnumMatchRequired, "Enum required for matching but found: "+name)
				}
				tempd, er := me.hmfile.varInit(en.name+"."+name, temp, false)
				if er != nil {
					return nil, er
				}
				me.hmfile.scope.variables[temp] = tempd
				tempv := nodeInit("variable")
				tempv.idata = newidvariable(me.hmfile, temp)
				tempv.copyData(tempd.data())
				caseNode.push(tempv)
			}
			if me.token.is == "line" {
				me.eat("line")
				b, er := me.block()
				if er != nil {
					return nil, er
				}
				n.push(b)
			} else {
				e, er := me.expression()
				if er != nil {
					return nil, er
				}
				n.push(e)
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
				b, er := me.block()
				if er != nil {
					return nil, er
				}
				n.push(b)
			} else {
				e, er := me.expression()
				if er != nil {
					return nil, er
				}
				n.push(e)
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
				b, er := me.block()
				if er != nil {
					return nil, er
				}
				n.push(b)
			} else {
				e, er := me.expression()
				if er != nil {
					return nil, er
				}
				n.push(e)
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
				b, er := me.block()
				if er != nil {
					return nil, er
				}
				n.push(b)
			} else {
				e, er := me.expression()
				if er != nil {
					return nil, er
				}
				n.push(e)
			}
			if me.token.is == "line" {
				me.eat("line")
			}
		} else if literal, ok := literals[me.token.is]; ok {
			if literal != matchType.print() {
				return nil, err(me, ECodeLiteralMismatch, "Literal does not match type.")
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
					return nil, err(me, ECodeUnexpectedType, "Expecting matching type.")
				}
				if literal != matchType.print() {
					return nil, err(me, ECodeTypeMismatch, "Literal does not match type.")
				}
				value := me.token.value
				me.eat(me.token.is)
				caseNodes.push(nodeInit(value))
			}
			n.push(caseNodes)
			me.eat("=>")
			if me.token.is == "line" {
				me.eat("line")
				b, er := me.block()
				if er != nil {
					return nil, er
				}
				n.push(b)
			} else {
				e, er := me.expression()
				if er != nil {
					return nil, er
				}
				n.push(e)
			}
			if me.token.is == "line" {
				me.eat("line")
			}
		} else {
			return nil, err(me, ECodeUnexpectedToken, "Unknown match expression.")
		}
	}

	return n, nil
}
