package main

import (
	"strconv"
	"strings"
)

func (me *cfile) defineEnum(enum *enum) {

	base := enum.baseEnum()
	impl := base != enum
	hmBaseEnumName := enum.baseEnum().cname

	if !impl {
		code := "\nenum " + hmBaseEnumName + " {\n"
		for ix, enumUnion := range enum.typesOrder {
			if ix == 0 {
				code += fmc(1) + enumTypeName(hmBaseEnumName, enumUnion.name)
			} else {
				code += ",\n" + fmc(1) + enumTypeName(hmBaseEnumName, enumUnion.name)
			}
		}
		code += "\n};\n"
		me.headEnumSection.WriteString(code)
		me.headEnumTypeDefSection.WriteString("\ntypedef enum " + hmBaseEnumName + " " + hmBaseEnumName + ";")
	}

	if enum.simple || len(enum.generics) > 0 {
		return
	}

	me.dependencyReq.add(base.pathLocal)

	code := ""
	hmBaseUnionName := enum.ucname
	me.headStructTypeDefSection.WriteString("\ntypedef struct " + hmBaseUnionName + " " + hmBaseUnionName + ";")
	code += "\nstruct " + hmBaseUnionName + " {\n"
	code += fmc(1) + hmBaseEnumName + " type;\n"
	code += fmc(1) + "union {\n"
	for _, enumUnion := range enum.typesOrder {
		num := len(enumUnion.types)
		if num == 1 {
			typed := enumUnion.types[0]
			me.dependencyGraph(typed)
			code += fmc(2) + fmtassignspace(typed.typeSig()) + enumUnion.name + ";\n"
		} else if num != 0 {
			code += fmc(2) + "struct {\n"
			for ix, typed := range enumUnion.types {
				me.dependencyGraph(typed)
				code += fmc(3) + fmtassignspace(typed.typeSig()) + "var" + strconv.Itoa(ix) + ";\n"
			}
			code += fmc(2) + "} " + enumUnion.name + ";\n"
		}
	}
	code += fmc(1) + "};\n"
	code += "};\n"
	me.headStructSection.WriteString(code)
}

func (me *cfile) typedefClass(c *class) string {
	hmName := c.cname
	me.headStructTypeDefSection.WriteString("\ntypedef struct " + hmName + " " + hmName + ";")
	return hmName
}

func (me *cfile) typedefEnum(e *enum) string {
	hmBaseEnumName := e.baseEnum().cname
	me.headEnumTypeDefSection.WriteString("\ntypedef enum " + hmBaseEnumName + " " + hmBaseEnumName + ";")
	return hmBaseEnumName
}

func (me *cfile) defineClass(c *class) {
	if c.doNotDefine {
		return
	}
	hmName := me.typedefClass(c)
	var code strings.Builder
	code.WriteString("\nstruct " + hmName + " {\n")
	for _, name := range c.variableOrder {
		field := c.variables[name]
		me.dependencyGraph(field.data())
		code.WriteString(fmc(1) + field.data().typeSigOf(field.name, true) + ";\n")
	}
	code.WriteString("};\n")
	me.headStructSection.WriteString(code.String())
}

func (me *cfile) dependencyGraph(data *datatype) {
	switch data.is {
	case dataTypeNone:
		{
			if data.member != nil {
				me.dependencyGraph(data.member)
			}
		}
	case dataTypeMaybe:
		fallthrough
	case dataTypeArray:
		fallthrough
	case dataTypeSlice:
		{
			me.dependencyGraph(data.member)
		}
	case dataTypeClass:
		if me.pathGlobal != data.class.pathGlobal {
			me.dependencyReq.add(data.class.pathGlobal)
		}
	case dataTypeEnum:
		if me.pathGlobal != data.enum.pathGlobal {
			me.dependencyReq.add(data.enum.pathGlobal)
		}
	case dataTypeString:
		me.libReq.add(HmLibString)
	case dataTypeUnknown:
		return
	case dataTypePrimitive:
		return
	case dataTypeFunction:
		return
	case dataTypeVoid:
		return
	default:
		data.missingCase()
	}
}
