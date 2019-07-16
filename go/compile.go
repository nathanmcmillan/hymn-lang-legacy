package main

import (
	"fmt"
	"strconv"
	"strings"
)

func (me *hmfile) generateC(folder, name string) string {
	cfile := me.cFileInit()

	guard := me.defNameSpace(name)
	cfile.headPrefix += "#ifndef " + guard + "\n"
	cfile.headPrefix += "#define " + guard + "\n\n"

	cfile.headIncludeSection += "#include <stdio.h>\n"
	cfile.headIncludeSection += "#include <stdlib.h>\n"
	cfile.headIncludeSection += "#include <stdbool.h>\n"
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
		if typed == "class" {
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
	code += "void " + me.funcNameSpace("init") + "()\n{\n\n}\n\n"

	for _, f := range me.functionOrder {
		if f == "main" {
			decl, impl := cfile.mainc(me.functions[f])
			cfile.headFuncSection += decl
			code += impl
		} else {
			decl, impl := cfile.function(f, me.functions[f])
			cfile.headFuncSection += decl
			code += impl
		}
	}

	fmt.Println("=== end C ===")

	fileCode := folder + "/" + name + ".c"
	create(fileCode, code)

	cfile.headSuffix += "\n#endif\n"
	create(folder+"/"+name+".h", cfile.head())

	return fileCode
}

func (me *cfile) allocEnum(module *hmfile, typed string, n *node) string {
	enumOf := module.enums[typed]
	if enumOf.simple {
		enumBase := module.enumNameSpace(typed)
		enumType := n.value
		globalName := module.enumTypeName(enumBase, enumType)
		return globalName
	}
	if n.attribute("no-malloc") {
		return ""
	}
	enumType := n.value
	unionOf := enumOf.types[enumType]
	code := ""
	code += module.unionFnNameSpace(enumOf, unionOf) + "("
	if len(unionOf.types) == 1 {
		unionHas := n.has[0]
		code += me.eval(unionHas).code
	} else {
		for ix := range unionOf.types {
			if ix > 0 {
				code += ", "
			}
			unionHas := n.has[ix]
			code += me.eval(unionHas).code
		}
	}
	code += ")"
	return code
}

func (me *hmfile) allocClass(typed string, n *node) string {
	if n.attribute("no-malloc") {
		return ""
	}
	typed = me.classNameSpace(typed)
	return "malloc(sizeof(" + typed + "))"
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
	atype := typeOfArray(n.typed)
	mtype := me.typeSig(atype)
	return "malloc((" + size.code + ") * sizeof(" + mtype + "))"
}

func (me *hmfile) checkIsClass(typed string) bool {
	_, ok := me.classes[typed]
	return ok
}

func (me *hmfile) checkIsEnum(typed string) bool {
	_, ok := me.enums[typed]
	return ok
}

func (me *hmfile) checkIsUnion(typed string) bool {
	def, ok := me.enums[typed]
	if ok {
		return !def.simple
	}
	return false
}

func (me *cfile) typeSig(typed string) string {
	if checkIsArray(typed) {
		arrayType := typeOfArray(typed)
		return fmtptr(me.typeSig(arrayType))
	}
	module, trueType := me.hmfile.moduleAndName(typed)
	if module.checkIsClass(trueType) {
		return module.classNameSpace(trueType) + " *"
	} else if module.checkIsEnum(trueType) {
		if module.enums[trueType].simple {
			return module.enumNameSpace(trueType)
		}
		return module.unionNameSpace(trueType) + " *"
	} else if typed == "string" {
		return "char *"
	}
	return typed
}

func (me *cfile) noMalloctypeSig(typed string) string {
	if checkIsArray(typed) {
		arraytype := typeOfArray(typed)
		return fmtptr(me.noMalloctypeSig(arraytype))
	}
	module, trueType := me.hmfile.moduleAndName(typed)
	if module.checkIsClass(trueType) {
		return module.classNameSpace(typed)
	} else if module.checkIsEnum(trueType) {
		if module.enums[trueType].simple {
			return module.enumNameSpace(trueType)
		}
		return module.unionNameSpace(trueType)
	} else if typed == "string" {
		return "char *"
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
			codesig := fmtassignspace(me.typeSig(typed))
			module, trueType := me.hmfile.moduleAndName(typed)
			if mutable {
				code = codesig
			} else if checkIsArray(typed) || module.checkIsClass(trueType) || module.checkIsUnion(trueType) {
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
			codesig := fmtassignspace(me.noMalloctypeSig(typed))
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
	rightCode := me.eval(right).code
	code += me.eval(left).code + me.maybeLet(rightCode) + rightCode
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
		code := module.allocClass(className, n)
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "enum" {
		module, enumName := me.hmfile.moduleAndName(n.typed)
		code := me.allocEnum(module, enumName, n)
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
	if op == "concat" {
		paren := n.attribute("parenthesis")
		code := ""
		if paren {
			code += "("
		}
		code += me.eval(n.has[0]).code
		code += " + "
		code += me.eval(n.has[1]).code
		if paren {
			code += ")"
		}
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
	if op == "match" {
		code := ""
		depth := me.depth
		variable := me.getvar(n.value)
		_, isEnum := me.hmfile.enums[variable.typed]
		baseEnum := me.hmfile.enumNameSpace(variable.typed)
		code += "switch (" + n.value + "->type) {\n"
		ix := 0
		size := len(n.has)
		for ix < size {
			caseOf := n.has[ix]
			thenDo := n.has[ix+1]
			thenBlock := me.eval(thenDo).code
			if caseOf.is == "_" {
				code += fmc(depth) + "default:\n"
			} else {
				if isEnum {
					code += fmc(depth) + "case " + me.hmfile.enumTypeName(baseEnum, caseOf.is) + ":\n"
				} else {
					code += fmc(depth) + "case " + caseOf.is + ":\n"
				}
			}
			code += fmc(depth+1) + thenBlock + me.maybeColon(thenBlock) + "\n"
			code += fmc(depth+1) + "break;\n"
			ix += 2
		}
		code += fmc(depth) + "}"
		cn := codeNode(n.is, n.value, n.typed, code)
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
		code := "if (" + me.eval(n.has[0]).code + ") {\n"
		me.depth++
		code += me.eval(n.has[1]).code
		me.depth--
		code += fmc(me.depth) + "}"
		ix := 2
		for ix < hsize && n.has[ix].is == "elif" {
			code += " else if (" + me.eval(n.has[ix].has[0]).code + ") {\n"
			me.depth++
			code += me.eval(n.has[ix].has[1]).code
			me.depth--
			code += fmc(me.depth) + "}"
			ix++
		}
		if ix >= 2 && ix < hsize && n.has[ix].is == "block" {
			code += " else {\n"
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

func (me *cfile) generateUnionFn(en *enum, un *union) {
	name := me.hmfile.unionFnNameSpace(en, un)

	head := ""
	head += fmtassignspace(me.typeSig(en.name)) + name + "("
	if len(un.types) == 1 {
		unionHas := un.types[0]
		head += fmtassignspace(me.typeSig(unionHas)) + un.name
	} else {
		for ix := range un.types {
			if ix > 0 {
				head += ", "
			}
			unionHas := un.types[ix]
			head += fmtassignspace(me.typeSig(unionHas)) + un.name + strconv.Itoa(ix)
		}
	}

	head += ");\n"
	me.headFuncSection += head

	// code
	// HmEnumsUnionMammal *hm_enums_new_mammal_cat(const char *cat)
	// {
	//   HmEnumsUnionMammal *const var = malloc(sizeof(HmEnumsUnionMammal));
	//   var->type = HmEnumsMammalCat;
	//   var->cat = cat;
	//   return var;
	// }
	//

	// baseName := me.hmfile.enumNameSpace(vType)
	// enumType := right.value
	// code += vName + "->type = " + me.hmfile.enumTypeName(baseName, enumType)
	// unionOf := enumOf.types[enumType]
	// if len(unionOf.types) == 1 {
	// 	unionHas := right.has[0]
	// 	code += ";\n" + fmc(me.depth) + vName + "->" + unionOf.name + " = " + me.eval(unionHas).code
	// } else {
	// 	for ix := range unionOf.types {
	// 		unionHas := right.has[ix]
	// 		code += ";\n" + fmc(me.depth) + vName + "->" + unionOf.name + ".var" + strconv.Itoa(ix) + " = " + me.eval(unionHas).code
	// 	}
	// }
}

func (me *cfile) defineEnum(enum *enum) {
	fmt.Println("define enum \"" + enum.name + "\"")
	hmBaseEnumName := me.hmfile.enumNameSpace(enum.name)
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

	if enum.simple || len(enum.generics) > 0 {
		return
	}

	code = ""
	hmBaseUnionName := me.hmfile.unionNameSpace(enum.name)
	me.headTypeDefSection += "typedef struct " + hmBaseUnionName + " " + hmBaseUnionName + ";\n"
	code += "struct " + hmBaseUnionName + " {\n"
	code += fmc(1) + hmBaseEnumName + " type;\n"
	code += fmc(1) + "union {\n"
	for _, enumUnion := range enum.typesOrder {
		me.generateUnionFn(enum, enumUnion)
		if len(enumUnion.types) == 1 {
			typed := enumUnion.types[0]
			code += fmc(2) + fmtassignspace(me.typeSig(typed)) + enumUnion.name + ";\n"
		} else {
			code += fmc(2) + "struct {\n"
			for ix, typed := range enumUnion.types {
				code += fmc(3) + fmtassignspace(me.typeSig(typed)) + "var" + strconv.Itoa(ix) + ";\n"
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
		code += fmc(1) + fmtassignspace(me.typeSig(field.typed)) + field.name + ";\n"
	}
	code += "};\n\n"
	me.headTypesSection += code
}

func (me *cfile) maybeColon(code string) string {
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
		code += me.maybeColon(code) + "\n"
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
		codeblock += me.maybeColon(codeblock) + "\n"
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
		block += me.maybeColon(block) + "\n"
	}
	me.popScope()
	code := ""
	code += fmtassignspace(me.typeSig(fn.typed)) + me.hmfile.funcNameSpace(name) + "("
	for ix, arg := range args {
		if ix > 0 {
			code += ", "
		}
		typed := arg.typed
		module, trueType := me.hmfile.moduleAndName(typed)
		codesig := fmtassignspace(me.typeSig(typed))
		if checkIsArray(typed) || module.checkIsClass(trueType) || module.checkIsUnion(trueType) {
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
