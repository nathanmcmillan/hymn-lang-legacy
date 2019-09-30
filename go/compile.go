package main

import (
	"fmt"
	"strconv"
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
	cfile.headIncludeSection += "#include \"" + libDir + "/hmlib_strings.h\"\n"
	cfile.hmfile.program.sources["hmlib_strings.c"] = libDir + "/hmlib_strings.c"
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
			cfile.defineMain(me.functions[f])
		} else {
			cfile.defineFunction(f, me.functions[f])
		}
	}

	fmt.Println("=== end C ===")

	fileCode := folder + "/" + name + ".c"

	write(fileCode, code+strings.Join(cfile.codeFn, ""))

	cfile.headSuffix += "\n#endif\n"
	write(folder+"/"+name+".h", cfile.head())

	return fileCode
}

func (me *cfile) hintEval(n *node, hint *varData) *cnode {
	op := n.is
	if op == "=" || op == ":=" {
		code := me.assingment(n)
		return codeNode(n, code)
	}
	if op == "+=" || op == "-=" || op == "*=" || op == "/=" || op == "%=" || op == "&=" || op == "|=" || op == "^=" || op == "<<=" || op == ">>=" {
		code := me.assignmentUpdate(n)
		return codeNode(n, code)
	}
	if op == "new" {
		return me.allocClass(n)
	}
	if op == "enum" {
		code := me.allocEnum(n)
		return codeNode(n, code)
	}
	if op == "cast" {
		return me.compileCast(n)
	}
	if op == "concat" {
		size := len(n.has)
		code := ""
		if size == 2 {
			code += "hmlib_concat("
			code += me.eval(n.has[0]).code
			code += ", "
			code += me.eval(n.has[1]).code
		} else {
			code += "hmlib_concat_varg(" + strconv.Itoa(size)
			for _, snode := range n.has {
				code += ", " + me.eval(snode).code
			}
		}
		code += ")"
		return codeNode(n, code)
	}
	if op == "+sign" {
		return me.compilePrefixPos(n)
	}
	if op == "-sign" {
		return me.compilePrefixNeg(n)
	}
	if op == "+" || op == "-" || op == "*" || op == "/" || op == "%" || op == "&" || op == "|" || op == "^" || op == "<<" || op == ">>" {
		return me.compileBinaryOp(n)
	}
	if op == "fn-ptr" {
		return me.compileFnPtr(n, hint)
	}
	if op == "variable" {
		return me.compileVariable(n, hint)
	}
	if op == "root-variable" {
		v := me.getvar(n.idata.name)
		return codeNode(n, v.cName)
	}
	if op == "array-member" {
		index := me.eval(n.has[0])
		root := me.eval(n.has[1])
		code := root.code + "[" + index.code + "]"
		return codeNode(n, code)
	}
	if op == "member-variable" {
		return me.compileMemberVariable(n)
	}
	if op == "tuple-index" {
		return me.compileTupleIndex(n)
	}
	if op == "call" {
		return me.compileCall(n)
	}
	if op == "array" {
		code := me.allocArray(n)
		return codeNode(n, code)
	}
	if op == "return" {
		in := me.eval(n.has[0])
		code := "return " + in.code
		cn := codeNode(n, code)
		cn.push(in)
		return cn
	}
	if op == "boolexpr" {
		code := me.eval(n.has[0]).code
		return codeNode(n, code)
	}
	if op == "equal" {
		return me.compileEqual(n)
	}
	if op == "not" {
		code := "!" + me.eval(n.has[0]).code
		return codeNode(n, code)
	}
	if op == "not-equal" {
		code := me.eval(n.has[0]).code
		code += " != "
		code += me.eval(n.has[1]).code
		return codeNode(n, code)
	}
	if op == ">" || op == ">=" || op == "<" || op == "<=" {
		code := me.eval(n.has[0]).code
		code += " " + op + " "
		code += me.eval(n.has[1]).code
		return codeNode(n, code)
	}
	if op == "?" {
		return me.compileTernary(n)
	}
	if op == "and" || op == "or" {
		return me.compileAndOr(n)
	}
	if op == "block" {
		return me.block(n)
	}
	if op == "break" {
		return codeNode(n, "break")
	}
	if op == "continue" {
		return codeNode(n, "continue")
	}
	if op == "goto" {
		return codeNode(n, "goto "+n.value)
	}
	if op == "label" {
		return codeNode(n, n.value+":")
	}
	if op == "pass" {
		return codeNode(n, "")
	}
	if op == "match" {
		return me.compileMatch(n)
	}
	if op == "is" {
		return me.compileIs(n)
	}
	if op == "for" {
		return me.compileFor(n)
	}
	if op == "if" {
		return me.compileIf(n)
	}
	if op == TokenRawString {
		return me.compileRawString(n)
	}
	if op == TokenString {
		return me.compileString(n)
	}
	if op == "none" {
		return me.compileNone(n)
	}
	if _, ok := primitives[op]; ok {
		return codeNode(n, n.value)
	}
	panic("eval unknown operation " + n.string(0))
}

