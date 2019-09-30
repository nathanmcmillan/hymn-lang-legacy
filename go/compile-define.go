package main

import (
	"strconv"
)

func (me *cfile) defineEnum(enum *enum) {

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
		code += fmc(1) + field.vdat.typeSigOf(field.name, true) + ";\n"
	}
	code += "};\n\n"
	me.headTypesSection += code
}

func (me *cfile) defineFunction(name string, fn *function) {
	args := fn.args
	expressions := fn.expressions
	block := ""
	me.pushScope()
	me.depth = 1
	for _, arg := range args {
		me.scope.variables[arg.name] = arg.variable
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
		code += arg.vdat.typeSigOf(arg.name, false)
	}
	head := code + ");\n"
	code += ") {\n"
	code += block
	code += "}\n\n"

	me.headFuncSection += head
	me.codeFn = append(me.codeFn, code)
}

func (me *cfile) defineMain(fn *function) {
	if len(fn.args) > 0 {
		panic("main can not have arguments")
	}
	expressions := fn.expressions
	codeblock := ""
	returns := false
	me.pushScope()
	me.depth = 1
	list := me.hmfile.program.hmorder
	for x := len(list) - 1; x >= 0; x-- {
		file := list[x]
		if file.needInit {
			codeblock += fmc(me.depth) + file.funcNameSpace("init") + "();\n"
		}
	}
	for _, expr := range expressions {
		c := me.eval(expr)
		if c.is == "return" {
			if c.getType() != TokenInt {
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
