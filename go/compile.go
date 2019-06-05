package main

import "fmt"

func compile(p *program) string {
	// main := me.functions["main"]
	// delete(me.functions, "main")

	cf := cfileInit()

	code := ""
	code += "#include <stdio.h>\n\n"

	fcode := make(map[string]string)
	for _, f := range p.functionOrder {
		fcode[f] = cf.function(f, p.functions[f])
		code += fcode[f]
	}

	fmt.Println("===")
	return code
}

func (me *cfile) eval(n *node) *cnode {
	op := n.is
	if op == "assign" {
		name := n.value
		in := me.eval(n.has[0])
		code := name + " = " + in.code + ";"
		if _, ok := me.scope.variables[name]; !ok {
			code = in.typed + " " + code
			me.scope.variables[name] = varInit(in.typed, name)
		}
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "call" {
		code := me.call(n.value, n.has)
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "+" || op == "-" || op == "*" || op == "/" {
		code := me.eval(n.has[0]).code
		code += op
		code += me.eval(n.has[1]).code
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "int" || op == "string" {
		cn := codeNode(n.is, n.value, n.typed, n.value)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "id" {
		name := n.value
		_, ok := me.scope.variables[name]
		if !ok {
			panic("unknown variable " + name)
		}
		cn := codeNode(n.is, n.value, n.typed, n.value)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "return" {
		in := me.eval(n.has[0])
		cn := codeNode(n.is, n.value, n.typed, "return "+in.code+";")
		cn.push(in)
		fmt.Println(cn.string(0))
		return cn
	}
	panic("eval unknown operation " + op)
}

func (me *cfile) main() string {
	return ""
}

func (me *cfile) free(name string) string {
	return "free(" + name + ");"
}

func (me *cfile) object(name string, fields []*variable) string {
	code := "struct " + name + " {\n"
	for _, field := range fields {
		code += field.is + " " + field.name + ";"
	}
	code += "};\n"
	return code
}

func (me *cfile) construct(is, name string) string {
	code := fmt.Sprint("struct {} *{} = (struct {} *)malloc(sizeof(struct {}));", is, name, is, is)
	return code
}

func (me *cfile) function(name string, fn *function) string {
	args := fn.args
	expressions := fn.expressions
	block := ""
	returns := "void"
	me.pushScope()
	for _, arg := range args {
		me.scope.variables[arg.name] = arg
	}
	for _, expr := range expressions {
		c := me.eval(expr)
		if c.is == "return" {
			returns = c.typed
		}
		block += fmc(1) + c.code + "\n"
	}
	me.popScope()
	code := ""
	code += returns + " " + name + "("
	for ix, arg := range args {
		if ix > 0 {
			code += ", "
		}
		code += arg.is + " " + arg.name
	}
	code += ")\n{\n"
	code += block
	code += "}\n\n"
	return code
}

func (me *cfile) call(name string, parameters []*node) string {
	if name == "echo" {
		param := me.eval(parameters[0])
		if param.typed == "string" {
			return "printf(\"%s\\n\", \"" + param.code + "\");"
		} else if param.typed == "int" {
			return "printf(\"%d\\n\", " + param.code + ");"
		}
		panic("argument for echo was " + param.string(0))
	}
	code := name + "("
	for ix, parameter := range parameters {
		if ix > 0 {
			code += ", "
		}
		code += me.eval(parameter).code
	}
	code += ")"
	return code
}
