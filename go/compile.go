package main

import (
	"fmt"
	"strconv"
	"strings"
)

func (me *hmfile) generateC(folder, name, libDir string) string {
	cfile := me.cFileInit()

	guard := me.defNameSpace(name)
	cfile.headPrefix += "#ifndef " + guard + "\n"
	cfile.headPrefix += "#define " + guard + "\n\n"

	cfile.headIncludeSection += "#include <stdio.h>\n"
	cfile.headIncludeSection += "#include <stdlib.h>\n"
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

	if len(me.statics) > 0 {
		for _, s := range me.statics {
			decl, impl := cfile.assignStatic(s)
			cfile.headExternSection += decl
			code += impl
		}
		cfile.headExternSection += "\n"
		code += "\n"
	}

	// TODO init func
	cfile.headFuncSection += "void " + me.funcNameSpace("init") + "();\n"
	code += "void " + me.funcNameSpace("init") + "() {\n\n}\n\n"

	for _, f := range me.functionOrder {
		if f == "main" {
			cfile.mainc(me.functions[f])
		} else {
			cfile.function(f, me.functions[f])
		}
	}

	fmt.Println("=== end C ===")

	fileCode := folder + "/" + name + ".c"
	create(fileCode, code+strings.Join(cfile.codeFn, ""))

	cfile.headSuffix += "\n#endif\n"
	create(folder+"/"+name+".h", cfile.head())

	return fileCode
}

