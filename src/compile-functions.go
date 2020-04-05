package main

import (
	"strings"
)

func cleanCode(code string) (string, bool) {
	if code == "" {
		return "", false
	}
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

func (me *cfile) happyOut(e *codeblock) string {
	block := ""
	for _, c := range e.flatten() {
		if c == nil {
			continue
		}
		code, ok := cleanCode(c.code)
		if ok {
			block += fmc(me.depth) + code + me.maybeColon(code) + me.maybeNewLine(code)
		}
	}
	return block
}

func (me *cfile) defineFunctionHead(fn *function) {
	me.addHeadFunc("\n" + me.functionHead(fn) + ";")
}

func (me *cfile) functionHead(fn *function) string {
	args := fn.args
	returns := fn.returns
	var code strings.Builder
	code.WriteString(fmtassignspace(returns.typeSig(me)) + fn.getcname() + "(")
	me.dependencyGraph(returns)
	for ix, arg := range args {
		me.dependencyGraph(arg.data())
		if ix > 0 {
			code.WriteString(", ")
		}
		if arg.used == false {
			code.WriteString("__attribute__((unused)) ")
		}
		code.WriteString(arg.data().typeSigOf(me, arg.name, false))
	}
	code.WriteString(")")
	return code.String()
}

func (me *cfile) compileFunction(name string, fn *function, use bool) {
	cls := fn.forClass
	if cls != nil {
		if len(cls.generics) > 0 {
			return
		} else if !use && cls.base != nil {
			return
		}
	} else if len(fn.generics) > 0 {
		return
	}

	args := fn.args
	expressions := fn.expressions
	var block strings.Builder
	me.pushScope()
	me.scope.fn = fn
	me.depth = 1
	for _, arg := range args {
		me.scope.variables[arg.name] = arg.variable
	}
	for _, expr := range expressions {
		e := me.eval(expr)
		block.WriteString(me.happyOut(e))
	}
	me.popScope()

	var code strings.Builder
	code.WriteString("\n")
	if len(fn.comments) > 0 {
		for _, comment := range fn.comments {
			code.WriteString("//")
			code.WriteString(comment)
			code.WriteString("\n")
		}
	}
	head := me.functionHead(fn)
	code.WriteString(head)
	code.WriteString(" {\n")
	code.WriteString(block.String())
	code.WriteString("}\n")

	me.addHeadFunc("\n" + head + ";")
	me.codeFn = append(me.codeFn, code)
}

func (me *cfile) compileMain(fn *function) {
	args := fn.args
	expressions := fn.expressions
	var block strings.Builder
	me.pushScope()
	me.getFuncScope().fn = fn
	me.depth = 1
	for _, arg := range args {
		me.scope.variables[arg.name] = arg.variable
	}
	list := me.hmfile.program.hmorder
	for x := len(list) - 1; x >= 0; x-- {
		file := list[x]
		if file.needStatic {
			block.WriteString(fmc(me.depth) + file.funcNameSpace("static") + "();\n")
		}
		if _, ok := file.functions["init"]; ok {
			block.WriteString(fmc(me.depth) + file.funcNameSpace("init") + "();\n")
		}
	}
	for _, expr := range expressions {
		e := me.eval(expr)
		block.WriteString(me.happyOut(e))
	}
	if expressions[len(expressions)-1].is != "return" {
		block.WriteString(fmc(me.depth) + "return 0;\n")
	}
	me.popScope()
	var code strings.Builder
	code.WriteString("\n")
	if len(fn.comments) > 0 {
		for _, comment := range fn.comments {
			code.WriteString("//")
			code.WriteString(comment)
			code.WriteString("\n")
		}
	}

	head := "int main("
	if len(args) > 0 {
		me.libReqAdd(HmLibSlice)
		me.libReqAdd(HmLibString)
		head += "int argc, char** argv"
	}
	head += ")"

	code.WriteString(head)
	code.WriteString(" {\n")

	if len(args) > 0 {
		code.WriteString(fmc(1))
		code.WriteString("char **")
		code.WriteString(args[0].name)
		code.WriteString(" = hmlib_array_to_slice(argv, sizeof(char *), argc);\n")
	}

	code.WriteString(block.String())
	code.WriteString("}\n")

	me.addHeadFunc("\n" + head + ";")
	me.codeFn = append(me.codeFn, code)
}
