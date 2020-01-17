package main

func (me *hmfile) typeToVarDataWithAttributes(typed string, attributes map[string]string) *datatype {
	data := typeToVarData(me, typed)
	if _, ok := attributes["stack"]; ok {
		data.setIsOnStack(true)
	}
	return data
}

func typeToVarData(module *hmfile, typed string) *datatype {
	return getdatatype(module, typed)
}
