package main

import (
	"fmt"
	"strings"
)

var (
	globalClassPrefix = "Hm"
	globalFuncPrefix  = "hm_"
	globalVarPrefix   = "hm"
	definePrefix      = "HM_"
)

func (me *hmfile) generateC(folder, name string) string {
	cfile := me.cFileInit()

	head := ""
	guard := definePrefix + strings.ToUpper(name) + "_H"
	head += "#ifndef " + guard + "\n"
	head += "#define " + guard + "\n\n"
	head += "#include <stdio.h>\n"
	head += "#include <stdlib.h>\n"
	head += "#include <stdbool.h>\n"
	for importName := range me.imports {
		head += "#include \"" + importName + ".h\"\n"
	}
	head += "\n"

	code := ""
	code += "#include \"" + name + ".h\"\n"
	code += "\n"

	for _, c := range me.classOrder {
		head += cfile.defineClass(me.classes[c])
	}

	if len(me.statics) > 0 {
		for _, s := range me.statics {
			decl, impl := cfile.assignStatic(s)
			head += decl
			code += impl
		}
		head += "\n"
		code += "\n"
	}

	for _, f := range me.functionOrder {
		if f == "main" {
			decl, impl := cfile.mainc(me.functions[f])
			head += decl
			code += impl
		} else {
			decl, impl := cfile.function(f, me.functions[f])
			head += decl
			code += impl
		}
	}
	fmt.Println("=== end C ===")

	fileCode := folder + "/" + name + ".c"
	create(fileCode, code)

	head += "\n#endif\n"
	create(folder+"/"+name+".h", head)

	return fileCode
}

func (me *hmfile) allocstruct(typed string, n *node) string {
	if n.attribute("no-malloc") {
		return ""
	}
	typed = me.classNameSpace(typed)
	return fmt.Sprintf("(%s *)malloc(sizeof(%s))", typed, typed)
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
	size := me.eval(n.has[0])
	if n.attribute("no-malloc") {
		return "[" + size.code + "]"
	}
	atype := parseArrayType(n.typed)
	mtype := me.typesig(atype)
	return "(" + fmtptr(mtype) + ")malloc((" + size.code + ") * sizeof(" + mtype + "))"
}

func capital(id string) string {
	head := strings.ToUpper(id[0:1])
	body := id[1:]
	return head + body
}

func (me *hmfile) checkIsClass(typed string) bool {
	_, ok := me.classes[typed]
	return ok
}

func (me *cfile) typesig(typed string) string {
	if checkIsArray(typed) {
		arraytype := parseArrayType(typed)
		return fmtptr(me.typesig(arraytype))
	}
	module, truetype := me.hmfile.moduleAndName(typed)
	if module.checkIsClass(truetype) {
		return module.classNameSpace(truetype) + " *"
	} else if typed == "string" {
		return "char *"
	}
	return typed
}

func (me *cfile) noMallocTypeSig(typed string) string {
	if checkIsArray(typed) {
		arraytype := parseArrayType(typed)
		return fmtptr(me.noMallocTypeSig(arraytype))
	}
	if typed == "string" {
		return "char *"
	}
	module, truetype := me.hmfile.moduleAndName(typed)
	if module.checkIsClass(truetype) {
		return module.classNameSpace(typed)
	}
	return typed
}

func (me *cfile) declare(n *node) string {
	code := ""
	name := n.value
	if me.getvar(name) == nil {
		malloc := true
		if n.attribute("no-malloc") {
			malloc = false
		}
		mutable := false
		if n.attribute("mutable") {
			mutable = true
		}
		if malloc {
			typed := n.typed
			me.scope.variables[name] = me.hmfile.varInit(typed, name, mutable, malloc)
			codesig := fmtassignspace(me.typesig(typed))
			module, truetype := me.hmfile.moduleAndName(typed)
			if mutable {
				code = codesig
			} else if checkIsArray(typed) || module.checkIsClass(truetype) {
				code += codesig + "const "
			} else {
				code += "const " + codesig
			}
		} else {
			typed := n.typed
			module, _ := me.hmfile.moduleAndName(typed)
			newVar := me.hmfile.varInit(typed, name, mutable, malloc)
			newVar.cName = module.varNameSpace(name)
			me.scope.variables[name] = newVar
			codesig := fmtassignspace(me.noMallocTypeSig(typed))
			code += codesig
		}
	}
	return code
}

