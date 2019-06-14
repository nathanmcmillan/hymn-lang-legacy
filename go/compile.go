package main

import (
	"fmt"
	"strings"
)

func compile(p *program) string {
	cf := cfileInit()
	cf.classes = p.classes
	cf.types = p.types
	cf.primitives = p.primitives

	code := ""
	code += "#include <stdio.h>\n"
	code += "#include <stdlib.h>\n"
	code += "#include <stdbool.h>\n"
	code += "\n"

	for _, c := range p.classOrder {
		code += cf.object(p.classes[c])
	}

	for _, f := range p.functionOrder {
		var s string
		if f == "main" {
			s = cf.mainc(p.functions[f])
		} else {
			s = cf.function(f, p.functions[f])
		}
		code += s
	}
	fmt.Println("===")
	return code
}

func (me *cfile) construct(class string) string {
	return fmt.Sprintf("(%s *)malloc(sizeof(%s))", class, class)
}

func fmtptr(ptr string) string {
	if strings.HasSuffix(ptr, "*") {
		return ptr + "*"
	}
	return ptr + " *"
}

func fmtassignspace(s string) string {
	if strings.HasSuffix(s, "*") {
		return s
	}
	return s + " "
}

func (me *cfile) allocarray(n *node) string {
	size := n.has[0].value
	atype := parseArrayType(n.typed)
	typed := me.typesig(atype)
	code := "(" + fmtptr(typed) + ")malloc(" + size + " * sizeof(" + typed + "))"
	return code
}

func (me *cfile) typesig(typed string) string {
	if _, ok := me.classes[typed]; ok {
		return typed + " *"
	} else if typed == "string" {
		return "char *"
	}
	if checkIsArray(typed) {
		atype := parseArrayType(typed)
		return fmtptr(me.typesig(atype))
	}
	return typed
}

func (me *cfile) assignvar(name, typed string) string {
	if _, ok := me.scope.variables[name]; !ok {
		me.scope.variables[name] = varInit(typed, name)
		return fmtassignspace(me.typesig(typed))
	}
	return ""
}

