package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (me *hmfile) generateC() string {

	filename := fileName(me.path)
	hmlibs := me.libs

	if debug {
		fmt.Println("compile>", filename)
	}

	guard := me.headerFileGuard(me.pack, "")

	cfile := me.cFileInit(guard)
	cfile.master = true

	if len(me.importOrder) > 0 {
		for _, iname := range me.importOrder {
			importing := me.imports[iname]
			path := importing.includes + "/" + importing.name
			cfile.headReqIncludeSection.WriteString("\n#include \"" + path + ".h\"")
		}
	}

	var code strings.Builder
	code.WriteString("#include \"" + filename + ".h\"\n")

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
			if name == filename {
				cfile.defineClass(def.class)
				continue
			}
			fname = def.class.pathLocal
		} else if def.enum != nil {
			name = def.enum.name
			if name == filename {
				cfile.defineEnum(def.enum)
				continue
			}
			fname = def.enum.pathLocal
		} else {
			panic("Missing definition")
		}
		filterOrder = append(filterOrder, name)
		filters[name] = fname
		cfile.addHeadSubInclude("\n#include \"" + fname + ".h\"")
	}

	if len(me.statics) != 0 {
		me.needStatic = true
	}

	if me.needStatic {
		for _, s := range me.statics {
			code.WriteString(cfile.declareStatic(s))
		}
		code.WriteString("\n\n")

		cfile.addHeadFunc("\nvoid " + me.funcNameSpace("static") + "();")
		code.WriteString("void " + me.funcNameSpace("static") + "() {\n")
		cfile.depth++
		for _, s := range me.statics {
			code.WriteString(cfile.happyOut(cfile.initStatic(s)))
		}
		cfile.depth--
		code.WriteString("}\n")
	}

	if len(me.top) > 0 {
		fn := funcInit(me, "init", nil)
		fn.returns = newdatavoid()
		fn.expressions = me.top
		me.pushFunction("init", fn)
	}

	for _, f := range me.functionOrder {
		if _, ok := filters[filename]; ok {
			continue
		}
		fun := me.functions[f]
		if fun.forClass != nil && fun.forClass.name != filename {
			continue
		}
		if f == "main" {
			if !me.program.testing {
				cfile.compileMain(fun)
			}
		} else {
			cfile.compileFunction(f, fun, false)
		}
	}

	for _, f := range filterOrder {
		cfile.subC(me.destination, filename, hmlibs, f, filters[f])
	}

	fileOut := ""

	if len(cfile.codeFn) > 0 || me.needStatic {
		fileOut = filepath.Join(me.destination, filename+".c")

		write(fileOut, code.String())

		for _, cfn := range cfile.codeFn {
			fileappend(fileOut, cfn.String())
		}
		cfile.headSuffix.WriteString("\n")

		if len(me.comments) > 0 {
			code.Reset()
			code.WriteString("\n")
			for _, comment := range me.comments {
				code.WriteString("//")
				code.WriteString(comment)
				code.WriteString("\n")
			}
			fileappend(fileOut, code.String())
		}
	}

	cfile.headSuffix.WriteString("\n#endif\n")
	write(filepath.Join(me.destination, filename+".h"), cfile.head())

	return fileOut
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

func (me *cfile) compileUnionMemberVariable(n *node) *codeblock {
	key := n.value
	root := me.eval(n.has[0])
	data := root.data()
	_, un, _ := data.isEnum()
	code := root.code() + "->"
	if un.types.size() == 1 {
		code += un.name
	} else {
		code += un.name + "." + key
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
	fn := n.fn
	code := "&" + fn.getcname()
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
