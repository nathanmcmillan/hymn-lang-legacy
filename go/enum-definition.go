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

	me.hmfile.namespace[name] = "enum"
	me.hmfile.types[name] = "enum"

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
			unionList := make([]string, 0)
			unionGOrder := make([]string, 0)
			if me.token.is == "(" {
				isSimple = false
				me.eat("(")
				for {
					if me.token.is == ")" {
						break
					}
					if me.token.is == "," {
						me.eat(",")
						continue
					}
					unionArgType := me.token.value
					me.wordOrPrimitive()
					if _, ok := me.hmfile.getType(unionArgType); !ok {
						if _, ok2 := genericsDict[unionArgType]; ok2 {
							unionGOrder = append(unionGOrder, unionArgType)
						} else {
							panic(me.fail() + "union type name \"" + unionArgType + "\" does not exist")
						}
					}
					unionList = append(unionList, unionArgType)
				}
				me.eat(")")
			}
			me.eat("line")
			un := unionInit(me.hmfile, name, typeName, unionList, unionGOrder)
			typesOrder = append(typesOrder, un)
			typesMap[typeName] = un
			continue
		}
		panic(me.fail() + "bad token \"" + token.is + "\" in enum")
	}

	// TODO: recursive enum pointer

	enumDef := enumInit(me.hmfile, name, isSimple, typesOrder, typesMap, datatypels(genericsOrder), genericsDict)

	me.hmfile.enums[name] = enumDef
	me.hmfile.defineOrder = append(me.hmfile.defineOrder, &defineType{enum: enumDef})
}
