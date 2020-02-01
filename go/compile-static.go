package main

func (me *cfile) declareStatic(n *node) string {
	left := n.has[0]
	right := n.has[1]

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

func (me *cfile) defineStatic(v *variable) {
	// left := n.has[0]
	// declareCode := me.compileDeclare(left)
	// head := "\nextern " + declareCode + ";"
	head := "\nextern " + me.declareExtern(v) + ";"
	me.headExternSection.WriteString(head)
}

func (me *cfile) initStatic(n *node) *codeblock {
	left := n.has[0]
	right := n.has[1]

	declareCode := me.compileDeclare(left)
	rightCode := me.eval(right)

	setSign := me.maybeLet(rightCode.code(), right.attributes)

	if setSign == "" {
		return codeBlockOne(nodeInit(""), "")
	}

	code := declareCode + setSign + rightCode.pop() + ";\n"

	cb := &codeblock{}
	cb.current = codeNode(nodeInit(""), code)
	cb.prepend(rightCode.pre)

	return cb
}
