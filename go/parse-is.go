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
		} else if is == "none" {
			right = nodeInit("none")
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
