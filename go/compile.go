package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (me *hmfile) generateC(module *hmfile) string {

	folder := module.out
	filename := fileName(module.path)
	hmlibs := module.libs

	if debug {
		fmt.Println("=== compile: " + filename + " ===")
	}

	cfile := me.cFileInit()
	cfile.master = true

	guard := me.headerFileGuard("", filename)

	cfile.headStdIncludeSection.WriteString("#ifndef " + guard + "\n")
	cfile.headStdIncludeSection.WriteString("#define " + guard + "\n")

	cfile.headStdIncludeSection.WriteString("\n#include <stdio.h>")
	cfile.headStdIncludeSection.WriteString("\n#include <stdlib.h>")
	cfile.headStdIncludeSection.WriteString("\n#include <stdint.h>")
	cfile.headStdIncludeSection.WriteString("\n#include <inttypes.h>")
	cfile.headStdIncludeSection.WriteString("\n#include <stdbool.h>")

	if len(me.importOrder) > 0 {
		for _, iname := range me.importOrder {
			importing := me.imports[iname]
			path := importing.name + "/" + importing.name
			cfile.headReqIncludeSection.WriteString("\n#include \"" + path + ".h\"")
		}
	}

	root, _ := filepath.Abs(folder)
	filterOrder := make([]string, 0)
	filters := make(map[string]string)

	for _, def := range me.defineOrder {
		var name string
		var fname string
		if def.class != nil {
			if def.class.doNotDefine {
				continue
			}
			name = def.class.name
			fname = def.class.pathLocal
		} else if def.enum != nil {
			name = def.enum.name
			fname = def.enum.pathLocal
		} else {
			panic("Missing definition")
		}
		if name == filename {
			continue
		}
		filterOrder = append(filterOrder, name)
		filters[name] = fname
		cfile.headReqIncludeSection.WriteString("\n#include \"" + fname + ".h\"")
	}

	var code strings.Builder
	code.WriteString("#include \"" + filename + ".h\"\n")

	for _, def := range me.defineOrder {
		if def.class != nil {
			if def.class.doNotDefine {
				continue
			}
			name := def.class.name
			if _, ok := filters[name]; ok {
				continue
			}
			cfile.defineClass(def.class)

		} else if def.enum != nil {
			name := def.enum.name
			if _, ok := filters[name]; ok {
				continue
			}
			cfile.defineEnum(def.enum)

		} else {
			panic("Missing definition")
		}
	}

	if len(me.statics) != 0 {
		me.needStatic = true
	}

	if me.needStatic {
		for _, s := range me.statics {
			code.WriteString(cfile.declareStatic(s))
		}
		code.WriteString("\n\n")

		cfile.headFuncSection.WriteString("\nvoid " + me.funcNameSpace("static") + "();")
		code.WriteString("void " + me.funcNameSpace("static") + "() {\n")
		cfile.depth++
		for _, s := range me.statics {
			code.WriteString(cfile.happyOut(cfile.initStatic(s)))
		}
		cfile.depth--
		code.WriteString("}\n")
	}

	for _, f := range me.functionOrder {
		if _, ok := filters[filename]; ok {
			continue
		}
		fun := me.functions[f]
		if fun.forClass != nil {
			continue
		}
		if f == "main" {
			cfile.compileMain(fun)
		} else {
			cfile.compileFunction(f, fun, false)
		}
	}

	for _, f := range filterOrder {
		cfile.subC(root, folder, filename, hmlibs, f, filters[f])
	}

	if debug {
		fmt.Println("=== compile: end ===")
	}

	fileCode := filepath.Join(folder, filename+".c")

	write(fileCode, code.String())

	if len(cfile.codeFn) > 0 {
		for _, cfn := range cfile.codeFn {
			fileappend(fileCode, cfn.String())
		}
		cfile.headSuffix.WriteString("\n")
	}

	if len(me.comments) > 0 {
		code.Reset()
		code.WriteString("\n")
		for _, comment := range me.comments {
			code.WriteString("//")
			code.WriteString(comment)
			code.WriteString("\n")
		}
		fileappend(fileCode, code.String())
	}

	cfile.headSuffix.WriteString("\n#endif\n")
	write(filepath.Join(folder, filename+".h"), cfile.head())

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
	typ := n.data().cname()
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
	_, un, _ := data.isEnum()
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
				cname = vr.cname
			} else {
				vr = data.getmodule().getStatic(head.idata.name)
				cname = head.idata.getcname()
			}
			if data.isArrayOrSlice() {
				code = cname + code
			} else {
				code = cname + data.memoryGet() + code
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
				code = head.idata.name + head.data().memoryGet() + code
			}
			head = head.has[0]
		} else {
			panic("missing member variable")
		}
	}
	return codeBlockOne(n, code)
}