func (me *cfile) eval(n *node) *cnode {
	op := n.is
	if op == "=" {
		code := me.assingment(n)
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "+=" || op == "-=" || op == "*=" || op == "/=" || op == "&=" || op == "|=" || op == "^=" || op == "<<=" || op == ">>=" {
		code := me.assignmentUpdate(n)
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "new" {
		cn := me.allocClass(n)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "enum" {
		data := me.hmfile.typeToVarData(n.typed)
		code := me.allocEnum(data.module, data.typed, n)
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "call" {
		data := me.hmfile.typeToVarData(n.value)
		code := me.call(data.module, data.typed, n.has)
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "concat" {
		size := len(n.has)
		code := "hmlib_concat_varg(" + strconv.Itoa(size)
		for _, snode := range n.has {
			code += ", " + me.eval(snode).code
		}
		code += ")"
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "+sign" {
		return me.compilePrefixPos(n)
	}
	if op == "-sign" {
		return me.compilePrefixNeg(n)
	}
	if op == "+" || op == "-" || op == "*" || op == "/" || op == "&" || op == "|" || op == "^" || op == "<<" || op == ">>" {
		return me.compileBinaryOp(n)
	}
	if op == "tuple-index" {
		return me.compileTupleIndex(n)
	}
	if op == "member-variable" {
		return me.compileMemberVariable(n)
	}
	if op == "variable" {
		return me.compileVariable(n)
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
		code := me.allocArray(n)
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
		return me.compileEqual(n)
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
	if op == "and" || op == "or" {
		return me.compileAndOr(n)
	}
	if op == "block" {
		return me.block(n)
	}
	if op == "break" {
		cn := codeNode(n.is, n.value, n.typed, "break")
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "continue" {
		cn := codeNode(n.is, n.value, n.typed, "continue")
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "goto" {
		cn := codeNode(n.is, "", "", "goto "+n.value)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "label" {
		cn := codeNode(n.is, "", "", n.value+":")
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "pass" {
		cn := codeNode(n.is, "", "", "")
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "match" {
		return me.compileMatch(n)
	}
	if op == "for" {
		return me.compileFor(n)
	}
	if op == "if" {
		return me.compileIf(n)
	}
	if op == "string" {
		cn := codeNode(n.is, n.value, n.typed, "\""+n.value+"\"")
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "none" {
		return me.compileNone(n)
	}
	if _, ok := primitives[op]; ok {
		cn := codeNode(n.is, n.value, n.typed, n.value)
		fmt.Println(cn.string(0))
		return cn
	}
	panic("eval unknown operation " + n.string(0))
}

func (me *cfile) compilePrefixPos(n *node) *cnode {
	code := me.eval(n.has[0]).code
	cn := codeNode(n.is, n.value, n.typed, code)
	fmt.Println(cn.string(0))
	return cn
}

func (me *cfile) compilePrefixNeg(n *node) *cnode {
	code := "-" + me.eval(n.has[0]).code
	cn := codeNode(n.is, n.value, n.typed, code)
	fmt.Println(cn.string(0))
	return cn
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
	cn := codeNode(n.is, n.value, n.typed, code)
	fmt.Println(cn.string(0))
	return cn
}

func (me *cfile) compileTupleIndex(n *node) *cnode {
	dotIndexStr := n.value
	root := me.eval(n.has[0])
	data := me.hmfile.typeToVarData(root.typed)
	_, un, _ := data.checkIsEnum()
	code := root.code + "->"
	if len(un.types) == 1 {
		code += un.name
	} else {
		code += un.name + ".var" + dotIndexStr
	}
	cn := codeNode(n.is, n.value, n.typed, code)
	fmt.Println(cn.string(0))
	return cn
}

func (me *cfile) compileMemberVariable(n *node) *cnode {
	root := n.has[0]
	code := n.value
	for {
		if root.is == "root-variable" {
			data := me.hmfile.typeToVarData(root.value)
			var vr *variable
			var cname string
			if data.module == me.hmfile {
				vr = me.getvar(data.typed)
				cname = vr.cName
			} else {
				vr = data.module.getStatic(data.typed)
				cname = data.module.varNameSpace(data.typed)
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
			if code[0] == '[' {
				code = root.value + code
			} else {
				code = root.value + "->" + code
			}
			root = root.has[0]
		} else {
			panic("missing member variable")
		}
	}
	cn := codeNode(n.is, n.value, n.typed, code)
	fmt.Println(cn.string(0))
	return cn
}

func (me *cfile) compileVariable(n *node) *cnode {
	data := me.hmfile.typeToVarData(n.value)
	var v *variable
	var cname string
	if data.module == me.hmfile {
		v = me.getvar(data.typed)
		cname = v.cName
	} else {
		v = data.module.getStatic(data.typed)
		cname = data.module.varNameSpace(data.typed)
	}
	if v == nil {
		panic("unknown variable " + data.typed)
	}
	cn := codeNode(n.is, n.value, n.typed, cname)
	fmt.Println(cn.string(0))
	return cn
}

func (me *cfile) compileNone(n *node) *cnode {
	code := "NULL"
	cn := codeNode(n.is, n.value, n.typed, code)
	fmt.Println(cn.string(0))
	return cn
}

func (me *cfile) compileMatch(n *node) *cnode {

	using := n.has[0]
	match := me.eval(using)

	if strings.HasPrefix(match.typed, "maybe") {
		return me.compileNullCheck(match, n)
	}

	code := ""
	test := match.code
	isEnum := false
	var enumNameSpace string

	if using.is == "variable" {
		var baseEnum *enum
		baseEnum, isEnum = me.hmfile.enums[using.typed]
		if isEnum {
			enumNameSpace = me.hmfile.enumNameSpace(using.typed)
			if !baseEnum.simple {
				test = using.value + "->type"
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
	cn := codeNode(n.is, n.value, n.typed, code)
	fmt.Println(cn.string(0))
	return cn
}

func (me *cfile) compileNullCheck(match *cnode, n *node) *cnode {
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

	code := ""

	if ifNull != "" && ifNotNull != "" {
		code = "if (" + match.code + " == NULL) {\n"
		code += me.maybeFmc(ifNull, me.depth+1) + ifNull + me.maybeColon(ifNull)
		code += "\n" + fmc(me.depth) + "} else {\n"
		code += me.maybeFmc(ifNotNull, me.depth+1) + ifNotNull + me.maybeColon(ifNotNull)
		code += "\n" + fmc(me.depth) + "}"

	} else if ifNull != "" {
		code = "if (" + match.code + " == NULL) {\n"
		code += me.maybeFmc(ifNull, me.depth+1) + ifNull + me.maybeColon(ifNull)
		code += "\n" + fmc(me.depth) + "}"

	} else {
		code = "if (" + match.code + " != NULL) {\n"
		code += me.maybeFmc(ifNotNull, me.depth+1) + ifNotNull + me.maybeColon(ifNotNull)
		code += "\n" + fmc(me.depth) + "}"

	}

	cn := codeNode(n.is, n.value, n.typed, code)
	fmt.Println(cn.string(0))
	return cn
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
	cn := codeNode(n.is, n.value, n.typed, code)
	fmt.Println(cn.string(0))
	return cn
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
	cn := codeNode(n.is, n.value, n.typed, code)
	fmt.Println(cn.string(0))
	return cn
}

func (me *cfile) compileIf(n *node) *cnode {
	hsize := len(n.has)
	code := "if (" + me.eval(n.has[0]).code + ") {\n"
	code += me.eval(n.has[1]).code
	code += "\n" + fmc(me.depth) + "}"
	ix := 2
	for ix < hsize && n.has[ix].is == "elif" {
		code += " else if (" + me.eval(n.has[ix].has[0]).code + ") {\n"
		code += me.eval(n.has[ix].has[1]).code
		code += "\n" + fmc(me.depth) + "}"
		ix++
	}
	if ix >= 2 && ix < hsize && n.has[ix].is == "block" {
		code += " else {\n"
		code += me.eval(n.has[ix]).code
		code += "\n" + fmc(me.depth) + "}"
	}
	cn := codeNode(n.is, n.value, n.typed, code)
	fmt.Println(cn.string(0))
	return cn
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
	code += me.eval(n.has[ix]).code
	code += "\n" + fmc(me.depth) + "}"
	cn := codeNode(n.is, n.value, n.typed, code)
	fmt.Println(cn.string(0))
	return cn
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
	size := me.eval(n.has[0])
	if _, ok := n.attributes["no-malloc"]; ok {
		return "[" + size.code + "]"
	}
	mtype := me.hmfile.typeToVarData(n.typed).typeSig()
	return "malloc((" + size.code + ") * sizeof(" + mtype + "))"
}

func (me *cfile) declare(n *node) string {
	code := ""
	name := n.value
	if me.getvar(name) == nil {
		malloc := true
		if _, ok := n.attributes["no-malloc"]; ok {
			malloc = false
		}
		mutable := false
		if _, ok := n.attributes["mutable"]; ok {
			mutable = true
		}
		if malloc {
			data := me.hmfile.typeToVarData(n.typed)
			me.scope.variables[name] = me.hmfile.varInit(data.full, name, mutable, malloc)
			codesig := fmtassignspace(data.typeSig())
			if mutable {
				code = codesig
			} else if data.postfixConst() {
				code += codesig + "const "
			} else {
				code += "const " + codesig
			}
		} else {
			data := me.hmfile.typeToVarData(n.typed)
			newVar := me.hmfile.varInit(data.full, name, mutable, malloc)
			newVar.cName = data.module.varNameSpace(name)
			me.scope.variables[name] = newVar
			codesig := fmtassignspace(data.noMallocTypeSig())
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
	right.attributes["no-malloc"] = "true"
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
	if _, ok := left.attributes["mutable"]; ok {
		right.attributes["no-malloc"] = "true"
	}
	if left.is == "variable" {
		code = me.declare(left)
	}
	rightCode := me.eval(right).code
	code += me.eval(left).code + me.maybeLet(rightCode) + rightCode
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
	code += fmc(1) + typeOf + "var = malloc(sizeof(" + unionName + "));\n"
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

func (me *cfile) defineEnum(enum *enum) {
	fmt.Println("define enum \"" + enum.name + "\"")

	impl, hmBaseEnumName := me.hmfile.enumMaybeImplNameSpace(enum.name)
	if !impl {
		me.headTypeDefSection += "typedef enum " + hmBaseEnumName + " " + hmBaseEnumName + ";\n"
		code := "enum " + hmBaseEnumName + " {\n"
		for ix, enumUnion := range enum.typesOrder {
			if ix == 0 {
				code += fmc(1) + me.hmfile.enumTypeName(hmBaseEnumName, enumUnion.name)
			} else {
				code += ",\n" + fmc(1) + me.hmfile.enumTypeName(hmBaseEnumName, enumUnion.name)
			}
		}
		code += "\n};\n\n"
		me.headTypesSection += code
	}

	if enum.simple || len(enum.generics) > 0 {
		return
	}

	code := ""
	hmBaseUnionName := me.hmfile.unionNameSpace(enum.name)
	me.headTypeDefSection += "typedef struct " + hmBaseUnionName + " " + hmBaseUnionName + ";\n"
	code += "struct " + hmBaseUnionName + " {\n"
	code += fmc(1) + hmBaseEnumName + " type;\n"
	code += fmc(1) + "union {\n"
	for _, enumUnion := range enum.typesOrder {
		me.generateUnionFn(enum, enumUnion)
		num := len(enumUnion.types)
		if num == 1 {
			typed := enumUnion.types[0]
			code += fmc(2) + fmtassignspace(typed.typeSig()) + enumUnion.name + ";\n"
		} else if num != 0 {
			code += fmc(2) + "struct {\n"
			for ix, typed := range enumUnion.types {
				code += fmc(3) + fmtassignspace(typed.typeSig()) + "var" + strconv.Itoa(ix) + ";\n"
			}
			code += fmc(2) + "} " + enumUnion.name + ";\n"
		}
	}
	code += fmc(1) + "};\n"
	code += "};\n\n"
	me.headTypesSection += code
}

func (me *cfile) defineClass(class *class) {
	if len(class.generics) > 0 {
		return
	}
	hmName := me.hmfile.classNameSpace(class.name)
	me.headTypeDefSection += "typedef struct " + hmName + " " + hmName + ";\n"
	code := "struct " + hmName + " {\n"
	for _, name := range class.variableOrder {
		field := class.variables[name]
		code += fmc(1) + fmtassignspace(field.vdat.typeSig()) + field.name + ";\n"
	}
	code += "};\n\n"
	me.headTypesSection += code
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
	if code == "" || code[0] == ' ' {
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
	cn := codeNode(n.is, n.value, n.typed, code)
	fmt.Println(cn.string(0))
	me.depth--
	return cn
}

func (me *cfile) mainc(fn *function) {
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
		codeblock += fmc(me.depth) + c.code + me.maybeColon(c.code) + "\n"
	}
	if !returns {
		codeblock += fmc(me.depth) + "return 0;\n"
	}
	me.popScope()
	code := ""
	code += "int main() {\n"
	code += codeblock
	code += "}\n"

	me.headFuncSection += "int main();\n"
	me.codeFn = append(me.codeFn, code)
}

func (me *cfile) function(name string, fn *function) {
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
		if c.code != "" {
			block += fmc(me.depth) + c.code + me.maybeColon(c.code) + "\n"
		}
	}
	me.popScope()
	code := ""
	code += fmtassignspace(fn.typed.typeSig()) + me.hmfile.funcNameSpace(name) + "("
	for ix, arg := range args {
		if ix > 0 {
			code += ", "
		}
		codesig := fmtassignspace(arg.vdat.typeSig())
		if arg.vdat.postfixConst() {
			code += codesig + "const "
		} else {
			code += "const " + codesig
		}
		code += arg.name
	}
	head := code + ");\n"
	code += ") {\n"
	code += block
	code += "}\n\n"

	me.headFuncSection += head
	me.codeFn = append(me.codeFn, code)
}

func (me *cfile) call(module *hmfile, name string, parameters []*node) string {
	code := me.builtin(name, parameters)
	if code != "" {
		return code
	}
	code = module.funcNameSpace(name) + "("
	for ix, parameter := range parameters {
		if ix > 0 {
			code += ", "
		}
		code += me.eval(parameter).code
	}
	code += ")"
	return code
}

func (me *cfile) builtin(name string, parameters []*node) string {
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
	if name == "string" {
		param := me.eval(parameters[0])
		if param.typed == "string" {
			panic("redundant string cast")

		} else if param.typed == "int" {
			return "hmlib_int_to_string(" + param.code + ")"

		} else if param.typed == "float" {
			return "hmlib_float_to_string(" + param.code + ")"

		} else if param.typed == "bool" {
			return "(" + param.code + " ? \"true\" : \"false\")"
		}
		panic("argument for string cast was " + param.string(0))
	}
	if name == "int" {
		param := me.eval(parameters[0])
		if param.typed == "int" {
			panic("redundant int cast")

		} else if param.typed == "float" {
			return "((int) " + param.code + ")"

		} else if param.typed == "string" {
			return "hmlib_string_to_int(" + param.code + ")"
		}
		panic("argument for int cast was " + param.string(0))
	}
	if name == "float" {
		param := me.eval(parameters[0])
		if param.typed == "float" {
			panic("redundant float cast")

		} else if param.typed == "int" {
			return "((float) " + param.code + ")"

		} else if param.typed == "string" {
			return "hmlib_string_to_float(" + param.code + ")"
		}
		panic("argument for float cast was " + param.string(0))
	}
	return ""
}