func (me *cfile) eval(n *node) *cnode {
	return me.hintEval(n, nil)
}

func (me *cfile) compilePrefixPos(n *node) *cnode {
	code := me.eval(n.has[0]).code
	return codeNode(n, code)
}

func (me *cfile) compilePrefixNeg(n *node) *cnode {
	code := "-" + me.eval(n.has[0]).code
	return codeNode(n, code)
}

func (me *cfile) compileCast(n *node) *cnode {
	typ := getCName(n.vdata.full)
	code := "(" + typ + ")" + me.eval(n.has[0]).code
	return codeNode(n, code)
}

func (me *cfile) compileBinaryOp(n *node) *cnode {
	_, paren := n.attributes["parenthesis"]
	code := ""
	if paren {
		code += "("
	}
	code += me.eval(n.has[0]).code
	code += " " + n.is + " "
	code += me.eval(n.has[1]).code
	if paren {
		code += ")"
	}
	return codeNode(n, code)
}

func (me *cfile) compileTupleIndex(n *node) *cnode {
	dotIndexStr := n.value
	root := me.eval(n.has[0])
	data := root.vdata
	_, un, _ := data.checkIsEnum()
	code := root.code + "->"
	if len(un.types) == 1 {
		code += un.name
	} else {
		code += un.name + ".var" + dotIndexStr
	}
	return codeNode(n, code)
}

func (me *cfile) compileMemberVariable(n *node) *cnode {
	head := n.has[0]
	code := n.idata.name
	for {
		if head.is == "root-variable" {
			data := head.asVar()
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
				code = cname + data.memPtr() + code
			}
			break
		} else if head.is == "array-member" {
			index := me.eval(head.has[0])
			code = "[" + index.code + "]" + "->" + code
			head = head.has[1]
		} else if head.is == "member-variable" {
			if code[0] == '[' {
				code = head.idata.name + code
			} else {
				code = head.idata.name + head.asVar().memPtr() + code
			}
			head = head.has[0]
		} else {
			panic("missing member variable")
		}
	}
	return codeNode(n, code)
}

func (me *cfile) compileFnPtr(n *node, hint *varData) *cnode {
	code := ""
	fn := n.fn
	code += "&" + fn.module.funcNameSpace(fn.name)
	return codeNode(n, code)
}

func (me *cfile) compileVariable(n *node, hint *varData) *cnode {
	code := ""
	if n.idata.module == me.hmfile {
		v := me.getvar(n.idata.name)
		vd := v.vdat
		code = v.cName
		if hint != nil && hint.isptr && !vd.isptr {
			code = "&" + code
		}
	} else {
		code = n.idata.module.varNameSpace(n.idata.name)
	}
	return codeNode(n, code)
}

func (me *cfile) compileRawString(n *node) *cnode {
	return codeNode(n, "\""+n.value+"\"")
}

func (me *cfile) compileString(n *node) *cnode {
	code := "hmlib_string_init(\"" + n.value + "\")"
	return codeNode(n, code)
}

func (me *cfile) compileNone(n *node) *cnode {
	code := "NULL"
	return codeNode(n, code)
}

