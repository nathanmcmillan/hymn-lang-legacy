package main

import "fmt"

func (me *parser) defineEnum() *parseError {
	if er := me.eat("enum"); er != nil {
		return er
	}
	token := me.token
	name := token.value
	if _, ok := me.hmfile.namespace[name]; ok {
		return err(me, ECodeNameConflict, "name \""+name+"\" already defined")
	}
	if er := me.eat("id"); er != nil {
		return er
	}
	genericsOrder, interfaces, er := me.genericHeader()
	if er != nil {
		return er
	}
	if er := me.eat("line"); er != nil {
		return er
	}
	generics := datatypels(genericsOrder)

	uid := me.hmfile.reference(name)

	me.hmfile.namespace[uid] = "enum"
	me.hmfile.types[uid] = "enum"

	me.hmfile.namespace[name] = "enum"
	me.hmfile.types[name] = "enum"

	enumDef := enumInit(me.hmfile, name)

	me.hmfile.defineOrder = append(me.hmfile.defineOrder, &defineType{enum: enumDef})

	me.hmfile.enums[uid] = enumDef
	me.hmfile.enums[name] = enumDef

	types := make([]*union, 0)
	isSimple := true

	for {
		token := me.token
		if token.is == "line" {
			if er := me.eat("line"); er != nil {
				return er
			}
			break
		}
		if token.is == "eof" || token.is == "comment" {
			break
		}
		if token.is == "id" {
			typeName := token.value
			if er := me.eat("id"); er != nil {
				return er
			}
			if getUnionType(types, typeName) != nil {
				return err(me, ECodeNameConflict, "type name \""+typeName+"\" already used")
			}
			unionOrderedData := newordereddata()
			if me.token.is == "(" {
				if er := me.eat("("); er != nil {
					return er
				}
				if me.token.is == ")" {
					goto closing
				}
				isSimple = false
				if me.token.is == "line" {
					if er := me.eat("line"); er != nil {
						return er
					}
				}
				for {
					if me.token.is == ")" {
						break
					}
					key := me.token.value
					if er := me.eat("id"); er != nil {
						return er
					}
					unionArgType, er := me.declareType()
					if er != nil {
						return er
					}
					if unionArgType.isUnknown() && inList(generics, unionArgType.print()) < 0 {
						return err(me, ECodeClassTypeExpected, fmt.Sprintf("I could not find the declared type `%s`", unionArgType.print()))
					}
					unionOrderedData.push(key, unionArgType)

					if me.token.is == "," {
						if er := me.eat(","); er != nil {
							return er
						}
					} else if me.token.is == "line" {
						if er := me.eat("line"); er != nil {
							return er
						}
					} else {
						goto closing
					}
				}
			closing:
				if er := me.eat(")"); er != nil {
					return er
				}
			}
			if er := me.eat("line"); er != nil {
				return er
			}
			un := unionInit(me.hmfile, name, typeName, unionOrderedData)
			types = append(types, un)
			continue
		}
		return err(me, ECodeUnexpectedToken, "bad token \""+token.is+"\" in enum")
	}

	enumDef.finishInit(isSimple, types, generics, interfaces)

	for _, implementation := range enumDef.implementations {
		me.finishEnumGenericDefinition(implementation)
	}

	return nil
}
