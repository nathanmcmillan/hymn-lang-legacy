package main

import (
	"fmt"
	"strings"
)

func (me *hmfile) generateC(folder, name, libDir string) string {

	if debug {
		fmt.Println("=== generate C ===")
	}

	cfile := me.cFileInit()

	guard := me.defNameSpace(name)
	cfile.headPrefix += "#ifndef " + guard + "\n"
	cfile.headPrefix += "#define " + guard + "\n\n"

	cfile.headIncludeSection += "#include <stdio.h>\n"
	cfile.headIncludeSection += "#include <stdlib.h>\n"
	cfile.headIncludeSection += "#include <stdint.h>\n"
	cfile.headIncludeSection += "#include <inttypes.h>\n"
	cfile.headIncludeSection += "#include <stdbool.h>\n"

	cfile.headIncludeSection += "#include \"" + libDir + "/hmlib_string.h\"\n"
	cfile.headIncludeSection += "#include \"" + libDir + "/hmlib_slice.h\"\n"

	cfile.hmfile.program.sources["hmlib_string.c"] = libDir + "/hmlib_string.c"
	cfile.hmfile.program.sources["hmlib_slice.c"] = libDir + "/hmlib_slice.c"

	for _, importName := range me.importOrder {
		cfile.headIncludeSection += "#include \"" + importName + ".h\"\n"
	}
	cfile.headIncludeSection += "\n"

	code := ""
	code += "#include \"" + name + ".h\"\n"
	code += "\n"

	for _, c := range me.defineOrder {
		underscore := strings.LastIndex(c, "_")
		name := c[0:underscore]
		typed := c[underscore+1:]
		if typed == "type" {
			cfile.defineClass(me.classes[name])
		} else if typed == "enum" {
			cfile.defineEnum(me.enums[name])
		}
	}

	if len(me.statics) != 0 {
		me.needInit = true
	}

	if me.needInit {
		for _, s := range me.statics {
			code += cfile.declareStatic(s)
		}
		cfile.headExternSection += "\n"
		code += "\n"

		cfile.headFuncSection += "void " + me.funcNameSpace("init") + "();\n"
		code += "void " + me.funcNameSpace("init") + "() {\n"
		for _, s := range me.statics {
			code += cfile.initStatic(s)
		}
		code += "}\n\n"
	}

	for _, f := range me.functionOrder {
		if f == "main" {
			cfile.compileMain(me.functions[f])
		} else {
			cfile.compileFunction(f, me.functions[f])
		}
	}

	fmt.Println("=== end C ===")

	fileCode := folder + "/" + name + ".c"

	write(fileCode, code+strings.Join(cfile.codeFn, ""))

	cfile.headSuffix += "\n#endif\n"
	write(folder+"/"+name+".h", cfile.head())

	return fileCode
}

func (me *cfile) eval(n *node) *codeblock {
	return me.hintEval(n, nil)
}

func (me *cfile) compilePrefixPos(n *node) *codeblock {
	code := me.eval(n.has[0]).code()
	return codeBlockOne(n, code)
}

func (me *cfile) compilePrefixNeg(n *node) *codeblock {
	code := "-" + me.eval(n.has[0]).code()
	return codeBlockOne(n, code)
}

func (me *cfile) compileCast(n *node) *codeblock {
	typ := getCName(n.data().full)
	code := "(" + typ + ")" + me.eval(n.has[0]).code()
	return codeBlockOne(n, code)
}

func (me *cfile) compileBinaryOp(n *node) *codeblock {
	_, paren := n.attributes["parenthesis"]
	code := ""
	if paren {
		code += "("
	}
	code += me.eval(n.has[0]).code()
	code += " " + n.is + " "
	code += me.eval(n.has[1]).code()
	if paren {
		code += ")"
	}
	return codeBlockOne(n, code)
}

func (me *cfile) compileTupleIndex(n *node) *codeblock {
	dotIndexStr := n.value
	root := me.eval(n.has[0])
	data := root.data()
	_, un, _ := data.checkIsEnum()
	code := root.code() + "->"
	if len(un.types) == 1 {
		code += un.name
	} else {
		code += un.name + ".var" + dotIndexStr
	}
	return codeBlockOne(n, code)
}