func (me *cfile) compileIs(n *node) *cnode {
	code := ""
	code += me.walrusMatch(n)
	using := n.has[0]
	match := me.eval(using)

	if match.vdata.maybe {
		caseOf := n.has[1]
		if caseOf.is == "some" {
			code += match.code + " != NULL"
		} else if caseOf.is == "none" {
			code += match.code + " == NULL"
		}
		return codeNode(n, code)
	}

	baseEnum, _, _ := using.vdata.checkIsEnum()
	if baseEnum.simple {
		code += match.code
	} else {
		code += using.idata.name + "->type"
	}

	code += " == "

	caseOf := n.has[1]
	if caseOf.is == "match-enum" {
		matchBaseEnum, matchBaseUn, _ := caseOf.vdata.checkIsEnum()
		enumNameSpace := me.hmfile.enumNameSpace(matchBaseEnum.name)
		code += me.hmfile.enumTypeName(enumNameSpace, matchBaseUn.name)
	} else {
		compare := me.eval(caseOf)
		compareEnum, _, _ := compare.vdata.checkIsEnum()
		code += compare.code
		if !compareEnum.simple {
			code += "->type"
		}
	}

	return codeNode(n, code)
}

func (me *cfile) compileMatch(n *node) *cnode {
	code := ""
	code += me.walrusMatch(n)
	using := n.has[0]
	match := me.eval(using)

	if match.vdata.maybe {
		return me.compileNullCheck(match, n, code)
	}

	test := match.code
	isEnum := false
	var enumNameSpace string

	if using.is == "variable" {
		if baseEnum, _, ok := using.vdata.checkIsEnum(); ok {
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
		thenBlock := me.eval(thenDo).code
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
	return codeNode(n, code)
}

func (me *cfile) compileNullCheck(match *cnode, n *node, code string) *cnode {
	ifNull := ""
	ifNotNull := ""
	ix := 1
	size := len(n.has)
	for ix < size {
		caseOf := n.has[ix]
		thenDo := n.has[ix+1]
		thenBlock := me.eval(thenDo).code
		if caseOf.is == "some" {
			ifNotNull = thenBlock
		} else if caseOf.is == "none" {
			ifNull = thenBlock
		}
		ix += 2
	}

	if ifNull != "" && ifNotNull != "" {
		code += "if (" + match.code + " == NULL) {\n"
		code += me.maybeFmc(ifNull, me.depth+1) + ifNull + me.maybeColon(ifNull)
		code += "\n" + fmc(me.depth) + "} else {\n"
		code += me.maybeFmc(ifNotNull, me.depth+1) + ifNotNull + me.maybeColon(ifNotNull)
		code += "\n" + fmc(me.depth) + "}"

	} else if ifNull != "" {
		code += "if (" + match.code + " == NULL) {\n"
		code += me.maybeFmc(ifNull, me.depth+1) + ifNull + me.maybeColon(ifNull)
		code += "\n" + fmc(me.depth) + "}"

	} else {
		code += "if (" + match.code + " != NULL) {\n"
		code += me.maybeFmc(ifNotNull, me.depth+1) + ifNotNull + me.maybeColon(ifNotNull)
		code += "\n" + fmc(me.depth) + "}"

	}

	return codeNode(n, code)
}

func (me *cfile) compileEqual(n *node) *cnode {
	_, paren := n.attributes["parenthesis"]
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
	return codeNode(n, code)
}

func (me *cfile) compileTernary(n *node) *cnode {
	code := ""
	code += me.eval(n.has[0]).code
	code += " ? "
	code += me.eval(n.has[1]).code
	code += " : "
	code += me.eval(n.has[2]).code
	return codeNode(n, code)
}

func (me *cfile) compileAndOr(n *node) *cnode {
	_, paren := n.attributes["parenthesis"]
	code := ""
	if paren {
		code += "("
	}
	code += me.eval(n.has[0]).code
	if n.is == "and" {
		code += " && "
	} else {
		code += " || "
	}
	code += me.eval(n.has[1]).code
	if paren {
		code += ")"
	}
	return codeNode(n, code)
}

func (me *cfile) compileIf(n *node) *cnode {
	hsize := len(n.has)
	code := ""
	code += me.walrusIf(n)
	code += "if (" + me.eval(n.has[0]).code + ") {\n"
	c := me.eval(n.has[1]).code
	code += me.maybeFmc(c, me.depth+1) + c + me.maybeColon(c)
	code += "\n" + fmc(me.depth) + "}"
	ix := 2
	for ix < hsize && n.has[ix].is == "elif" {
		code += " else if (" + me.eval(n.has[ix].has[0]).code + ") {\n"
		c := me.eval(n.has[ix].has[1]).code
		code += me.maybeFmc(c, me.depth+1) + c + me.maybeColon(c)
		code += "\n" + fmc(me.depth) + "}"
		ix++
	}
	if ix >= 2 && ix < hsize {
		code += " else {\n"
		c := me.eval(n.has[ix]).code
		code += me.maybeFmc(c, me.depth+1) + c + me.maybeColon(c)
		code += "\n" + fmc(me.depth) + "}"
	}
	return codeNode(n, code)
}

func (me *cfile) compileFor(n *node) *cnode {
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
		vinit := me.assingment(vset)
		condition := me.eval(n.has[1]).code
		inc := me.assignmentUpdate(n.has[2])
		code += "for (" + vinit + "; " + condition + "; " + inc + ") {\n"
	} else if size > 1 {
		ix++
		code += me.walrusLoop(n)
		code += "while (" + me.eval(n.has[0]).code + ") {\n"
	} else {
		code += "while (true) {\n"
	}
	code += me.eval(n.has[ix]).code
	code += "\n" + fmc(me.depth) + "}"
	return codeNode(n, code)
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

func (me *cfile) allocArray(n *node) string {
	size := "0"
	if len(n.has) > 0 {
		size = me.eval(n.has[0]).code
	}
	if _, ok := n.attributes["no-malloc"]; ok {
		return "[" + size + "]"
	}
	mtype := n.asVar().typeSig()
	return "malloc((" + size + ") * sizeof(" + mtype + "))"
}

func (me *cfile) declare(n *node) string {
	if n.is != "variable" {
		return me.eval(n).code
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
		if _, ok := n.attributes["use-stack"]; ok {
			useStack = true
			malloc = false
		}
		mutable := false
		if _, ok := n.attributes["mutable"]; ok {
			mutable = true
		}
		data := n.vdata
		newVar := me.hmfile.varInitFromData(data, name, mutable, malloc)
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

func (me *cfile) maybeLet(code string, attributes map[string]string) string {
	if code == "" || strings.HasPrefix(code, "[") {
		return ""
	}
	if _, ok := attributes["use-stack"]; ok {
		return ""
	}
	return " = "
}

func (me *cfile) declareStatic(n *node) string {
	left := n.has[0]
	right := n.has[1]
	right.attributes["no-malloc"] = "true"

	declareCode := me.declare(left)
	rightCode := me.eval(right)
	setSign := me.maybeLet(rightCode.code, right.attributes)

	head := "extern " + declareCode
	if setSign == "" {
		head += rightCode.code
	}
	head += ";\n"
	me.headExternSection += head

	if setSign == "" {
		return declareCode + setSign + rightCode.code + ";\n"
	}
	return declareCode + ";\n"
}

func (me *cfile) initStatic(n *node) string {
	left := n.has[0]
	right := n.has[1]
	right.attributes["no-malloc"] = "true"

	declareCode := me.declare(left)
	rightCode := me.eval(right)
	setSign := me.maybeLet(rightCode.code, right.attributes)

	if setSign == "" {
		return ""
	}

	code := fmc(1) + declareCode + setSign + rightCode.code + ";\n"
	return code
}

func (me *cfile) assingment(n *node) string {
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
	rightCode := me.eval(right).code
	code += me.declare(left) + me.maybeLet(rightCode, right.attributes) + rightCode
	if paren {
		code += ")"
	}
	return code
}

func (me *cfile) assignmentUpdate(n *node) string {
	left := me.eval(n.has[0])
	right := me.eval(n.has[1])
	return left.code + " " + n.is + " " + right.code
}

func (me *cfile) free(name string) string {
	return "free(" + name + ");"
}

func (me *cfile) generateUnionFn(en *enum, un *union) {
	_, enumName := me.hmfile.enumMaybeImplNameSpace(en.name)
	unionName := me.hmfile.unionNameSpace(en.name)
	fnName := me.hmfile.unionFnNameSpace(en, un)
	typeOf := fmtassignspace(en.typeSig())
	head := ""
	head += typeOf + fnName + "("
	if len(un.types) == 1 {
		unionHas := un.types[0]
		head += fmtassignspace(unionHas.typeSig()) + un.name
	} else {
		for ix := range un.types {
			if ix > 0 {
				head += ", "
			}
			unionHas := un.types[ix]
			head += fmtassignspace(unionHas.typeSig()) + un.name + strconv.Itoa(ix)
		}
	}
	head += ")"
	code := head
	head += ";\n"
	code += " {\n"
	code += fmc(1) + typeOf + "const var = malloc(sizeof(" + unionName + "));\n"
	code += fmc(1) + "var->type = " + me.hmfile.enumTypeName(enumName, un.name)
	if len(un.types) == 1 {
		code += ";\n" + fmc(1) + "var->" + un.name + " = " + un.name
	} else {
		for ix := range un.types {
			code += ";\n" + fmc(1) + "var->" + un.name + ".var" + strconv.Itoa(ix) + " = " + un.name + strconv.Itoa(ix)
		}
	}
	code += ";\n" + fmc(1) + "return var;\n"
	code += "}\n\n"
	me.headFuncSection += head
	me.codeFn = append(me.codeFn, code)
}

func (me *cfile) maybeColon(code string) string {
	size := len(code)
	if size == 0 {
		return ""
	}
	last := code[size-1]
	if last == '}' || last == ':' || last == ';' {
		return ""
	}
	return ";"
}

func (me *cfile) maybeFmc(code string, depth int) string {
	if code == "" || code[0] == spaceChar {
		return ""
	}
	return fmc(depth)
}

func (me *cfile) block(n *node) *cnode {
	me.depth++
	expressions := n.has
	code := ""
	for ix, expr := range expressions {
		c := me.eval(expr)
		if c.code != "" {
			if ix > 0 {
				code += "\n"
			}
			code += me.maybeFmc(c.code, me.depth) + c.code + me.maybeColon(c.code)
		}
	}
	me.depth--
	return codeNode(n, code)
}

func (me *cfile) compileCall(node *node) *cnode {
	fn := node.fn
	if fn == nil {
		head := node.has[0]
		sig := head.vdata.fn
		code := "(*" + me.eval(head).code + ")("
		parameters := node.has[1:len(node.has)]
		for ix, parameter := range parameters {
			if ix > 0 {
				code += ", "
			}
			arg := sig.args[ix]
			code += me.hintEval(parameter, arg.vdat).code
		}
		code += ")"
		cn := codeNode(node, code)
		fmt.Println(cn.string(0))
		return cn
	}
	module := fn.module
	name := fn.name
	parameters := node.has
	code := me.builtin(name, parameters)
	if code == "" {
		if fn.forClass != nil {
			name = nameOfClassFunc(fn.forClass.name, fn.name)
		}
		code = module.funcNameSpace(name) + "("
		for ix, parameter := range parameters {
			if ix > 0 {
				code += ", "
			}
			arg := node.fn.args[ix]
			code += me.hintEval(parameter, arg.vdat).code
		}
		code += ")"
	}

	cn := codeNode(node, code)
	fmt.Println(cn.string(0))
	return cn
}

func (me *cfile) builtin(name string, parameters []*node) string {
	fmt.Println("CHECK FOR BUILT IN FN ::", name)
	switch name {
	case libPush:
		param0 := me.eval(parameters[0])
		p := param0.asVar(me.hmfile)
		if p.checkIsArray() {
			param1 := me.eval(parameters[1])
			return "push(" + param0.code + "," + param1.code + ")"
		}
		panic("argument for push was not an array \"" + p.full + "\"")
	case libLength:
		param := me.eval(parameters[0])
		switch param.getType() {
		case TokenRawString:
			return "((int) strlen(" + param.code + "))"
		case TokenString:
			return "hmlib_string_len_int(" + param.code + ")"
		}
		p := param.asVar(me.hmfile)
		if p.checkIsArray() {
			return "todo_array_get_len(" + param.code + ")"
		}
		panic("argument for echo was " + param.string(0))
	case libOpen:
		param0 := me.eval(parameters[0])
		param1 := me.eval(parameters[1])
		return "fopen(" + param0.code + ", " + param1.code + ")"
	case libEcho:
		param := me.eval(parameters[0])
		switch param.getType() {
		case TokenString:
			return "printf(\"%s\\n\", " + param.code + ")"
		case TokenRawString:
			return "printf(\"%s\\n\", " + param.code + ")"
		case TokenInt:
			return "printf(\"%d\\n\", " + param.code + ")"
		case TokenInt8:
			return "printf(\"%\" PRId8 \"\\n\", " + param.code + ")"
		case TokenInt16:
			return "printf(\"%\" PRId16 \"\\n\", " + param.code + ")"
		case TokenInt32:
			return "printf(\"%\" PRId32 \"\\n\", " + param.code + ")"
		case TokenInt64:
			return "printf(\"%\" PRId64 \"\\n\", " + param.code + ")"
		case TokenUInt:
			return "printf(\"%u\\n\", " + param.code + ")"
		case TokenUInt8:
			return "printf(\"%\" PRId8 \"\\n\", " + param.code + ")"
		case TokenUInt16:
			return "printf(\"%\" PRId16 \"\\n\", " + param.code + ")"
		case TokenUInt32:
			return "printf(\"%\" PRId32 \"\\n\", " + param.code + ")"
		case TokenUInt64:
			return "printf(\"%\" PRId64 \"\\n\", " + param.code + ")"
		case TokenFloat:
			return "printf(\"%f\\n\", " + param.code + ")"
		case TokenFloat32:
			return "printf(\"%f\\n\", " + param.code + ")"
		case TokenFloat64:
			return "printf(\"%f\\n\", " + param.code + ")"
		case "bool":
			return "printf(\"%s\\n\", " + param.code + " ? \"true\" : \"false\")"
		case TokenLibSize:
			return "printf(\"%zu\\n\", " + param.code + ")"
		}
		panic("argument for echo was " + param.string(0))
	case libToStr:
		param := me.eval(parameters[0])
		switch param.getType() {
		case TokenString:
			panic("redundant string cast")
		case TokenInt:
			return "hmlib_int_to_string(" + param.code + ")"
		case TokenInt8:
			return "hmlib_int8_to_string(" + param.code + ")"
		case TokenInt16:
			return "hmlib_int16_to_string(" + param.code + ")"
		case TokenInt32:
			return "hmlib_int32_to_string(" + param.code + ")"
		case TokenInt64:
			return "hmlib_int64_to_string(" + param.code + ")"
		case TokenUInt:
			return "hmlib_uint_to_string(" + param.code + ")"
		case TokenUInt8:
			return "hmlib_uint8_to_string(" + param.code + ")"
		case TokenUInt16:
			return "hmlib_uint16_to_string(" + param.code + ")"
		case TokenUInt32:
			return "hmlib_uint32_to_string(" + param.code + ")"
		case TokenUInt64:
			return "hmlib_uint64_to_string(" + param.code + ")"
		case TokenFloat:
			return "hmlib_float_to_string(" + param.code + ")"
		case TokenFloat32:
			return "hmlib_float32_to_string(" + param.code + ")"
		case TokenFloat64:
			return "hmlib_float64_to_string(" + param.code + ")"
		case "bool":
			return "(" + param.code + " ? \"true\" : \"false\")"
		}
		panic("argument for string cast was " + param.string(0))
	case libToInt:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_int(" + param.code + ")"
		}
		panic("argument for conversion to int was " + param.string(0))
	case libToInt8:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_int8(" + param.code + ")"
		}
		panic("argument for conversion to int8 was " + param.string(0))
	case libToInt16:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_int16(" + param.code + ")"
		}
		panic("argument for conversion to int16 was " + param.string(0))
	case libToInt32:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_int32(" + param.code + ")"
		}
		panic("argument for conversion to int32 was " + param.string(0))
	case libToInt64:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_int64(" + param.code + ")"
		}
		panic("argument for conversion to int64 was " + param.string(0))
	case libToUInt:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_uint(" + param.code + ")"
		}
		panic("argument for conversion to uint was " + param.string(0))
	case libToUInt8:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_uint8(" + param.code + ")"
		}
		panic("argument for conversion to uint8 was " + param.string(0))
	case libToUInt16:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_uint16(" + param.code + ")"
		}
		panic("argument for conversion to uint16 was " + param.string(0))
	case libToUInt32:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_uint32(" + param.code + ")"
		}
		panic("argument for conversion to uint32 was " + param.string(0))
	case libToUInt64:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_uint64(" + param.code + ")"
		}
		panic("argument for conversion to uint64 was " + param.string(0))
	case libToFloat:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_float(" + param.code + ")"
		}
		panic("argument for conversion to float was " + param.string(0))
	case libToFloat32:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_float32(" + param.code + ")"
		}
		panic("argument for conversion to float32 was " + param.string(0))
	case libToFloat64:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_float64(" + param.code + ")"
		}
		panic("argument for conversion to float64 was " + param.string(0))
	default:
		return ""
	}
}
