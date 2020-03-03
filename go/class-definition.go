package main

func (me *parser) defineClass() *parseError {
	me.eat("class")
	token := me.token
	name := token.value
	module := me.hmfile
	if _, ok := module.namespace[name]; ok {
		return err(me, ECodeNameConflict, "name \""+name+"\" already defined")
	}
	me.eat("id")
	typedGenerics, typedGenericsInterfaces, er := me.genericHeader()
	if er != nil {
		return er
	}
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

			interfaceModule := module

			if m, ok := module.imports[interfaceName]; ok && me.token.is == "." {
				interfaceModule = m
				me.eat(".")
				interfaceName = me.token.value
				me.eat("id")
			}

			in, ok := interfaceModule.interfaces[interfaceName]
			if !ok {
				return err(me, ECodeUnknownInterface, "Unknown interface: "+interfaceName)
			}

			if in.requiresGenerics() {
				generics, er := me.declareGeneric(len(in.generics))
				if er != nil {
					return er
				}
				if len(generics) != len(in.generics) {
					e := me.fail()
					e += "Class '" + name + "' implementing interface '" + in.name + "' does not have correct generics."
					panic(e)
				}
				intname := in.name + genericslist(generics)
				if gotInterface, ok := interfaceModule.interfaces[intname]; ok {
					in = gotInterface
				} else {
					in, er = me.defineInterfaceImplementation(in, generics)
					if er != nil {
						return er
					}
				}
			}
			interfaces[in.uid()] = in
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

	classDef := classInit(module, name, generics, typedGenericsInterfaces, interfaces)

	module.defineOrder = append(module.defineOrder, &defineType{class: classDef})

	module.classes[uid] = classDef
	module.classes[name] = classDef

	members := make([]*variable, 0)

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
			if getVariable(members, mname) != nil {
				return err(me, ECodeMemberNameConflict, "Member name '"+mname+"' already used")
			}
			if i := inList(generics, mname); i >= 0 {
				return err(me, ECodeMemberNameConflict, "Cannot use '"+mname+"' as member name")
			}

			isptr := true
			if me.token.is == "'" {
				me.eat("'")
				isptr = false
			}

			mtype, er := me.declareType()
			if er != nil {
				return er
			}
			mtype.setIsPointer(isptr)
			if mcl, ok := mtype.isClass(); ok {
				if mcl == classDef {
					return err(me, ECodeClassRecursiveDefinition, "recursive type definition for \""+classDef.name+"\"")
				}
			}
			me.eat("line")
			members = append(members, mtype.getnamedvariable(mname, true))
			continue
		}
		return err(me, ECodeUnexpectedToken, "bad token \""+token.is+"\" in class")
	}

	classDef.initMembers(members)

	for _, implementation := range classDef.implementations {
		me.finishClassGenericDefinition(implementation)
	}

	return nil
}
