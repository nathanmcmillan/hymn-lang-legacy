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
	dataTypeUnknown   = 6
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
	case dataTypeUnknown:
		fallthrough
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

	g := strings.Index(typed, "<")
	if g != -1 {
		base := typed[0:g]
		var is int
		if _, ok := me.classes[typed]; ok {
			is = dataTypeClass
		} else if _, ok := me.enums[typed]; ok {
			is = dataTypeEnum
		} else {
			is = dataTypeUnknown
		}
		generics := typed[g+1:]
		fmt.Println("GET DATA TYPE GENERIC ::", base, "|", generics, "| is:", is)
		graw := me.getdatatypegenerics(typed)
		glist := make([]*datatype, len(graw))
		for i, r := range graw {
			glist[i] = me.getdatatype(r)
			fmt.Println("GET DATA TYPE GENERIC ITEM ::", r, "->", glist[i].print())
		}
		return &datatype{is: is, canonical: base, generics: glist}
	}

	dot := strings.Split(typed, ".")
	if len(dot) != 1 {
		if _, ok := me.imports[dot[0]]; ok {
			fmt.Println(":: FIRST DOT IS FROM AN IMPORT")
			return &datatype{}
		}
		fmt.Println(":: FIRST DOT IS FOR AN ENUM")
		return &datatype{}
	}

	var is int
	if _, ok := me.classes[typed]; ok {
		is = dataTypeClass
	} else if _, ok := me.enums[typed]; ok {
		is = dataTypeEnum
	} else {
		is = dataTypeUnknown
	}

	return &datatype{is: is, canonical: typed}
}

func (me *hmfile) getdatatypegenerics(typed string) []string {
	fmt.Println("getdatatypegenerics in ::", typed)
	var order []string
	stack := make([]*gstack, 0)
	rest := typed
	for {
		begin := strings.Index(rest, "<")
		end := strings.Index(rest, ">")
		comma := strings.Index(rest, ",")
		if begin != -1 && (begin < end || end == -1) && (begin < comma || comma == -1) {
			name := rest[:begin]
			current := &gstack{}
			current.name = name
			stack = append(stack, current)
			rest = rest[begin+1:]
		} else if end != -1 && (end < begin || begin == -1) && (end < comma || comma == -1) {
			size := len(stack) - 1
			current := stack[size]
			if end == 0 {
			} else {
				sub := rest[:end]
				current.order = append(current.order, sub)
			}
			stack = stack[:size]
			if size == 0 {
				order = current.order
				break
			} else {
				pop := current.name + "<" + strings.Join(current.order, ",") + ">"
				next := stack[len(stack)-1]
				next.order = append(next.order, pop)
			}
			if end == 0 {
				rest = rest[1:]
			} else {
				rest = rest[end+1:]
			}
		} else if comma != -1 && (comma < begin || begin == -1) && (comma < end || end == -1) {
			current := stack[len(stack)-1]
			if comma == 0 {
				rest = rest[1:]
				continue
			}
			sub := rest[:comma]
			current.order = append(current.order, sub)
			rest = rest[comma+1:]
		}
	}
	return order
}
