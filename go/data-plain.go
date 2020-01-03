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
	dataTypeString    = 9
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
	c.hmlib = me.hmlib
	c.original = me.original
	c.mutable = me.mutable
	c.heap = me.heap
	c.pointer = me.pointer
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
	return me.is == dataTypeString
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

func (me *datatype) isIndexable() bool {
	return me.is == dataTypeString || me.isArrayOrSlice()
}

func (me *datatype) isPointerInC() bool {
	if me.isPrimitive() {
		return false
	}
	return me.pointer
}

func (me *datatype) isPrimitive() bool {
	return me.is == dataTypePrimitive
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
	case dataTypeString:
		return "string"
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
	case dataTypeString:
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
	case dataTypeString:
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

func newdatatype(is int) *datatype {
	d := &datatype{}
	d.is = is
	d.mutable = true
	d.pointer = true
	d.heap = true
	return d
}

func newdatamaybe(member *datatype) *datatype {
	d := newdatatype(dataTypeMaybe)
	d.member = member
	return d
}

func newdatanone() *datatype {
	return newdatatype(dataTypeNone)
}

func newdatastring() *datatype {
	d := newdatatype(dataTypeString)
	d.canonical = TokenString
	return d
}

func newdataprimitive(canonical string) *datatype {
	d := newdatatype(dataTypePrimitive)
	d.canonical = canonical
	d.pointer = false
	d.heap = false
	return d
}

func newdataclass(module *hmfile, canonical string, generics []*datatype) *datatype {
	d := newdatatype(dataTypeClass)
	d.module = module
	d.canonical = canonical
	d.generics = generics
	return d
}

func newdataenum(module *hmfile, canonical string, generics []*datatype) *datatype {
	d := newdatatype(dataTypeEnum)
	d.module = module
	d.canonical = canonical
	d.generics = generics
	return d
}

func newdataunknown(module *hmfile, canonical string, generics []*datatype) *datatype {
	d := newdatatype(dataTypeUnknown)
	d.module = module
	d.canonical = canonical
	d.generics = generics
	return d
}

func newdataarray(size string, member *datatype) *datatype {
	d := newdatatype(dataTypeArray)
	d.size = size
	d.member = member
	return d
}

func newdataslice(member *datatype) *datatype {
	d := newdatatype(dataTypeSlice)
	d.member = member
	return d
}

func newdatafunction(module *hmfile, parameters []*datatype, returns *datatype) *datatype {
	d := newdatatype(dataTypeFunction)
	d.module = module
	d.parameters = parameters
	d.returns = returns
	return d
}

func getdatatype(me *hmfile, typed string) *datatype {

	if typed == TokenString {
		return newdatastring()
	}

	if checkIsPrimitive(typed) {
		return newdataprimitive(typed)
	}

	if strings.HasPrefix(typed, "maybe") {
		return newdatamaybe(getdatatype(me, typed[6:len(typed)-1]))

	} else if typed == "none" {
		return newdatanone()

	} else if strings.HasPrefix(typed, "none") {
		return newdatamaybe(getdatatype(me, typed[5:len(typed)-1]))
	}

	if checkIsArray(typed) {
		size, member := typeOfArrayOrSlice(typed)
		return newdataarray(size, getdatatype(me, member))
	}

	if checkIsSlice(typed) {
		_, member := typeOfArrayOrSlice(typed)
		return newdataslice(getdatatype(me, member))
	}

	if me == nil {
		return newdataunknown(nil, typed, nil)
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
		return newdatafunction(module, list, getdatatype(me, returns))
	}

	g := strings.Index(typed, "<")
	if g != -1 {
		base := typed[0:g]
		graw := me.getdatatypegenerics(typed)
		glist := make([]*datatype, len(graw))
		for i, r := range graw {
			glist[i] = getdatatype(me, r)
		}
		if _, ok := module.classes[base]; ok {
			return newdataclass(module, base, glist)
		} else if _, ok := module.enums[base]; ok {
			return newdataenum(module, base, glist)
		} else {
			return newdataunknown(module, base, glist)
		}
	}

	d = strings.Index(typed, ".")
	if d != -1 {
		base := typed[0:d]
		if _, ok := module.enums[base]; ok {
			return newdataenum(module, base, nil)
		}
		return newdataunknown(module, base, nil)
	}

	if _, ok := module.classes[typed]; ok {
		return newdataclass(module, typed, nil)
	} else if _, ok := module.enums[typed]; ok {
		return newdataenum(module, typed, nil)
	}

	return newdataunknown(module, typed, nil)
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