func (me *cfile) eval(n *node) *cnode {
	op := n.is
	if op == "assign" {
		right := me.eval(n.has[1])
		var code string
		nodeLeft := n.has[0]
		var left *cnode
		if nodeLeft.is == "variable" {
			code = me.assignvar(nodeLeft.value, right.typed)
			left = me.eval(nodeLeft)
		} else {
			left = me.eval(nodeLeft)
			code = ""
		}
		code += left.code + " = " + right.code + ";"
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "new" {
		code := me.construct(n.typed)
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
	if op == "member-variable" {
		root := n.has[0]
		code := n.value
		for {
			if root.is == "root-variable" {
				if checkIsArray(root.typed) {
					code = root.value + code
				} else {
					code = root.value + "->" + code
				}
				break
			} else if root.is == "array-member" {
				code = "[" + root.value + "]" + "->" + code
				root = root.has[0]
			} else if root.is == "member-variable" {
				code = root.value + "->" + code
				root = root.has[0]
			} else {
				panic("missing member variable")
			}
		}
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "variable" {
		name := n.value
		_, ok := me.scope.variables[name]
		if !ok {
			panic("unknown variable " + name)
		}
		cn := codeNode(n.is, n.value, n.typed, n.value)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "array-member" {
		root := n.has[0]
		code := root.value + "[" + n.value + "]"
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "array" {
		code := me.allocarray(n)
		cn := codeNode(n.is, n.value, n.typed, code)
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
	if op == "equal" {
		code := me.eval(n.has[0]).code
		code += " == "
		code += me.eval(n.has[1]).code
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == ">" || op == ">=" || op == "<" || op == "<=" {
		code := me.eval(n.has[0]).code
		code += " " + op + " "
		code += me.eval(n.has[1]).code
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "scope" {
		return me.enclose(n)
	}
	if op == "if" {
		hsize := len(n.has)
		code := "if (" + me.eval(n.has[0]).code + ")\n"
		code += fmc(me.depth) + "{\n"
		me.depth++
		code += me.eval(n.has[1]).code
		me.depth--
		code += fmc(me.depth) + "}"
		ix := 2
		for ix < hsize && n.has[ix].is == "elif" {
			code += "\n" + fmc(me.depth) + "else if (" + me.eval(n.has[ix].has[0]).code + ")\n" + fmc(me.depth) + "{\n"
			me.depth++
			code += me.eval(n.has[ix].has[1]).code
			me.depth--
			code += fmc(me.depth) + "}"
			ix++
		}
		if ix >= 2 && ix < hsize && n.has[ix].is == "scope" {
			code += "\n" + fmc(me.depth) + "else\n" + fmc(me.depth) + "{\n"
			me.depth++
			code += me.eval(n.has[ix]).code
			me.depth--
			code += fmc(me.depth) + "}"
		}
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "string" {
		cn := codeNode(n.is, n.value, n.typed, "\""+n.value+"\"")
		fmt.Println(cn.string(0))
		return cn
	}
	if _, ok := me.primitives[op]; ok {
		cn := codeNode(n.is, n.value, n.typed, n.value)
		fmt.Println(cn.string(0))
		return cn
	}
	panic("eval unknown operation " + n.string(0))
}

func (me *cfile) free(name string) string {
	return "free(" + name + ");"
}

func (me *cfile) object(class *class) string {
	code := "struct " + class.name + "\n{\n"
	for _, name := range class.variableOrder {
		field := class.variables[name]
		code += fmc(1) + fmtassignspace(me.typesig(field.is)) + field.name + ";\n"
	}
	code += "};\ntypedef struct " + class.name + " " + class.name + ";\n\n"
	return code
}

func (me *cfile) enclose(n *node) *cnode {
	expressions := n.has
	block := ""
	for _, expr := range expressions {
		c := me.eval(expr)
		block += fmc(me.depth) + c.code + "\n"
	}
	cn := codeNode(n.is, n.value, n.typed, block)
	fmt.Println(cn.string(0))
	return cn
}

func (me *cfile) mainc(fn *function) string {
	args := fn.args
	expressions := fn.expressions
	block := ""
	returns := false
	me.pushScope()
	for _, arg := range args {
		me.scope.variables[arg.name] = arg
	}
	me.depth = 1
	for _, expr := range expressions {
		c := me.eval(expr)
		if c.is == "return" {
			if c.typed != "int" {
				panic("main must return int")
			} else {
				returns = true
			}
		}
		block += fmc(me.depth) + c.code + "\n"
	}
	if !returns {
		block += fmc(me.depth) + "return 0;\n"
	}
	me.popScope()
	code := ""
	code += "int main("
	for ix, arg := range args {
		if ix > 0 {
			code += ", "
		}
		code += arg.is + " " + arg.name
	}
	code += ")\n{\n"
	code += block
	code += "}\n"
	return code
}

func (me *cfile) function(name string, fn *function) string {
	args := fn.args
	expressions := fn.expressions
	block := ""
	me.pushScope()
	me.depth = 1
	for _, arg := range args {
		me.scope.variables[arg.name] = arg
	}
	for _, expr := range expressions {
		c := me.eval(expr)
		block += fmc(me.depth) + c.code + "\n"
	}
	me.popScope()
	code := ""
	code += fmtassignspace(me.typesig(fn.typed)) + name + "("
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
			return "printf(\"%s\\n\", " + param.code + ");"
		} else if param.typed == "int" {
			return "printf(\"%d\\n\", " + param.code + ");"
		} else if param.typed == "float" {
			return "printf(\"%f\\n\", " + param.code + ");"
		} else if param.typed == "bool" {
			return "printf(\"%s\\n\", " + param.code + " ? \"true\" : \"false\");"
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