func (me *cfile) compileFnPtr(n *node, hint *datatype) *codeblock {
	code := ""
	fn := n.fn
	code += "&" + fn.getcname()
	return codeBlockOne(n, code)
}

func (me *cfile) compileRootVariable(n *node, hint *datatype) *codeblock {
	v := me.getvar(n.idata.name)
	code := v.cname
	if hint != nil && hint.isPointer() && !v.data().isPointer() {
		code = "&" + code
	}
	return codeBlockOne(n, code)
}

func (me *cfile) compileVariable(n *node, hint *datatype) *codeblock {
	code := ""
	if n.idata.module == me.hmfile {
		name := n.idata.name
		v := me.getvar(name)
		if v == nil {
			module := me.hmfile
			if st, ok := module.staticScope[name]; ok {
				me.defineStatic(st)
			} else {
				panic("Could not find static variable \"" + name + "\"")
			}
			v = me.getvar(name)
		}
		code = v.cname
		if hint != nil && hint.isPointer() && !v.data().isPointer() {
			code = "&" + code
		}
	} else {
		code = n.idata.getcname()
	}
	return codeBlockOne(n, code)
}

func (me *cfile) compileRawString(n *node) *codeblock {
	return codeBlockOne(n, "\""+n.value+"\"")
}

func (me *cfile) compileString(n *node) *codeblock {
	me.libReq.add(HmLibString)
	code := "hmlib_string_init(\"" + n.value + "\")"
	return codeBlockOne(n, code)
}

func (me *cfile) compileChar(n *node) *codeblock {
	return codeBlockOne(n, n.value)
}

func (me *cfile) compileNone(n *node) *codeblock {
	code := "NULL"
	return codeBlockOne(n, code)
}

func (me *cfile) compileComment(n *node) *codeblock {
	code := "//" + n.value
	return codeBlockOne(n, code)
}

func (me *cfile) compileEqual(op string, n *node) *codeblock {
	a := me.eval(n.has[0])
	b := me.eval(n.has[1])
	code := ""
	if a.data().isString() && b.data().isString() {
		me.libReq.add(HmLibString)
		code = "hmlib_string_equal(" + a.code() + ", " + b.code() + ")"
		if op == "not-equal" {
			code = "!" + code
		}
	} else {
		_, paren := n.attributes["parenthesis"]
		if paren {
			code += "("
		}
		code += a.code()
		if op == "equal" {
			code += " == "
		} else {
			code += " != "
		}
		code += b.code()
		if paren {
			code += ")"
		}
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
	paren := true
	if n.parent != nil && n.parent.is == "if" {
		paren = false
	}
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

func (me *cfile) compileAssign(n *node) *codeblock {
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
	declare := me.compileDeclare(left)
	value := me.eval(right)
	pre := value.precode()
	post := value.pop()

	code += pre
	if n.is != ":=" {
		code += me.maybeFmc(code, me.depth)
	}
	code += declare + me.maybeLet(post, right.attributes) + post

	if paren {
		code += ")"
	}
	return codeBlockOne(n, code)
}

func (me *cfile) assignmentUpdate(n *node) string {
	left := me.eval(n.has[0])
	right := me.eval(n.has[1])
	if left.data().isString() {
		return left.code() + " = hmlib_concat(" + left.code() + ", " + right.code() + ")"
	}
	return left.code() + " " + n.is + " " + right.code()
}

func (me *cfile) free(name string) string {
	return "free(" + name + ");"
}

func (me *cfile) block(n *node) *codeblock {
	me.depth++
	expressions := n.has
	code := ""
	for _, expr := range expressions {
		e := me.eval(expr)
		code += me.happyOut(e)
	}
	me.depth--
	return codeBlockOne(n, code)
}
