package main

func (me *parser) defineInterface() *parseError {
	me.eat("interface")
	token := me.token
	name := token.value
	module := me.hmfile
	if _, ok := module.namespace[name]; ok {
		return err(me, "name \""+name+"\" already defined")
	}
	me.eat("id")
	generics, _, er := me.genericHeader()
	if er != nil {
		return er
	}
	me.eat("line")

	uid := module.reference(name)

	module.namespace[uid] = "interface"
	module.types[uid] = "interface"

	module.namespace[name] = "interface"
	module.types[name] = "interface"

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
				return err(me, "Name \""+mname+"\" already used")
			}
			mtype, er := me.declareType()
			if er != nil {
				return er
			}
			sig := mtype.funcSig
			if sig == nil {
				return err(me, "Interface must define a function signature, but found: "+mtype.error())
			}
			me.eat("line")
			self := fnArgInit(newdataanypointer().getvariable())
			sig.args = append([]*funcArg{self}, sig.args...)
			functions[mname] = sig
			continue
		}
		return err(me, "Bad token '"+token.is+"' in interface '"+name+"' definition")
	}

	interfaceDef := interfaceInit(module, name, generics, functions)

	module.interfaces[uid] = interfaceDef
	module.interfaces[name] = interfaceDef

	return nil
}
