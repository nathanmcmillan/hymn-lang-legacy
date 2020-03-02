package main

func (me *parser) defineEnum() *parseError {
	me.eat("enum")
	token := me.token
	name := token.value
	if _, ok := me.hmfile.namespace[name]; ok {
		return err(me, "name \""+name+"\" already defined")
	}
	me.eat("id")
	genericsOrder, interfaces, er := me.genericHeader()
	if er != nil {
		return er
	}
	me.eat("line")

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
			me.eat("line")
			break
		}
		if token.is == "eof" || token.is == "comment" {
			break
		}
		if token.is == "id" {
			typeName := token.value
			me.eat("id")
			if getUnionType(types, typeName) != nil {
				return err(me, "type name \""+typeName+"\" already used")
			}
			unionOrderedData := newordereddata()
			if me.token.is == "(" {
				me.eat("(")
				if me.token.is == ")" {
					goto closing
				}
				isSimple = false
				if me.token.is == "line" {
					me.eat("line")
				}
				for {
					if me.token.is == ")" {
						break
					}
					key := me.token.value
					me.eat("id")
					unionArgType, er := me.declareType()
					if er != nil {
						return er
					}
					unionOrderedData.push(key, unionArgType)

					if me.token.is == "," {
						me.eat(",")
					} else if me.token.is == "line" {
						me.eat("line")
					} else {
						goto closing
					}
				}
			closing:
				me.eat(")")
			}
			me.eat("line")
			un := unionInit(me.hmfile, name, typeName, unionOrderedData)
			types = append(types, un)
			continue
		}
		return err(me, "bad token \""+token.is+"\" in enum")
	}

	enumDef.finishInit(isSimple, types, datatypels(genericsOrder), interfaces)

	for _, implementation := range enumDef.implementations {
		me.finishEnumGenericDefinition(implementation)
	}

	return nil
}
