package main

func (me *parser) genericHeader() ([]string, map[string]int) {
	order := make([]string, 0)
	dict := make(map[string]int)
	if me.token.is == "<" {
		me.eat("<")
		for {
			gname := me.token.value
			me.wordOrPrimitive()
			dict[gname] = len(order)
			order = append(order, gname)
			if me.token.is == "," {
				me.eat(",")
				continue
			}
			if me.token.is == ">" {
				break
			}
			panic(me.fail() + "bad token \"" + me.token.is + "\" in class generic")
		}
		me.eat(">")
	}
	return order, dict
}

func (me *parser) defineClass() {
	me.eat("type")
	token := me.token
	name := token.value
	if _, ok := me.hmfile.namespace[name]; ok {
		panic(me.fail() + "name \"" + name + "\" already defined")
	}
	me.eat("id")

	genericsOrder, genericsDict := me.genericHeader()
	me.eat("line")

	me.hmfile.namespace[name] = "type"
	me.hmfile.types[name] = ""
	me.hmfile.defineOrder = append(me.hmfile.defineOrder, name+"_type")

	classDef := classInit(me.hmfile, name, genericsOrder, genericsDict)
	me.hmfile.classes[name] = classDef

	memberOrder := make([]string, 0)
	memberMap := make(map[string]*variable)

	for {
		if me.token.is == "line" {
			break
		}
		if me.token.is == "eof" || me.token.is == "comment" {
			break
		}
		if me.token.is == "id" {
			mname := me.token.value
			me.eat("id")
			if _, ok := memberMap[mname]; ok {
				panic(me.fail() + "member name \"" + mname + "\" already used")
			}
			if _, ok := genericsDict[mname]; ok {
				panic(me.fail() + "cannot use \"" + mname + "\" as member name")
			}

			isptr := true
			if me.token.is == "'" {
				me.eat("'")
				isptr = false
			}

			mtype := me.declareType(false)
			mtype.isptr = isptr
			if mcl, ok := mtype.checkIsClass(); ok {
				if mcl == classDef {
					panic(me.fail() + "recursive type definition for \"" + classDef.name + "\"")
				}
			}
			me.eat("line")
			memberOrder = append(memberOrder, mname)
			memberMap[mname] = me.hmfile.varInitFromData(mtype, mname, true)
			continue
		}
		panic(me.fail() + "bad token \"" + token.is + "\" in class")
	}

	classDef.initMembers(memberOrder, memberMap)
}

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
	me.hmfile.types[name] = ""
	me.hmfile.defineOrder = append(me.hmfile.defineOrder, name+"_enum")

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
			un := me.hmfile.unionInit(typeName, unionList, unionGOrder)
			typesOrder = append(typesOrder, un)
			typesMap[typeName] = un
			continue
		}
		panic(me.fail() + "bad token \"" + token.is + "\" in enum")
	}
	me.hmfile.enums[name] = enumInit(me.hmfile, name, isSimple, typesOrder, typesMap, genericsOrder, genericsDict)
}
