package main

import (
	"fmt"
	"strings"
)

const (
	dataTypePrimitive = 0
	dataTypeMaybe     = 1
	dataTypeArray     = 2
	dataTypeFunction  = 3
	dataTypeClass     = 4
	dataTypeEnum      = 5
)

type datatype struct {
	is         int
	canonical  string
	size       string
	member     *datatype
	parameters []*datatype
	returns    *datatype
	generics   []*datatype
}

func (me *datatype) standard() string {
	return me.string(false)
}

func (me *datatype) print() string {
	return me.string(true)
}

func (me *datatype) string(expand bool) string {
	switch me.is {
	case dataTypePrimitive:
		{
			return me.canonical
		}
	case dataTypeMaybe:
		{
			if expand {
				return "maybe<" + me.member.print() + ">"
			}
			return me.member.print()
		}
	case dataTypeArray:
		{
			return "[" + me.size + "]" + me.member.print()
		}
	case dataTypeFunction:
		{
			f := me.canonical + "("
			for i, p := range me.parameters {
				if i > 0 {
					f += ","
				}
				f += p.print()
			}
			f += ") " + me.returns.print()
			return f
		}
	case dataTypeClass:
		fallthrough
	case dataTypeEnum:
		{
			f := me.canonical
			if len(me.generics) > 0 {
				f += "<"
				for i, g := range me.generics {
					if i > 0 {
						f += ","
					}
					f += g.print()
				}
				f += ">"
			}
			return f
		}
	default:
		panic("missing data type")
	}
}

func (me *hmfile) getdatatype(typed string) *datatype {

	if checkIsPrimitive(typed) {
		return &datatype{is: dataTypePrimitive, canonical: typed}
	}

	if strings.HasPrefix(typed, "maybe") {
		return &datatype{is: dataTypeMaybe, member: me.getdatatype(typed[6 : len(typed)-1])}

	} else if strings.HasPrefix(typed, "none") {
		return &datatype{is: dataTypeMaybe, member: me.getdatatype(typed[5 : len(typed)-1])}
	}

	if checkIsArray(typed) || checkIsSlice(typed) {
		size, member := typeOfArrayOrSlice(typed)
		return &datatype{is: dataTypeArray, size: size, member: me.getdatatype(member)}
	}

	if checkIsFunction(typed) {
		parameters, returns := functionSigType(typed)
		list := make([]*datatype, len(parameters))
		for i, p := range parameters {
			list[i] = me.getdatatype(p)
		}
		return &datatype{is: dataTypeFunction, parameters: list, returns: me.getdatatype(returns)}
	}

	dot := strings.Split(typed, ".")
	if len(dot) != 1 {
		if _, ok := me.imports[dot[0]]; ok {
			fmt.Println("FIRST DOT IS FROM AN IMPORT")
			return &datatype{}
		}

		fmt.Println("FIRST DOT IS FOR AN ENUM")
		return &datatype{}
	}

	g := strings.Index(typed, "<")
	if g != -1 {
		return &datatype{}
	}

	return &datatype{is: dataTypeClass, canonical: typed}
}
