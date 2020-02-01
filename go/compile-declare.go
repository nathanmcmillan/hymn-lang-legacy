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
		newVar := data.getnamedvariable(name, mutable)
		me.scope.variables[name] = newVar
		if global {
			newVar.cname = idata.getcname()
			// code += fmtassignspace(data.noMallocTypeSig())
			// code += fmtassignspace(data.typeSig())
			code += data.typeSigOf(newVar.cname, true)
		} else if useStack {
			code += fmtassignspace(data.typeSig())
			code += name
		} else {
			code += data.typeSigOf(name, mutable)
		}
	} else {
		code += v.cname
	}
	return code
}

func (me *cfile) declareExtern(vnode *variableNode) string {
	v := vnode.v
	name := v.name
	newv := v.data().getnamedvariable(name, v.mutable)
	newv.cname = vnode.n.has[0].idata.getcname()
	me.scope.variables[name] = newv
	return newv.data().typeSigOf(newv.cname, true)
}
