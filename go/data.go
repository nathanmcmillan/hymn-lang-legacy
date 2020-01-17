package main

func functionSigToVarData(fsig *fnSig) *datatype {
	return getdatatype(nil, fsig.print())
}

func (me *hmfile) typeToVarDataWithAttributes(typed string, attributes map[string]string) *datatype {
	data := typeToVarData(me, typed)
	if _, ok := attributes["stack"]; ok {
		data.setIsOnStack(true)
	}
	return data
}

func (me *hmlib) literalType(typed string) *datatype {
	return getdatatype(nil, typed)
}

func typeToVarData(module *hmfile, typed string) *datatype {

	if module != nil && module.scope.fn != nil && module.scope.fn.aliasing != nil {
		if alias, ok := module.scope.fn.aliasing[typed]; ok {
			typed = alias
		}
	}

	return getdatatype(module, typed)
}

func (me *datatype) asVariable() *variable {
	v := &variable{}
	v.copyData(me)
	return v
}
