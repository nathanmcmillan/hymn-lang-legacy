package main

func (me *parser) parseIs(left *node, op string, n *node) (*node, *parseError) {
	n.setData(newdataprimitive("bool"))
	if er := me.eat(op); er != nil {
		return nil, er
	}

	negate := false
	if me.token.is == "not" {
		if er := me.eat("not"); er != nil {
			return nil, er
		}
		negate = true
	}

	var right *node
	data := left.data()
	if data.isSomeOrNone() {
		is := ""
		if me.token.is == "some" {
			if er := me.eat("some"); er != nil {
				return nil, er
			}
			if negate {
				is = "none"
			} else {
				is = "some"
			}
		} else if me.token.is == "none" {
			if er := me.eat("none"); er != nil {
				return nil, er
			}
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
				if er := me.eat("("); er != nil {
					return nil, er
				}
				temp := me.token.value
				if er := me.eat("id"); er != nil {
					return nil, er
				}
				if er := me.eat(")"); er != nil {
					return nil, er
				}
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
				if er := me.eat("id"); er != nil {
					return nil, er
				}
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
					if er := me.eat("("); er != nil {
						return nil, er
					}
					temp := me.token.value
					if er := me.eat("id"); er != nil {
						return nil, er
					}
					if er := me.eat(")"); er != nil {
						return nil, er
					}
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
