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
		code += fmc(1) + field.data().typeSigOf(field.name, true) + ";\n"
	}
	code += "};\n\n"
	me.headTypesSection += code
}
