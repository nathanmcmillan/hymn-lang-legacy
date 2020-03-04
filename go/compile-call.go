package main

func (me *cfile) compileCall(node *node) *codeblock {
	fn := node.fn
	if fn == nil {
		head := node.has[0]
		sig := head.data().functionSignature()
		code := "(*" + me.eval(head).code() + ")("
		parameters := node.has[1:len(node.has)]
		fnsize := len(sig.args)
		for ix, parameter := range parameters {
			if ix > 0 {
				code += ", "
			}
			var arg *funcArg
			if ix >= fnsize {
				arg = sig.argVariadic
			} else {
				arg = sig.args[ix]
			}
			code += me.hintEval(parameter, arg.data()).code()
		}
		code += ")"
		return codeBlockOne(node, code)
	}
	name := fn.getname()
	if !me.master {
		if _, ok := me.functions[name]; !ok {
			if _, ok2 := me.hmfile.functions[name]; ok2 {
				me.functions[name] = fn
				me.defineFunctionHead(fn)
			}
		}
	}
	parameters := node.has
	cb := me.compileBuiltin(node, name, parameters)
	if cb == nil {
		cb = &codeblock{}
		code := fn.getcname() + "("
		fnsize := len(node.fn.args)
		for ix, parameter := range parameters {
			if ix > 0 {
				code += ", "
			}
			var arg *funcArg
			if ix >= fnsize {
				arg = node.fn.argVariadic
			} else {
				arg = node.fn.args[ix]
			}
			value := me.hintEval(parameter, arg.data())
			code += value.pop()
			cb.prepend(value.pre)
		}
		code += ")"
		cb.current = codeNode(node, code)
	}
	return cb
}
