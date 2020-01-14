package main

func (me *cfile) compileDeclare(n *node) string {
	if n.is != "variable" {
		return me.eval(n).code()
	}
	if n.idata == nil {
		return ""
	}
	code := ""
	name := n.idata.name
	v := me.getvar(name)
	if v == nil {
		global := false
		useStack := false
		if _, ok := n.attributes["global"]; ok {
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
			newVar.cName = data.getmodule().varNameSpace(name)
			code += fmtassignspace(data.noMallocTypeSig())
			code += newVar.cName
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

func (me *cfile) declareStatic(n *node) string {
	left := n.has[0]
	right := n.has[1]
	right.attributes["global"] = "true"

	declareCode := me.compileDeclare(left)
	rightCode := me.eval(right)
	setSign := me.maybeLet(rightCode.code(), right.attributes)

	head := "\nextern " + declareCode
	if setSign == "" {
		head += rightCode.code()
	}
	head += ";"
	me.headExternSection.WriteString(head)

	declareCode = "\n" + declareCode
	if setSign == "" {
		return declareCode + setSign + rightCode.code() + ";"
	}
	return declareCode + ";"
}

func (me *cfile) defineStatic(n *node) {
	left := n.has[0]
	declareCode := me.compileDeclare(left)
	head := "\nextern " + declareCode + ";"
	me.headExternSection.WriteString(head)
}
