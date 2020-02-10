package main

func (me *parser) defineEnum() {
	me.eat("enum")
	token := me.token
	name := token.value
	if _, ok := me.hmfile.namespace[name]; ok {
		panic(me.fail() + "name \"" + name + "\" already defined")
	}
	me.eat("id")
	genericsOrder, genericsDict := me.genericHeader()
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

	typesOrder := make([]*union, 0)
	typesMap := make(map[string]*union)
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
			if _, ok := typesMap[typeName]; ok {
				panic(me.fail() + "type name \"" + typeName + "\" already used")
			}
			unionOrderedData := newordereddata()
			unionGOrder := make([]string, 0)
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
					declare := me.token.is
					unionArgType := me.declareType()
					if _, ok2 := genericsDict[declare]; ok2 {
						unionGOrder = append(unionGOrder, declare)
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
			un := unionInit(me.hmfile, name, typeName, unionOrderedData, unionGOrder)
			typesOrder = append(typesOrder, un)
			typesMap[typeName] = un
			continue
		}
		panic(me.fail() + "bad token \"" + token.is + "\" in enum")
	}

	enumDef.finishInit(isSimple, typesOrder, typesMap, datatypels(genericsOrder), genericsDict)

	for _, implementation := range enumDef.implementations {
		me.finishEnumGenericDefinition(implementation)
	}
}
