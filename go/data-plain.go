package main

import (
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
	dataTypeNone      = 7
)

type datatype struct {
	module     *hmfile
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

func (me *datatype) cname() string {
	switch me.is {
	case dataTypeUnknown:
		fallthrough
	case dataTypePrimitive:
		{
			return simpleCapitalize(me.canonical)
		}
	case dataTypeArray:
		{
			return "Array" + me.size + me.member.cname()
		}
	case dataTypeFunction:
		{
			f := simpleCapitalize(me.canonical) + "("
			for i, p := range me.parameters {
				if i > 0 {
					f += ","
				}
				f += p.cname()
			}
			f += ") " + me.returns.cname()
			return f
		}
	case dataTypeClass:
		fallthrough
	case dataTypeEnum:
		{
			f := simpleCapitalize(me.canonical)
			if len(me.generics) > 0 {
				for i, g := range me.generics {
					if i > 0 {
						f += ","
					}
					f += g.cname()
				}
			}
			return f
		}
	default:
		panic("missing data type")
	}
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
				return "maybe<" + me.member.string(expand) + ">"
			}
			return me.member.string(expand)
		}
	case dataTypeNone:
		{
			if me.member != nil {
				if expand {
					return "none<" + me.member.string(expand) + ">"
				}
				return me.member.string(expand)
			}
			return "none"
		}
	case dataTypeArray:
		{
			return "[" + me.size + "]" + me.member.string(expand)
		}
	case dataTypeFunction:
		{
			f := me.canonical + "("
			for i, p := range me.parameters {
				if i > 0 {
					f += ","
				}
				f += p.string(expand)
			}
			f += ") " + me.member.string(expand)
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
					f += g.string(expand)
				}
				f += ">"
			}
			return f
		}
	default:
		panic("missing data type")
	}
}

func getdatatype(me *hmfile, typed string) *datatype {

	if me == nil {
		return &datatype{module: me, is: dataTypeUnknown, canonical: typed}
	}

	module := me
	d := strings.Index(typed, ".")
	if d != -1 {
		base := typed[0:d]
		if imp, ok := me.imports[base]; ok {
			module = imp
			typed = typed[d+1:]
		}
	}

	if checkIsPrimitive(typed) {
		return &datatype{module: module, is: dataTypePrimitive, canonical: typed}
	}

	if strings.HasPrefix(typed, "maybe") {
		return &datatype{module: module, is: dataTypeMaybe, member: getdatatype(me, typed[6:len(typed)-1])}

	} else if typed == "none" {
		return &datatype{module: module, is: dataTypeNone}

	} else if strings.HasPrefix(typed, "none") {
		return &datatype{module: module, is: dataTypeMaybe, member: getdatatype(me, typed[5:len(typed)-1])}
	}

	if checkIsArray(typed) || checkIsSlice(typed) {
		size, member := typeOfArrayOrSlice(typed)
		return &datatype{module: module, is: dataTypeArray, size: size, member: getdatatype(me, member)}
	}

	if checkIsFunction(typed) {
		parameters, returns := functionSigType(typed)
		list := make([]*datatype, len(parameters))
		for i, p := range parameters {
			list[i] = getdatatype(me, p)
		}
		return &datatype{module: module, is: dataTypeFunction, parameters: list, returns: getdatatype(me, returns)}
	}

	g := strings.Index(typed, "<")
	if g != -1 {
		base := typed[0:g]
		var is int
		if _, ok := me.classes[base]; ok {
			is = dataTypeClass
		} else if _, ok := me.enums[base]; ok {
			is = dataTypeEnum
		} else {
			is = dataTypeUnknown
		}
		graw := me.getdatatypegenerics(typed)
		glist := make([]*datatype, len(graw))
		for i, r := range graw {
			glist[i] = getdatatype(me, r)
		}
		return &datatype{module: module, is: is, canonical: base, generics: glist}
	}

	d = strings.Index(typed, ".")
	if d != -1 {
		base := typed[0:d]
		var is int
		if _, ok := me.enums[base]; ok {
			is = dataTypeEnum
		} else {
			is = dataTypeUnknown
		}
		return &datatype{module: module, is: is, canonical: base}
	}

	var is int
	if _, ok := me.classes[typed]; ok {
		is = dataTypeClass
	} else if _, ok := me.enums[typed]; ok {
		is = dataTypeEnum
	} else {
		is = dataTypeUnknown
	}

	return &datatype{module: module, is: is, canonical: typed}
}

func (me *hmfile) getdatatypegenerics(typed string) []string {
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