func (me *cfile) maybeLet(code string) string {
	if code == "" || strings.HasPrefix(code, "[") {
		return ""
	}
	return " = "
}

func (me *cfile) assignStatic(n *node) (string, string) {
	left := n.has[0]
	right := n.has[1]
	right.pushAttribute("no-malloc")
	decl := me.declare(left)
	rightcode := me.eval(right).code
	leftcode := me.eval(left).code
	setsign := me.maybeLet(rightcode)
	code := decl + leftcode + setsign + rightcode + ";\n"
	decl = "extern " + decl + leftcode
	if setsign == "" {
		decl += rightcode
	}
	decl += ";\n"
	return decl, code
}

func (me *cfile) assingment(n *node) string {
	code := ""
	left := n.has[0]
	right := n.has[1]
	if left.attribute("no-malloc") {
		right.pushAttribute("no-malloc")
	}
	if left.is == "variable" {
		code = me.declare(left)
	}
	rightcode := me.eval(right).code
	code += me.eval(left).code + me.maybeLet(rightcode) + rightcode
	return code
}

func (me *cfile) assignmentUpdate(n *node) string {
	left := me.eval(n.has[0])
	right := me.eval(n.has[1])
	return left.code + " " + n.is + " " + right.code
}

func (me *cfile) eval(n *node) *cnode {
	op := n.is
	if op == "=" {
		code := me.assingment(n)
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "+=" || op == "-=" || op == "*=" || op == "/=" {
		code := me.assignmentUpdate(n)
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "new" {
		module, className := me.hmfile.moduleAndName(n.typed)
		code := module.allocstruct(className, n)
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "call" {
		module, callName := me.hmfile.moduleAndName(n.value)
		code := me.call(module, callName, n.has)
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "+" || op == "-" || op == "*" || op == "/" {
		paren := n.attribute("parenthesis")
		code := ""
		if paren {
			code += "("
		}
		code += me.eval(n.has[0]).code
		code += " " + op + " "
		code += me.eval(n.has[1]).code
		if paren {
			code += ")"
		}
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "member-variable" {
		root := n.has[0]
		code := n.value
		for {
			if root.is == "root-variable" {
				module, varName := me.hmfile.moduleAndName(root.value)
				var vr *variable
				var cname string
				if module == me.hmfile {
					vr = me.getvar(varName)
					cname = vr.cName
				} else {
					vr = module.getstatic(varName)
					cname = module.varNameSpace(varName)
				}
				if checkIsArray(root.typed) {
					code = cname + code
				} else {
					code = cname + vr.memget() + code
				}
				break
			} else if root.is == "array-member" {
				index := me.eval(root.has[0])
				code = "[" + index.code + "]" + "->" + code
				root = root.has[1]
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
		module, varName := me.hmfile.moduleAndName(n.value)
		var v *variable
		var cname string
		if module == me.hmfile {
			v = me.getvar(varName)
			cname = v.cName
		} else {
			v = module.getstatic(varName)
			cname = module.varNameSpace(varName)
		}
		if v == nil {
			panic("unknown variable " + varName)
		}
		cn := codeNode(n.is, n.value, n.typed, cname)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "root-variable" {
		v := me.getvar(n.value)
		cn := codeNode(n.is, n.value, n.typed, v.cName)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "array-member" {
		index := me.eval(n.has[0])
		root := me.eval(n.has[1])
		code := root.code + "[" + index.code + "]"
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
		cn := codeNode(n.is, n.value, n.typed, "return "+in.code)
		cn.push(in)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "boolexpr" {
		code := me.eval(n.has[0]).code
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "equal" {
		paren := n.attribute("parenthesis")
		code := ""
		if paren {
			code += "("
		}
		code += me.eval(n.has[0]).code
		code += " == "
		code += me.eval(n.has[1]).code
		if paren {
			code += ")"
		}
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "not" {
		code := "!" + me.eval(n.has[0]).code
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "not-equal" {
		code := me.eval(n.has[0]).code
		code += " != "
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
	if op == "block" {
		return me.block(n)
	}
	if op == "break" {
		cn := codeNode(n.is, n.value, n.typed, "break;")
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "continue" {
		cn := codeNode(n.is, n.value, n.typed, "continue;")
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "for" {
		size := len(n.has)
		ix := 0
		code := ""
		if size > 2 {
			ix += 3
			vset := n.has[0]
			if vset.is != "=" {
				panic("for loop must start with assign")
			}
			vobj := vset.has[0]
			if vobj.is != "variable" {
				panic("for loop must assign a regular variable")
			}
			vname := vobj.value
			vexist := me.getvar(vname)
			if vexist == nil {
				code += me.declare(vobj) + vname + ";\n" + fmc(me.depth)
			}
			vinit := me.assingment(vset)
			condition := me.eval(n.has[1]).code
			inc := me.assignmentUpdate(n.has[2])
			code += "for (" + vinit + "; " + condition + "; " + inc + ")\n"
		} else if size > 1 {
			ix++
			code += "while (" + me.eval(n.has[0]).code + ")\n"
		} else {
			code += "while (true)\n"
		}
		code += fmc(me.depth) + "{\n"
		me.depth++
		code += me.eval(n.has[ix]).code
		me.depth--
		code += fmc(me.depth) + "}"
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
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
		if ix >= 2 && ix < hsize && n.has[ix].is == "block" {
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
	if _, ok := primitives[op]; ok {
		cn := codeNode(n.is, n.value, n.typed, n.value)
		fmt.Println(cn.string(0))
		return cn
	}
	panic("eval unknown operation " + n.string(0))
}

func (me *cfile) free(name string) string {
	return "free(" + name + ");"
}

func (me *cfile) defineClass(class *class) string {
	hmName := me.hmfile.classNameSpace(class.name)
	code := "struct " + hmName + "\n{\n"
	for _, name := range class.variableOrder {
		field := class.variables[name]
		code += fmc(1) + fmtassignspace(me.typesig(field.typed)) + field.name + ";\n"
	}
	code += "};\ntypedef struct " + hmName + " " + hmName + ";\n\n"
	return code
}

func (me *cfile) maybecolon(code string) string {
	if strings.HasSuffix(code, "}") {
		return ""
	}
	return ";"
}

func (me *cfile) block(n *node) *cnode {
	expressions := n.has
	code := ""
	for _, expr := range expressions {
		c := me.eval(expr)
		code += fmc(me.depth) + c.code
		code += me.maybecolon(code) + "\n"
	}
	cn := codeNode(n.is, n.value, n.typed, code)
	fmt.Println(cn.string(0))
	return cn
}

func (me *cfile) mainc(fn *function) (string, string) {
	if len(fn.args) > 0 {
		panic("main can not have arguments")
	}
	expressions := fn.expressions
	codeblock := ""
	returns := false
	me.pushScope()
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
		codeblock += fmc(me.depth) + c.code
		codeblock += me.maybecolon(codeblock) + "\n"
	}
	if !returns {
		codeblock += fmc(me.depth) + "return 0;\n"
	}
	me.popScope()
	code := ""
	code += "int main()\n{\n"
	code += codeblock
	code += "}\n"
	return "int main();\n", code
}

func (me *cfile) function(name string, fn *function) (string, string) {
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
		block += fmc(me.depth) + c.code
		block += me.maybecolon(block) + "\n"
	}
	me.popScope()
	code := ""
	code += fmtassignspace(me.typesig(fn.typed)) + me.hmfile.funcNameSpace(name) + "("
	for ix, arg := range args {
		if ix > 0 {
			code += ", "
		}
		typed := arg.typed
		module, truetype := me.hmfile.moduleAndName(typed)
		codesig := fmtassignspace(me.typesig(typed))
		if module.checkIsClass(truetype) || checkIsArray(typed) {
			code += codesig + "const "
		} else {
			code += "const " + codesig
		}
		code += arg.name
	}
	head := code + ");\n"
	code += ")\n{\n"
	code += block
	code += "}\n\n"
	return head, code
}

func (me *cfile) call(module *hmfile, name string, parameters []*node) string {
	if name == "echo" {
		param := me.eval(parameters[0])
		if param.typed == "string" {
			return "printf(\"%s\\n\", " + param.code + ")"
		} else if param.typed == "int" {
			return "printf(\"%d\\n\", " + param.code + ")"
		} else if param.typed == "float" {
			return "printf(\"%f\\n\", " + param.code + ")"
		} else if param.typed == "bool" {
			return "printf(\"%s\\n\", " + param.code + " ? \"true\" : \"false\")"
		}
		panic("argument for echo was " + param.string(0))
	}
	code := module.funcNameSpace(name) + "("
	for ix, parameter := range parameters {
		if ix > 0 {
			code += ", "
		}
		code += me.eval(parameter).code
	}
	code += ")"
	return code
}
