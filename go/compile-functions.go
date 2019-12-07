package main

func cleanCode(code string) (string, bool) {
	if code != "" {
		for {
			size := len(code)
			ch := code[size-1]
			if ch == '\n' || ch == '\t' || ch == ' ' {
				code = code[0 : size-1]
			} else {
				break
			}
		}
		return code, true
	}
	return code, false
}

func (me *cfile) happyOut(e *codeblock) string {
	block := ""
	for _, c := range e.flatten() {
		code, ok := cleanCode(c.code)
		if ok {
			block += fmc(me.depth) + code + me.maybeColon(code) + me.maybeNewLine(code)
		}
	}
	return block
}

func (me *cfile) compileFunction(name string, fn *function) {
	cls := fn.forClass
	if (cls != nil && len(cls.generics) > 0) || len(fn.generics) > 0 {
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
		e := me.eval(expr)
		block += me.happyOut(e)
	}
	me.popScope()
	code := ""
	code += fmtassignspace(fn.returns.typeSig()) + me.hmfile.funcNameSpace(name) + "("
	for ix, arg := range args {
		if ix > 0 {
			code += ", "
		}
		code += arg.data().typeSigOf(arg.name, false)
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
	block := ""
	returns := false
	me.pushScope()
	me.depth = 1
	list := me.hmfile.program.hmorder
	for x := len(list) - 1; x >= 0; x-- {
		file := list[x]
		if file.needInit {
			block += fmc(me.depth) + file.funcNameSpace("init") + "();\n"
		}
	}
	for _, expr := range expressions {
		e := me.eval(expr)
		if e.is() == "return" {
			if e.getType() != TokenInt {
				panic("main must return int")
			} else {
				returns = true
			}
		}
		block += me.happyOut(e)
	}
	if !returns {
		block += fmc(me.depth) + "return 0;\n"
	}
	me.popScope()
	code := ""
	code += "int main() {\n"
	code += block
	code += "}\n"

	me.headFuncSection += "int main();\n"
	me.codeFn = append(me.codeFn, code)
}
