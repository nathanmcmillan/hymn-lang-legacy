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
	dataTypeSlice     = 8
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
	hmlib      *hmlib
	original   string
	mutable    bool
	heap       bool
	pointer    bool
	en         *enum
	un         *union
	cl         *class
	fn         *fnSig
}

func (me *datatype) copyTo(c *datatype) {
	c.module = me.module
	c.is = me.is
	c.canonical = me.canonical
	c.size = me.size
	if me.member != nil {
		c.member = me.member.copy()
	}
	if me.parameters != nil {
		c.parameters = make([]*datatype, len(me.parameters))
		for i, p := range me.parameters {
			c.parameters[i] = p.copy()
		}
	}
	if me.returns != nil {
		c.returns = me.returns.copy()
	}
	if me.generics != nil {
		c.generics = make([]*datatype, len(me.generics))
		for i, g := range me.generics {
			c.generics[i] = g.copy()
		}
	}
}

func (me *datatype) copy() *datatype {
	c := &datatype{}
	me.copyTo(c)
	return c
}

func (me *datatype) isSomeOrNone() bool {
	return me.is == dataTypeMaybe || me.is == dataTypeNone
}

func (me *datatype) isString() bool {
	return me.is == dataTypePrimitive && me.canonical == TokenString
}

func (me *datatype) isChar() bool {
	return me.is == dataTypePrimitive && me.canonical == TokenChar
}

func (me *datatype) isArray() bool {
	return me.is == dataTypeArray || (me.is == dataTypePrimitive && me.canonical == TokenString)
}

func (me *datatype) isSlice() bool {
	return me.is == dataTypeSlice
}

func (me *datatype) isArrayOrSlice() bool {
	return me.isArray() || me.isSlice()
}

func (me *datatype) isPointerInC() bool {
	return me.pointer
}

func (me *datatype) standard() string {
	return me.output(false)
}

func (me *datatype) print() string {
	return me.output(true)
}

func (me *datatype) nameIs() string {
	switch me.is {
	case dataTypePrimitive:
		return "primitive"
	case dataTypeMaybe:
		return "maybe"
	case dataTypeArray:
		return "array"
	case dataTypeSlice:
		return "slice"
	case dataTypeFunction:
		return "function"
	case dataTypeClass:
		return "class"
	case dataTypeEnum:
		return "enum"
	case dataTypeUnknown:
		return "unknown"
	case dataTypeNone:
		return "none"
	}
	panic("missing data type")
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
	case dataTypeSlice:
		{
			return "Slice" + me.member.cname()
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

func (me *datatype) output(expand bool) string {
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
				return "maybe<" + me.member.output(expand) + ">"
			}
			return me.member.output(expand)
		}
	case dataTypeNone:
		{
			if me.member != nil {
				if expand {
					return "none<" + me.member.output(expand) + ">"
				}
				return me.member.output(expand)
			}
			return "none"
		}
	case dataTypeArray:
		{
			return "[" + me.size + "]" + me.member.output(expand)
		}
	case dataTypeSlice:
		{
			return "[]" + me.member.output(expand)
		}
	case dataTypeFunction:
		{
			f := me.canonical + "("
			for i, p := range me.parameters {
				if i > 0 {
					f += ","
				}
				f += p.output(expand)
			}
			f += ") " + me.member.output(expand)
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
					f += g.output(expand)
				}
				f += ">"
			}
			return f
		}
	default:
		panic("missing data type")
	}
}

func newdatatypearray(module *hmfile, size string, member *datatype) *datatype {
	return &datatype{module: module, is: dataTypeArray, size: size, member: member}
}

func newdatatypeslice(module *hmfile, member *datatype) *datatype {
	return &datatype{module: module, is: dataTypeSlice, member: member}
}

func getdatatype(me *hmfile, typed string) *datatype {

	if checkIsPrimitive(typed) {
		return &datatype{is: dataTypePrimitive, canonical: typed}
	}

	if strings.HasPrefix(typed, "maybe") {
		return &datatype{is: dataTypeMaybe, member: getdatatype(me, typed[6:len(typed)-1])}

	} else if typed == "none" {
		return &datatype{is: dataTypeNone}

	} else if strings.HasPrefix(typed, "none") {
		return &datatype{is: dataTypeMaybe, member: getdatatype(me, typed[5:len(typed)-1])}
	}

	if checkIsArray(typed) {
		size, member := typeOfArrayOrSlice(typed)
		return &datatype{is: dataTypeArray, size: size, member: getdatatype(me, member)}
	}

	if checkIsSlice(typed) {
		size, member := typeOfArrayOrSlice(typed)
		return &datatype{is: dataTypeSlice, size: size, member: getdatatype(me, member)}
	}

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
		if _, ok := module.classes[base]; ok {
			is = dataTypeClass
		} else if _, ok := module.enums[base]; ok {
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
		if _, ok := module.enums[base]; ok {
			is = dataTypeEnum
		} else {
			is = dataTypeUnknown
		}
		return &datatype{module: module, is: is, canonical: base}
	}

	var is int
	if _, ok := module.classes[typed]; ok {
		is = dataTypeClass
	} else if _, ok := module.enums[typed]; ok {
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
