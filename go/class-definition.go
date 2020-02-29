package main

func (me *parser) defineClass() {
	me.eat("class")
	token := me.token
	name := token.value
	module := me.hmfile
	if _, ok := module.namespace[name]; ok {
		panic(me.fail() + "name \"" + name + "\" already defined")
	}
	me.eat("id")
	typedGenerics, _, _ := me.genericHeader()
	generics := datatypels(typedGenerics)

	var interfaces map[string]*classInterface
	if me.token.is == "line" && me.peek().is == "implements" {
		me.eat("line")
	}
	if me.token.is == "implements" {
		me.eat("implements")
		interfaces = make(map[string]*classInterface)
		for {
			interfaceName := me.token.value
			me.eat("id")
			in, ok := module.interfaces[interfaceName]
			if !ok {
				panic(me.fail() + "Unknown interface: " + interfaceName)
			}
			if in.requiresGenerics() {
				generics := me.declareGeneric(len(in.generics))
				if len(generics) != len(in.generics) {
					e := me.fail()
					e += "Class '" + name + "' implementing interface '" + in.name + "' does not have correct generics."
					panic(e)
				}
				intname := in.name + genericslist(generics)
				if gotInterface, ok := module.interfaces[intname]; ok {
					in = gotInterface
				} else {
					in = me.defineInterfaceImplementation(in, generics)
				}
			}
			interfaces[in.name] = in
			if me.token.is == "line" && me.peek().is == "and" {
				me.eat("line")
			}
			if me.token.is == "and" {
				me.eat("and")
				continue
			}
			break
		}
	}

	me.eat("line")

	uid := module.reference(name)

	module.namespace[uid] = "class"
	module.types[uid] = "class"

	module.namespace[name] = "class"
	module.types[name] = "class"

	classDef := classInit(module, name, generics, interfaces)

	module.defineOrder = append(module.defineOrder, &defineType{class: classDef})

	module.classes[uid] = classDef
	module.classes[name] = classDef

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
				panic(me.fail() + "Member name '" + mname + "' already used")
			}
			if i := inList(generics, mname); i >= 0 {
				panic(me.fail() + "Cannot use '" + mname + "' as member name")
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
		me.finishClassGenericDefinition(implementation)
	}
}
