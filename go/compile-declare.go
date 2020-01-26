package main

func (me *cfile) compileDeclare(n *node) string {
	if n.is != "variable" {
		return me.eval(n).code()
	}
	if n.idata == nil {
		return ""
	}
	code := ""
	idata := n.idata
	name := idata.name
	v := me.getvar(name)
	if v == nil {
		global := false
		useStack := false
		if idata.isGlobal() {
			global = true
		}
		if _, ok := n.attributes["stack"]; ok || n.data().isOnStack() {
			useStack = true
		}
		mutable := false
		if _, ok := n.attributes["mutable"]; ok {
			mutable = true
		}
		data := n.data()
		newVar := me.hmfile.varInitFromData(data, name, mutable)
		me.scope.variables[name] = newVar
		if global {
			newVar.cName = idata.getcname()
			// code += fmtassignspace(data.noMallocTypeSig())
			// code += fmtassignspace(data.typeSig())
			code += data.typeSigOf(newVar.cName, true)
		} else if useStack {
			code += fmtassignspace(data.typeSig())
			code += name
		} else {
			code += data.typeSigOf(name, mutable)
		}
	} else {
		code += v.cName
	}
	return code
}
