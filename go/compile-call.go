package main

func (me *cfile) compileCall(node *node) *codeblock {
	fn := node.fn
	if fn == nil {
		head := node.has[0]
		sig := head.data().fn
		code := "(*" + me.eval(head).code() + ")("
		parameters := node.has[1:len(node.has)]
		for ix, parameter := range parameters {
			if ix > 0 {
				code += ", "
			}
			arg := sig.args[ix]
			code += me.hintEval(parameter, arg.data()).code()
		}
		code += ")"
		return codeBlockOne(node, code)
	}
	module := fn.module
	name := fn.name
	parameters := node.has
	cb := me.compileBuiltin(node, name, parameters)
	if cb == nil {
		cb = &codeblock{}
		if fn.forClass != nil {
			name = fn.nameOfClassFunc()
		}
		code := module.funcNameSpace(name) + "("
		for ix, parameter := range parameters {
			if ix > 0 {
				code += ", "
			}
			arg := node.fn.args[ix]
			value := me.hintEval(parameter, arg.data())
			code += value.pop()
			cb.prepend(value.pre)
		}
		code += ")"
		cb.current = codeNode(node, code)
	}
	return cb
}
