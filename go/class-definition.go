package main

import "fmt"

func (me *parser) defineClass() *parseError {
	if er := me.eat("class"); er != nil {
		return er
	}
	token := me.token
	name := token.value
	module := me.hmfile
	if _, ok := module.namespace[name]; ok {
		return err(me, ECodeNameConflict, "name \""+name+"\" already defined")
	}
	if er := me.eat("id"); er != nil {
		return er
	}
	typedGenerics, typedGenericsInterfaces, er := me.genericHeader()
	if er != nil {
		return er
	}
	generics := datatypels(typedGenerics)

	var interfaces map[string]*classInterface
	if me.token.is == "line" && me.peek().is == "implements" {
		if er := me.eat("line"); er != nil {
			return er
		}
	}
	if me.token.is == "implements" {
		if er := me.eat("implements"); er != nil {
			return er
		}
		interfaces = make(map[string]*classInterface)
		for {
			interfaceName := me.token.value
			if er := me.eat("id"); er != nil {
				return er
			}

			interfaceModule := module

			if m, ok := module.imports[interfaceName]; ok && me.token.is == "." {
				interfaceModule = m
				if er := me.eat("."); er != nil {
					return er
				}
				interfaceName = me.token.value
				if er := me.eat("id"); er != nil {
					return er
				}
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
					return err(me, ECodeClassAndInterfaceMissingGenerics, e)
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
				if er := me.eat("line"); er != nil {
					return er
				}
			}
			if me.token.is == "and" {
				if er := me.eat("and"); er != nil {
					return er
				}
				continue
			}
			break
		}
	}

	if er := me.eat("line"); er != nil {
		return er
	}

	uid := module.reference(name)

	module.namespace[uid] = "class"
	module.types[uid] = "class"

	module.namespace[name] = "class"
	module.types[name] = "class"

	classDef := classInit(module, name, generics, typedGenericsInterfaces, interfaces)

	module.defineOrder = append(module.defineOrder, &defineType{class: classDef})

	module.classes[uid] = classDef
	module.classes[name] = classDef

	me.program.classes[uid] = classDef

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
			if er := me.eat("id"); er != nil {
				return er
			}
			if getVariable(members, mname) != nil {
				return err(me, ECodeMemberNameConflict, "Member name '"+mname+"' already used")
			}
			if inList(generics, mname) >= 0 {
				return err(me, ECodeMemberNameConflict, "Cannot use '"+mname+"' as member name")
			}

			isptr := true
			if me.token.is == "'" {
				if er := me.eat("'"); er != nil {
					return er
				}
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
			if er := me.eat("line"); er != nil {
				return er
			}
			if mtype.isUnknown() && inList(generics, mtype.print()) < 0 {
				return err(me, ECodeClassTypeExpected, fmt.Sprintf("I could not find the declared type `%s`", mtype.print()))
			}
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