func (me *cfile) compileMemberVariable(n *node) *codeblock {
	head := n.has[0]
	code := n.idata.name
	for {
		if head.is == "root-variable" {
			data := head.data()
			var vr *variable
			var cname string
			if head.idata.module == me.hmfile {
				vr = me.getvar(head.idata.name)
				cname = vr.cName
			} else {
				vr = data.module.getStatic(head.idata.name)
				cname = data.module.varNameSpace(head.idata.name)
			}
			if data.checkIsArrayOrSlice() {
				code = cname + code
			} else {
				code = cname + data.memPtr() + code
			}
			break
		} else if head.is == "array-member" {
			index := me.eval(head.has[0])
			code = "[" + index.code() + "]" + "->" + code
			head = head.has[1]
		} else if head.is == "member-variable" {
			if code[0] == '[' {
				code = head.idata.name + code
			} else {
				code = head.idata.name + head.data().memPtr() + code
			}
			head = head.has[0]
		} else {
			panic("missing member variable")
		}
	}
	return codeBlockOne(n, code)
}

func (me *cfile) compileFnPtr(n *node, hint *varData) *codeblock {
	code := ""
	fn := n.fn
	code += "&" + fn.module.funcNameSpace(fn.name)
	return codeBlockOne(n, code)
}

func (me *cfile) compileVariable(n *node, hint *varData) *codeblock {
	code := ""
	if n.idata.module == me.hmfile {
		v := me.getvar(n.idata.name)
		code = v.cName
		if hint != nil && hint.isptr && !v.data().isptr {
			code = "&" + code
		}
	} else {
		code = n.idata.module.varNameSpace(n.idata.name)
	}
	return codeBlockOne(n, code)
}

func (me *cfile) compileRawString(n *node) *codeblock {
	return codeBlockOne(n, "\""+n.value+"\"")
}

func (me *cfile) compileString(n *node) *codeblock {
	code := "hmlib_string_init(\"" + n.value + "\")"
	return codeBlockOne(n, code)
}

func (me *cfile) compileChar(n *node) *codeblock {
	code := "'" + n.value + "'"
	return codeBlockOne(n, code)
}

func (me *cfile) compileNone(n *node) *codeblock {
	code := "NULL"
	return codeBlockOne(n, code)
}

func (me *cfile) compileEqual(n *node) *codeblock {
	_, paren := n.attributes["parenthesis"]
	code := ""
	if paren {
		code += "("
	}
	code += me.eval(n.has[0]).code()
	code += " == "
	code += me.eval(n.has[1]).code()
	if paren {
		code += ")"
	}
	return codeBlockOne(n, code)
}

func (me *cfile) compileTernary(n *node) *codeblock {
	code := ""
	code += me.eval(n.has[0]).code()
	code += " ? "
	code += me.eval(n.has[1]).code()
	code += " : "
	code += me.eval(n.has[2]).code()
	return codeBlockOne(n, code)
}

func (me *cfile) compileAndOr(n *node) *codeblock {
	_, paren := n.attributes["parenthesis"]
	code := ""
	if paren {
		code += "("
	}
	code += me.eval(n.has[0]).code()
	if n.is == "and" {
		code += " && "
	} else {
		code += " || "
	}
	code += me.eval(n.has[1]).code()
	if paren {
		code += ")"
	}
	return codeBlockOne(n, code)
}

func (me *cfile) initStatic(n *node) string {
	left := n.has[0]
	right := n.has[1]
	right.attributes["global"] = "true"

	declareCode := me.declare(left)
	rightCode := me.eval(right)
	setSign := me.maybeLet(rightCode.code(), right.attributes)

	if setSign == "" {
		return ""
	}

	code := fmc(1) + declareCode + setSign + rightCode.code() + ";\n"
	return code
}

func (me *cfile) compileAssign(n *node) string {
	left := n.has[0]
	right := n.has[1]
	if _, ok := left.attributes["mutable"]; ok {
		right.attributes["mutable"] = "true"
	}
	code := ""
	_, paren := n.attributes["parenthesis"]
	if paren {
		code += "("
	}
	declare := me.declare(left)
	value := me.eval(right)

	preCode := value.precode()
	postCode := value.pop()

	code += preCode + me.maybeFmc(preCode, me.depth) + declare + me.maybeLet(postCode, right.attributes) + postCode
	if paren {
		code += ")"
	}
	return code
}

func (me *cfile) assignmentUpdate(n *node) string {
	left := me.eval(n.has[0])
	right := me.eval(n.has[1])
	return left.code() + " " + n.is + " " + right.code()
}

func (me *cfile) free(name string) string {
	return "free(" + name + ");"
}

func (me *cfile) block(n *node) *codeblock {
	me.depth++
	expressions := n.has
	code := ""
	for ix, expr := range expressions {
		c := me.eval(expr)
		if c.code() != "" {
			if ix > 0 {
				code += "\n"
			}
			code += me.maybeFmc(c.code(), me.depth) + c.code() + me.maybeColon(c.code())
		}
	}
	me.depth--
	return codeBlockOne(n, code)
}
