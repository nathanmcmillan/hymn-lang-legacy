package main

func (me *parser) defineInterface() *parseError {
	if er := me.eat("interface"); er != nil {
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
	generics, _, er := me.genericHeader()
	if er != nil {
		return er
	}
	if er := me.eat("line"); er != nil {
	return er
}

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
			if er := me.eat("id"); er != nil {
	return er
}
			if _, ok := functions[mname]; ok {
				return err(me, ECodeNameConflict, "Name \""+mname+"\" already used")
			}
			mtype, er := me.declareType()
			if er != nil {
				return er
			}
			sig := mtype.funcSig
			if sig == nil {
				return err(me, ECodeInterfaceDefinitionType, "Interface must define a function signature, but found: "+mtype.error())
			}
			if er := me.eat("line"); er != nil {
	return er
}
			self := fnArgInit(newdataanypointer().getvariable())
			sig.args = append([]*funcArg{self}, sig.args...)
			functions[mname] = sig
			continue
		}
		return err(me, ECodeUnexpectedToken, "Bad token '"+token.is+"' in interface '"+name+"' definition")
	}

	interfaceDef := interfaceInit(module, name, generics, functions)

	module.interfaces[uid] = interfaceDef
	module.interfaces[name] = interfaceDef

	return nil
}
