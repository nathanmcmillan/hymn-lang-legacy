package main

func (me *parser) defineInterface() {
	me.eat("interface")
	token := me.token
	name := token.value
	if _, ok := me.hmfile.namespace[name]; ok {
		panic(me.fail() + "name \"" + name + "\" already defined")
	}
	me.eat("id")
	me.eat("line")

	uid := me.hmfile.reference(name)

	me.hmfile.namespace[uid] = "interface"
	me.hmfile.types[uid] = "interface"

	me.hmfile.namespace[name] = "interface"
	me.hmfile.types[name] = "interface"

	functions := make(map[string]*fnSig)

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
			if _, ok := functions[mname]; ok {
				panic(me.fail() + "Name \"" + mname + "\" already used")
			}
			mtype := me.declareType()
			sig := mtype.funcSig
			if sig == nil {
				panic(me.fail() + "Interface must define a function signature, but found: " + mtype.error())
			}
			me.eat("line")
			self := fnArgInit(newdataany().getvariable())
			sig.args = append([]*funcArg{self}, sig.args...)
			functions[mname] = sig
			continue
		}
		panic(me.fail() + "Bad token '" + token.is + "' in interface '" + name + "' definition")
	}

	interfaceDef := interfaceInit(me.hmfile, name, functions)

	me.hmfile.interfaces[uid] = interfaceDef
	me.hmfile.interfaces[name] = interfaceDef
}
