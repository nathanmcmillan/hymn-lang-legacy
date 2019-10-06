package main

func (me *cfile) compileFunction(name string, fn *function) {
	if fn.forClass != nil && len(fn.forClass.impls) > 0 {
		return
	}
	args := fn.args
	expressions := fn.expressions
	block := ""
	me.pushScope()
	me.depth = 1
	for _, arg := range args {
		me.scope.variables[arg.name] = arg.variable
	}
	for _, expr := range expressions {
		c := me.eval(expr)
		if c.code != "" {
			block += fmc(me.depth) + c.code + me.maybeColon(c.code) + "\n"
		}
	}
	me.popScope()
	code := ""
	code += fmtassignspace(fn.typed.typeSig()) + me.hmfile.funcNameSpace(name) + "("
	for ix, arg := range args {
		if ix > 0 {
			code += ", "
		}
		code += arg.vdat.typeSigOf(arg.name, false)
	}
	head := code + ");\n"
	code += ") {\n"
	code += block
	code += "}\n\n"

	me.headFuncSection += head
	me.codeFn = append(me.codeFn, code)
}

func (me *cfile) compileMain(fn *function) {
	if len(fn.args) > 0 {
		panic("main can not have arguments")
	}
	expressions := fn.expressions
	codeblock := ""
	returns := false
	me.pushScope()
	me.depth = 1
	list := me.hmfile.program.hmorder
	for x := len(list) - 1; x >= 0; x-- {
		file := list[x]
		if file.needInit {
			codeblock += fmc(me.depth) + file.funcNameSpace("init") + "();\n"
		}
	}
	for _, expr := range expressions {
		c := me.eval(expr)
		if c.is == "return" {
			if c.getType() != TokenInt {
				panic("main must return int")
			} else {
				returns = true
			}
		}
		codeblock += fmc(me.depth) + c.code + me.maybeColon(c.code) + "\n"
	}
	if !returns {
		codeblock += fmc(me.depth) + "return 0;\n"
	}
	me.popScope()
	code := ""
	code += "int main() {\n"
	code += codeblock
	code += "}\n"

	me.headFuncSection += "int main();\n"
	me.codeFn = append(me.codeFn, code)
}
