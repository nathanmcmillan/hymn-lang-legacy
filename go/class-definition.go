package main

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

	uid := me.hmfile.reference(name)

	me.hmfile.namespace[uid] = "class"
	me.hmfile.types[uid] = "class"

	me.hmfile.namespace[name] = "class"
	me.hmfile.types[name] = "class"

	classDef := classInit(me.hmfile, name, datatypels(genericsOrder), genericsDict)

	me.hmfile.defineOrder = append(me.hmfile.defineOrder, &defineType{class: classDef})

	me.hmfile.classes[uid] = classDef
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

			mtype := me.declareType()
			mtype.setIsPointer(isptr)
			if mcl, ok := mtype.isClass(); ok {
				if mcl == classDef {
					panic(me.fail() + "recursive type definition for \"" + classDef.name + "\"")
				}
			}
			me.eat("line")
			memberOrder = append(memberOrder, mname)
			memberMap[mname] = mtype.getnamedvariable(mname, true)
			continue
		}
		panic(me.fail() + "bad token \"" + token.is + "\" in class")
	}

	classDef.initMembers(memberOrder, memberMap)

	for _, implementation := range classDef.implementations {
		me.finishClassDefinition(implementation)
	}
}
