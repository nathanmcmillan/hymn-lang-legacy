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

	for importName := range me.imports {
		cfile.headIncludeSection += "#include \"" + importName + ".h\"\n"
	}
	cfile.headIncludeSection += "\n"

	code := ""
	code += "#include \"" + name + ".h\"\n"
	code += "\n"

	for _, c := range me.defineOrder {
		def := strings.Split(c, "_")
		name := def[0]
		typed := def[1]
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
			if data.array {
				code = cname + code
			} else {
				fmt.Println("MEM ROOT HEAD ::", head.idata.name, "::", data.full, "?", data.isptr, data.onStack)
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
		vd := v.data()
		code = v.cName
		if hint != nil && hint.isptr && !vd.isptr {
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

func (me *cfile) compileIs(n *node) *codeblock {
	code := ""
	code += me.walrusMatch(n)
	using := n.has[0]
	match := me.eval(using)

	if match.data().maybe {
		caseOf := n.has[1]
		if caseOf.is == "some" {
			code += match.code() + " != NULL"
		} else if caseOf.is == "none" {
			code += match.code() + " == NULL"
		}
		return codeBlockOne(n, code)
	}

	baseEnum, _, _ := using.data().checkIsEnum()
	if baseEnum.simple {
		code += match.code()
	} else {
		code += using.idata.name + "->type"
	}

	code += " == "

	caseOf := n.has[1]
	if caseOf.is == "match-enum" {
		matchBaseEnum, matchBaseUn, _ := caseOf.data().checkIsEnum()
		enumNameSpace := me.hmfile.enumNameSpace(matchBaseEnum.name)
		code += me.hmfile.enumTypeName(enumNameSpace, matchBaseUn.name)
	} else {
		compare := me.eval(caseOf)
		compareEnum, _, _ := compare.data().checkIsEnum()
		code += compare.code()
		if !compareEnum.simple {
			code += "->type"
		}
	}

	return codeBlockOne(n, code)
}

func (me *cfile) compileMatch(n *node) *codeblock {
	code := ""
	code += me.walrusMatch(n)
	using := n.has[0]
	match := me.eval(using)

	if match.data().maybe {
		return me.compileNullCheck(match, n, code)
	}

	test := match.code()
	isEnum := false
	var enumNameSpace string

	if using.is == "variable" {
		if baseEnum, _, ok := using.data().checkIsEnum(); ok {
			isEnum = true
			enumNameSpace = me.hmfile.enumNameSpace(baseEnum.name)
			if !baseEnum.simple {
				test = using.idata.name + "->type"
			}
		}
	}

	code += "switch (" + test + ") {\n"
	ix := 1
	size := len(n.has)
	for ix < size {
		caseOf := n.has[ix]
		thenDo := n.has[ix+1]
		thenBlock := me.eval(thenDo).code()
		if caseOf.is == "_" {
			code += fmc(me.depth) + "default:\n"
		} else {
			if isEnum {
				code += fmc(me.depth) + "case " + me.hmfile.enumTypeName(enumNameSpace, caseOf.is) + ":\n"
			} else {
				code += fmc(me.depth) + "case " + caseOf.is + ":\n"
			}
		}
		me.depth++
		if thenBlock != "" {
			code += me.maybeFmc(thenBlock, me.depth) + thenBlock + me.maybeColon(thenBlock) + "\n"
		}
		if !strings.Contains(thenBlock, "return") {
			code += fmc(me.depth) + "break;\n"
		}
		me.depth--
		ix += 2
	}
	code += fmc(me.depth) + "}"
	return codeBlockOne(n, code)
}

func (me *cfile) compileNullCheck(match *codeblock, n *node, code string) *codeblock {
	ifNull := ""
	ifNotNull := ""
	ix := 1
	size := len(n.has)
	for ix < size {
		block := ""
		tempname := ""
		caseOf := n.has[ix]
		if caseOf.is == "some" {
			if len(caseOf.has) > 0 {
				temp := caseOf.has[0]
				tempname = temp.idata.name
				tempv := me.hmfile.varInitFromData(temp.data(), tempname, false)
				me.scope.variables[tempname] = tempv
				ref := me.eval(n.has[0]).code()
				block += fmtassignspace(temp.data().typeSig()) + tempname + " = " + ref + ";\n"
			}
		}
		c := me.eval(n.has[ix+1]).code()
		block += me.maybeFmc(c, me.depth+1) + c + me.maybeColon(c)
		if tempname != "" {
			delete(me.scope.variables, tempname)
		}
		if caseOf.is == "some" {
			ifNotNull = block
		} else if caseOf.is == "none" {
			ifNull = block
		}
		ix += 2
	}

	if ifNull != "" && ifNotNull != "" {
		code += "if (" + match.code() + " == NULL) {\n"
		code += me.maybeFmc(ifNull, me.depth+1) + ifNull + me.maybeColon(ifNull)
		code += "\n" + fmc(me.depth) + "} else {\n"
		code += me.maybeFmc(ifNotNull, me.depth+1) + ifNotNull + me.maybeColon(ifNotNull)
		code += "\n" + fmc(me.depth) + "}"

	} else if ifNull != "" {
		code += "if (" + match.code() + " == NULL) {\n"
		code += me.maybeFmc(ifNull, me.depth+1) + ifNull + me.maybeColon(ifNull)
		code += "\n" + fmc(me.depth) + "}"

	} else {
		code += "if (" + match.code() + " != NULL) {\n"
		code += me.maybeFmc(ifNotNull, me.depth+1) + ifNotNull + me.maybeColon(ifNotNull)
		code += "\n" + fmc(me.depth) + "}"

	}

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

func (me *cfile) compileIf(n *node) *codeblock {
	hsize := len(n.has)
	code := ""
	code += me.walrusIf(n)
	code += "if (" + me.eval(n.has[0]).code() + ") {\n"
	c := me.eval(n.has[1]).code()
	code += me.maybeFmc(c, me.depth+1) + c + me.maybeColon(c)
	code += "\n" + fmc(me.depth) + "}"
	ix := 2
	for ix < hsize && n.has[ix].is == "elif" {
		code += " else if (" + me.eval(n.has[ix].has[0]).code() + ") {\n"
		c := me.eval(n.has[ix].has[1]).code()
		code += me.maybeFmc(c, me.depth+1) + c + me.maybeColon(c)
		code += "\n" + fmc(me.depth) + "}"
		ix++
	}
	if ix >= 2 && ix < hsize {
		code += " else {\n"
		c := me.eval(n.has[ix]).code()
		code += me.maybeFmc(c, me.depth+1) + c + me.maybeColon(c)
		code += "\n" + fmc(me.depth) + "}"
	}
	return codeBlockOne(n, code)
}

func (me *cfile) compileFor(n *node) *codeblock {
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
		vexist := me.getvar(vobj.idata.name)
		if vexist == nil {
			code += me.declare(vobj) + ";\n" + fmc(me.depth)
		}
		vinit := me.compileAssign(vset)
		condition := me.eval(n.has[1]).code()
		inc := me.assignmentUpdate(n.has[2])
		code += "for (" + vinit + "; " + condition + "; " + inc + ") {\n"
	} else if size > 1 {
		ix++
		code += me.walrusLoop(n)
		code += "while (" + me.eval(n.has[0]).code() + ") {\n"
	} else {
		code += "while (true) {\n"
	}
	code += me.eval(n.has[ix]).code()
	code += "\n" + fmc(me.depth) + "}"
	return codeBlockOne(n, code)
}

func (me *cfile) declare(n *node) string {
	if n.is != "variable" {
		return me.eval(n).code()
	}
	if n.idata == nil {
		return ""
	}
	code := ""
	name := n.idata.name
	v := me.getvar(name)
	if v == nil {
		malloc := true
		useStack := false
		if _, ok := n.attributes["no-malloc"]; ok {
			malloc = false
		}
		if _, ok := n.attributes["stack"]; ok || n.data().onStack {
			useStack = true
			malloc = false
		}
		mutable := false
		if _, ok := n.attributes["mutable"]; ok {
			mutable = true
		}
		data := n.data()
		newVar := me.hmfile.varInitFromData(data, name, mutable)
		if useStack {
			me.scope.variables[name] = newVar
			code += fmtassignspace(data.typeSig())
			code += name

		} else if malloc {
			me.scope.variables[name] = newVar
			code += data.typeSigOf(name, mutable)

		} else {
			newVar.cName = data.module.varNameSpace(name)
			me.scope.variables[name] = newVar
			code += fmtassignspace(data.noMallocTypeSig())
			code += newVar.cName
		}
	} else {
		code += v.cName
	}

	return code
}

func (me *cfile) declareStatic(n *node) string {
	left := n.has[0]
	right := n.has[1]
	right.attributes["no-malloc"] = "true"

	declareCode := me.declare(left)
	rightCode := me.eval(right)
	setSign := me.maybeLet(rightCode.code(), right.attributes)

	head := "extern " + declareCode
	if setSign == "" {
		head += rightCode.code()
	}
	head += ";\n"
	me.headExternSection += head

	if setSign == "" {
		return declareCode + setSign + rightCode.code() + ";\n"
	}
	return declareCode + ";\n"
}

func (me *cfile) initStatic(n *node) string {
	left := n.has[0]
	right := n.has[1]
	right.attributes["no-malloc"] = "true"

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

	pre := value.preCode()
	post := value.current.code

	code += pre + me.maybeFmc(pre, me.depth) + declare + me.maybeLet(post, right.attributes) + post
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

func (me *cfile) compileCall(node *node) *codeblock {
	fn := node.fn
	if fn == nil {
		head := node.has[0]
		sig := head.data().fn
		code := "(*" + me.eval(head).code() + ")("
		parameters := node.has[1:len(node.has)]
		for ix, parameter := range parameters {
			if ix > 0 {
				code += ", "
			}
			arg := sig.args[ix]
			code += me.hintEval(parameter, arg.data()).code()
		}
		code += ")"
		return codeBlockOne(node, code)
	}
	module := fn.module
	name := fn.name
	parameters := node.has
	code := me.builtin(name, parameters)
	if code == "" {
		if fn.forClass != nil {
			name = fn.nameOfClassFunc()
		}
		code = module.funcNameSpace(name) + "("
		for ix, parameter := range parameters {
			if ix > 0 {
				code += ", "
			}
			arg := node.fn.args[ix]
			code += me.hintEval(parameter, arg.data()).code()
		}
		code += ")"
	}

	return codeBlockOne(node, code)
}
