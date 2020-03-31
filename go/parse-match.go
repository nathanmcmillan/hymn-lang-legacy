package main

func (me *parser) parseMatch() (*node, *parseError) {
	depth := me.token.depth
	if er := me.eat("match"); er != nil {
		return nil, er
	}
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

	single := true
	multi := false

	if me.isNewLine() {
		single = false
		me.newLine()
	} else if me.token.is == ":" {
		me.next()
	} else {
		return nil, err(me, ECodeUnexpectedToken, "Expected line or `:`")
	}

	for {
		if single {
			if multi {
				break
			}
		} else if me.token.depth <= depth {
			break
		}
		multi = true
		if me.token.is == "id" {
			name := me.token.value
			if er := me.eat("id"); er != nil {
				return nil, er
			}
			caseNode := nodeInit(name)
			temp := ""
			if me.token.is == "(" {
				if er := me.eat("("); er != nil {
					return nil, er
				}
				temp = me.token.value
				if er := me.eat("id"); er != nil {
					return nil, er
				}
				if er := me.eat(")"); er != nil {
					return nil, er
				}
			}
			n.push(caseNode)
			if temp != "" {
				en, _, ok := matchType.isEnum()
				if !ok {
					return nil, err(me, ECodeEnumMatchRequired, "Enum required for matching but found: "+name)
				}
				dd, er := en.getuniondata(me.hmfile, name)
				if er != nil {
					return nil, er
				}
				tempd := me.hmfile.varInitWithData(dd, temp, false)
				me.hmfile.scope.variables[temp] = tempd
				tempv := nodeInit("variable")
				tempv.idata = newidvariable(me.hmfile, temp)
				tempv.copyData(tempd.data())
				caseNode.push(tempv)
			}
			if me.isNewLine() {
				me.newLine()
				b, er := me.block()
				if er != nil {
					return nil, er
				}
				n.push(b)
			} else if me.token.is == ":" {
				me.next()
				e, er := me.expression()
				if er != nil {
					return nil, er
				}
				n.push(e)
			} else {
				return nil, err(me, ECodeUnexpectedToken, "Expected line or `:`")
			}
			if me.isNewLine() {
				me.newLine()
			}
			if temp != "" {
				delete(me.hmfile.scope.variables, temp)
			}
		} else if me.token.is == "some" {
			if er := me.eat("some"); er != nil {
				return nil, er
			}
			if matchVar != nil {
				if !matchType.isSomeOrNone() {
					panic("type \"" + matchVar.name + "\" is not \"maybe\"")
				}
			}
			temp := ""
			if me.token.is == "(" {
				if er := me.eat("("); er != nil {
					return nil, er
				}
				temp = me.token.value
				if er := me.eat("id"); er != nil {
					return nil, er
				}
				if er := me.eat(")"); er != nil {
					return nil, er
				}
			}
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
			if me.isNewLine() {
				me.newLine()
				b, er := me.block()
				if er != nil {
					return nil, er
				}
				n.push(b)
			} else if me.token.is == ":" {
				me.next()
				e, er := me.expression()
				if er != nil {
					return nil, er
				}
				n.push(e)
			} else {
				return nil, err(me, ECodeUnexpectedToken, "Expected line or `:`")
			}
			if me.isNewLine() {
				me.newLine()
			}
			if temp != "" {
				delete(me.hmfile.scope.variables, temp)
			}
		} else if me.token.is == "none" {
			if er := me.eat("none"); er != nil {
				return nil, er
			}
			n.push(nodeInit("none"))
			if me.isNewLine() {
				me.newLine()
				b, er := me.block()
				if er != nil {
					return nil, er
				}
				n.push(b)
			} else if me.token.is == ":" {
				me.next()
				e, er := me.expression()
				if er != nil {
					return nil, er
				}
				n.push(e)
			} else {
				return nil, err(me, ECodeUnexpectedToken, "Expected line or `:`")
			}
			if me.isNewLine() {
				me.newLine()
			}
		} else if me.token.is == "_" {
			if er := me.eat("_"); er != nil {
				return nil, er
			}
			n.push(nodeInit("_"))
			if me.isNewLine() {
				me.newLine()
				b, er := me.block()
				if er != nil {
					return nil, er
				}
				n.push(b)
			} else if me.token.is == ":" {
				me.next()
				e, er := me.expression()
				if er != nil {
					return nil, er
				}
				n.push(e)
			} else {
				return nil, err(me, ECodeUnexpectedToken, "Expected line or `:`")
			}
			if me.isNewLine() {
				me.newLine()
			}
		} else if literal, ok := literals[me.token.is]; ok {
			if literal != matchType.print() {
				return nil, err(me, ECodeLiteralMismatch, "Literal does not match type.")
			}
			value := me.token.value
			caseNodes := nodeInit(me.token.is)
			if er := me.eat(me.token.is); er != nil {
				return nil, er
			}
			caseNodes.push(nodeInit(value))
			for me.token.is == "|" {
				if er := me.eat("|"); er != nil {
					return nil, er
				}
				if me.isNewLine() {
					me.newLine()
				}
				literal, ok := literals[me.token.is]
				if !ok {
					return nil, err(me, ECodeUnexpectedType, "Expecting matching type.")
				}
				if literal != matchType.print() {
					return nil, err(me, ECodeTypeMismatch, "Literal does not match type.")
				}
				value := me.token.value
				if er := me.eat(me.token.is); er != nil {
					return nil, er
				}
				caseNodes.push(nodeInit(value))
			}
			n.push(caseNodes)
			if me.isNewLine() {
				me.newLine()
				b, er := me.block()
				if er != nil {
					return nil, er
				}
				n.push(b)
			} else if me.token.is == ":" {
				me.next()
				e, er := me.expression()
				if er != nil {
					return nil, er
				}
				n.push(e)
			} else {
				return nil, err(me, ECodeUnexpectedToken, "Expected line or `:`")
			}
			if me.isNewLine() {
				me.newLine()
			}
		} else {
			return nil, err(me, ECodeUnexpectedToken, "Unknown match expression.")
		}
	}

	return n, nil
}
