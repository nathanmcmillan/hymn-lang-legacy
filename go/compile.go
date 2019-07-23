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
	if op == "+=" || op == "-=" || op == "*=" || op == "/=" {
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
	if op == "+" || op == "-" || op == "*" || op == "/" || op == "&" || op == "|" || op == "^" || op == "<<" || op == ">>" {
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
	if op == "variable" {
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

func (me *cfile) allocClass(n *node) *cnode {
	if n.attribute("no-malloc") {
		return codeNode(n.is, n.value, n.typed, "")
	}

	data := me.hmfile.typeToVarData(n.typed)
	typed := data.module.classNameSpace(data.typed)

	// HmConstructorAttributeAttributeVec *const z = calloc(1, sizeof(HmConstructorAttributeAttributeVec));
	// z->on = true;
	// z->has = calloc(1, sizeof(HmConstructorAttributeVec));
	// z->has->on = true;
	// z->has->has = calloc(1, sizeof(HmConstructorVec));
	// z->has->has->x = 2 * 5;
	// z->has->has->y = 2 * 6;
	// z->has->has->z = 2 * 7;

	// HmConstructorVec *temp_0 = calloc(1, sizeof(HmConstructorVec));
	// temp_0->x = 2 * 5;
	// temp_0->y = 2 * 6;
	// temp_0->z = 2 * 7;
	// HmConstructorAttributeVec *temp_1 = calloc(1, sizeof(HmConstructorAttributeVec));
	// temp_1->on = true;
	// temp_1->has = temp_0;
	// HmConstructorAttributeAttributeVec *const z = calloc(1, sizeof(HmConstructorAttributeAttributeVec));
	// z->on = true;
	// z->has = temp_1;

	// {is:=, typed:void, has[
	//   {is:variable, value:temp_0, typed:vec}
	//   {is:new, typed:vec, has[...]}
	// ]}

	// {is:=, typed:void, has[
	//   {is:variable, value:temp_1, typed:vec}
	//   {is:new, typed:attribute<attribute<vec>>, has[...]}
	// ]}

	// {is:=, typed:void, has[
	//   {is:variable, value:z, typed:attribute<attribute<vec>>}
	//   {is:new, typed:attribute<attribute<vec>>, has[..]}
	// ]}

	code := ""
	code += "malloc(sizeof(" + typed + "))"
	// code += "calloc(1, sizeof(" + typed + "))"

	base := me.hmfile.classes[n.typed]
	ctor := n.has
	for ix, ini := range ctor {
		v := base.variables[base.variableOrder[ix]]
		code += ";\n" + fmc(me.depth) + n.value + "->" + v.name + " = " + me.eval(ini).code
	}

	return codeNode(n.is, n.value, n.typed, code)
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
	data := me.hmfile.typeToVarData(typed)
	if data.module.checkIsClass(data.typed) {
		return data.module.classNameSpace(data.typed) + " *"
	} else if data.module.checkIsEnum(data.typed) {
		if data.module.enums[data.typed].simple {
			return data.module.enumNameSpace(data.typed)
		}
		return data.module.unionNameSpace(data.typed) + " *"
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
	data := me.hmfile.typeToVarData(typed)
	if data.module.checkIsClass(data.typed) {
		return data.module.classNameSpace(typed)
	} else if data.module.checkIsEnum(data.typed) {
		if data.module.enums[data.typed].simple {
			return data.module.enumNameSpace(data.typed)
		}
		return data.module.unionNameSpace(data.typed)
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
			data := me.hmfile.typeToVarData(typed)
			if mutable {
				code = codesig
			} else if checkIsArray(typed) || data.module.checkIsClass(data.typed) || data.module.checkIsUnion(data.typed) {
				code += codesig + "const "
			} else {
				code += "const " + codesig
			}
		} else {
			typed := n.typed
			data := me.hmfile.typeToVarData(typed)
			newVar := me.hmfile.varInit(typed, name, mutable, malloc)
			newVar.cName = data.module.varNameSpace(name)
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

func (me *cfile) free(name string) string {
	return "free(" + name + ");"
}

func (me *cfile) generateUnionFn(en *enum, un *union) {
	_, enumName := me.hmfile.enumMaybeImplNameSpace(en.name)
	unionName := me.hmfile.unionNameSpace(en.name)
	fnName := me.hmfile.unionFnNameSpace(en, un)
	typeOf := fmtassignspace(me.typeSig(en.name))
	head := ""
	head += typeOf + fnName + "("
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
			code += fmc(2) + fmtassignspace(me.typeSig(typed)) + enumUnion.name + ";\n"
		} else if num != 0 {
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
	size := len(code)
	if size == 0 {
		return ""
	}
	last := code[size-1]
	if last == '}' || last == ':' {
		return ""
	}
	return ";"
}

func (me *cfile) maybeFmc(code string, depth int) string {
	if code == "" {
		return ""
	}
	return fmc(depth)
}

func (me *cfile) block(n *node) *cnode {
	expressions := n.has
	code := ""
	for _, expr := range expressions {
		c := me.eval(expr)
		if c.code != "" {
			code += me.maybeFmc(c.code, me.depth) + c.code + me.maybeColon(c.code) + "\n"
		}
	}
	cn := codeNode(n.is, n.value, n.typed, code)
	fmt.Println(cn.string(0))
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
	code += fmtassignspace(me.typeSig(fn.typed)) + me.hmfile.funcNameSpace(name) + "("
	for ix, arg := range args {
		if ix > 0 {
			code += ", "
		}
		typed := arg.typed
		data := me.hmfile.typeToVarData(typed)
		codesig := fmtassignspace(me.typeSig(typed))
		if checkIsArray(typed) || data.module.checkIsClass(data.typed) || data.module.checkIsUnion(data.typed) {
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
