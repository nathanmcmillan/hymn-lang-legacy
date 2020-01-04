package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

type subc struct {
	fname     string
	subfolder string
	base      bool
}

func (me *subc) location() string {
	return me.fname
}

func (me *hmfile) generateC(folder, name, hmlibs string) string {
	if debug {
		fmt.Println("=== " + name + " C ===")
	}

	cfile := me.cFileInit()
	guard := me.defNameSpace("", name)

	cfile.headStdIncludeSection.WriteString("#ifndef " + guard + "\n")
	cfile.headStdIncludeSection.WriteString("#define " + guard + "\n")

	cfile.headStdIncludeSection.WriteString("\n#include <stdio.h>")
	cfile.headStdIncludeSection.WriteString("\n#include <stdlib.h>")
	cfile.headStdIncludeSection.WriteString("\n#include <stdint.h>")
	cfile.headStdIncludeSection.WriteString("\n#include <inttypes.h>")
	cfile.headStdIncludeSection.WriteString("\n#include <stdbool.h>")

	if len(me.importOrder) > 0 {
		for _, iname := range me.importOrder {
			imp := me.imports[iname]
			path := imp.name + "/" + imp.name
			cfile.headReqIncludeSection.WriteString("\n#include \"" + path + ".h\"")
		}
	}

	root, _ := filepath.Abs(folder)
	filterOrder := make([]string, 0)
	filters := make(map[string]subc)

	for _, c := range me.defineOrder {
		underscore := strings.LastIndex(c, "_")
		name := c[0:underscore]
		typed := c[underscore+1:]
		subfolder := ""
		base := false
		if typed == "type" {
			if me.classes[name].doNotDefine() {
				continue
			}
		} else if typed == "enum" {
		} else {
			panic("missing type")
		}
		if strings.Index(name, "<") == -1 {
			subfolder = name
			base = true
		} else {
			subfolder = name[0:strings.Index(name, "<")]
		}
		fname := flatten(name)
		fname = strings.ReplaceAll(fname, "_", "-")
		filterOrder = append(filterOrder, name)
		s := subc{fname: fname, subfolder: subfolder, base: base}
		filters[name] = s
		cfile.headReqIncludeSection.WriteString("\n#include \"" + s.location() + ".h\"")
	}

	var code strings.Builder
	code.WriteString("#include \"" + name + ".h\"\n")

	for _, c := range me.defineOrder {
		underscore := strings.LastIndex(c, "_")
		name := c[0:underscore]
		if _, ok := filters[name]; ok {
			continue
		}
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
			code.WriteString(cfile.declareStatic(s))
		}
		code.WriteString("\n\n")

		cfile.headFuncSection.WriteString("\nvoid " + me.funcNameSpace("init") + "();")
		code.WriteString("void " + me.funcNameSpace("init") + "() {\n")
		for _, s := range me.statics {
			code.WriteString(cfile.initStatic(s))
		}
		code.WriteString("}\n")
	}

	for _, f := range me.functionOrder {
		if _, ok := filters[name]; ok {
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
		subc := filters[f]
		cfile.subC(root, folder, name, hmlibs, f, &subc, filterOrder, filters)
	}

	if debug {
		fmt.Println("=== end C ===")
	}

	fileCode := folder + "/" + name + ".c"

	write(fileCode, code.String())
	for _, cfn := range cfile.codeFn {
		fileappend(fileCode, cfn.String())
	}

	cfile.headSuffix.WriteString("\n\n#endif\n")
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
	typ, _ := getCName(n.data().full)
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
		name := n.idata.name
		v := me.getvar(name)
		if v == nil {
			module := me.hmfile
			for _, s := range module.statics {
				sname := s.has[0].idata.name
				if sname == name {
					me.defineStatic(s)
					break
				}
			}
			v = me.getvar(name)
		}
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
	me.libReq.add(HmLibString)
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

func (me *cfile) compileEqual(op string, n *node) *codeblock {
	a := me.eval(n.has[0])
	b := me.eval(n.has[1])
	code := ""
	if a.data().checkIsString() && b.data().checkIsString() {
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
	// TODO remove me?
	// _, paren := n.attributes["parenthesis"]
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

func (me *cfile) initStatic(n *node) string {
	left := n.has[0]
	right := n.has[1]
	right.attributes["global"] = "true"

	declareCode := me.compileDeclare(left)
	rightCode := me.eval(right)
	setSign := me.maybeLet(rightCode.code(), right.attributes)

	if setSign == "" {
		return ""
	}

	code := fmc(1) + declareCode + setSign + rightCode.code() + ";\n"
	return code
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
