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
	mutable    bool
	heap       bool
	pointer    bool
	class      *class
	enum       *enum
	union      *union
	funcSig    *fnSig
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
	c.mutable = me.mutable
	c.heap = me.heap
	c.pointer = me.pointer
	c.class = me.class
	c.enum = me.enum
	c.union = me.union
	c.funcSig = me.funcSig
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

func (me *datatype) isNumber() bool {
	return me.is == dataTypePrimitive && isNumber(me.canonical)
}

func (me *datatype) isBoolean() bool {
	return me.is == dataTypePrimitive && me.canonical == TokenBoolean
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

func (me *datatype) isPointer() bool {
	return me.pointer
}

func (me *datatype) isPointerInC() bool {
	if me.isPrimitive() {
		return false
	}
	return me.pointer
}

func (me *datatype) isPrimitive() bool {
	return me.is == dataTypePrimitive || me.is == dataTypeString
}

func (me *datatype) isAnyIntegerType() bool {
	return me.is == dataTypePrimitive && isAnyIntegerType(me.canonical)
}

func (me *datatype) isInt() bool {
	return me.is == dataTypePrimitive && me.canonical == TokenInt
}

func (me *datatype) isQuestion() bool {
	return me.is == dataTypeUnknown && me.canonical == "?"
}

func (me *datatype) isVoid() bool {
	return me.is == dataTypeUnknown && me.canonical == "void"
}

func (me *datatype) canCastToNumber() bool {
	if me.is == dataTypePrimitive {
		return canCastToNumber(me.canonical)
	}
	return false
}

func (me *datatype) isClass() (*class, bool) {
	if me.class == nil {
		return nil, false
	}
	return me.class, true
}

func (me *datatype) isEnum() (*enum, *union, bool) {
	if me.enum == nil {
		return nil, nil, false
	}
	return me.enum, me.union, true
}

func (me *datatype) getFunction(name string) (*function, bool) {
	if me.module != nil {
		f, ok := me.module.getFunction(name)
		return f, ok
	}
	f, ok := me.hmlib.functions[name]
	return f, ok
}

func (me *datatype) postfixConst() bool {
	if me.isArrayOrSlice() {
		return true
	}
	if me.isSomeOrNone() {
		return me.member.postfixConst()
	}
	if _, ok := me.isClass(); ok {
		return true
	}
	if _, _, ok := me.isEnum(); ok {
		return true
	}
	return false
}

func (me *datatype) noConst() bool {
	if !me.isPrimitive() {
		if !me.heap || !me.pointer {
			return true
		}
	}
	return false
}

func (me *datatype) setIsOnStack(flag bool) {
	me.heap = !flag
}

func (me *datatype) setIsPointer(flag bool) {
	me.pointer = flag
}

func (me *datatype) setOnStackNotPointer() {
	me.setIsPointer(false)
	me.setIsOnStack(true)
}

func (me *datatype) arraySize() string {
	return me.size
}

func (me *datatype) convertArrayToSlice() {
	me.is = dataTypeSlice
	me.size = ""
}

func (me *datatype) memoryGet() string {
	if me.pointer {
		return "->"
	}
	return "."
}

func (me *datatype) equals(b *datatype) bool {
	switch me.is {
	case dataTypeClass:
		fallthrough
	case dataTypeEnum:
		fallthrough
	case dataTypeUnknown:
		fallthrough
	case dataTypeString:
		fallthrough
	case dataTypePrimitive:
		{
			if b.is == dataTypeMaybe {
				b = b.member
			}
			if me.is != b.is {
				return false
			}
			if me.canonical != b.canonical {
				return false
			}
			if me.generics != nil || b.generics != nil {
				if me.generics == nil || b.generics == nil {
					return false
				}
				if len(me.generics) != len(b.generics) {
					return false
				}
				for i, ga := range me.generics {
					gb := b.generics[i]
					if ga.notEquals(gb) {
						return false
					}
				}
			}
		}
	case dataTypeNone:
		{
			return b.is == dataTypeNone || b.is == dataTypeMaybe
		}
	case dataTypeMaybe:
		{
			if b.is == dataTypeNone {
				return true
			}
			if me.member.notEquals(b) {
				return false
			}
		}
	case dataTypeSlice:
		{
			if b.is == dataTypeMaybe {
				b = b.member
			}
			if b.is != dataTypeSlice {
				return false
			}
			if me.member.notEquals(b.member) {
				return false
			}
		}
	case dataTypeArray:
		{
			if b.is == dataTypeMaybe {
				b = b.member
			}
			if b.is != dataTypeArray {
				return false
			}
			if me.size != b.size {
				return false
			}
			if me.member.notEquals(b.member) {
				return false
			}
		}
	case dataTypeFunction:
		{
			if b.is == dataTypeMaybe {
				b = b.member
			}
			if b.is != dataTypeFunction {
				return false
			}
			if len(me.parameters) != len(b.parameters) {
				return false
			}
			if me.returns.notEquals(b.returns) {
				return false
			}
			for i, pa := range me.parameters {
				pb := b.parameters[i]
				if pa.notEquals(pb) {
					return false
				}
			}
		}
	default:
		panic("switch statement is missing data type \"" + me.nameIs() + "\"")
	}
	return true
}

func (me *datatype) notEquals(b *datatype) bool {
	return !me.equals(b)
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
			if c, ok := getCName(me.canonical); ok {
				return c
			}
			return me.canonical
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
		{
			return me.class.cname
		}
	default:
		panic("missing data type")
	}
}

func (me *datatype) print() string {
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
			return "maybe<" + me.member.print() + ">"
		}
	case dataTypeNone:
		{
			if me.member != nil {
				return "none<" + me.member.print() + ">"
			}
			return "none"
		}
	case dataTypeArray:
		{
			return "[" + me.size + "]" + me.member.print()
		}
	case dataTypeSlice:
		{
			return "[]" + me.member.print()
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
			f += ") " + me.member.print()
			return f
		}
	case dataTypeClass:
		{
			f := me.class.baseClass().name
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
	case dataTypeEnum:
		{
			f := me.enum.baseEnum().name
			if me.union != nil {
				f += "." + me.union.name
			}
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

func newdataclass(module *hmfile, class *class, generics []*datatype) *datatype {
	d := newdatatype(dataTypeClass)
	d.module = module
	d.class = class
	d.generics = generics
	return d
}

func newdataenum(module *hmfile, enum *enum, union *union, generics []*datatype) *datatype {
	d := newdatatype(dataTypeEnum)
	d.module = module
	d.enum = enum
	d.union = union
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

func newdatafunction(funcSig *fnSig, parameters []*datatype, returns *datatype) *datatype {
	d := newdatatype(dataTypeFunction)
	d.funcSig = funcSig
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

	if strings.HasPrefix(typed, "maybe<") {
		return newdatamaybe(getdatatype(me, typed[6:len(typed)-1]))

	} else if typed == "none" {
		return newdatanone()

	} else if strings.HasPrefix(typed, "none<") {
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

	if checkIsFunction(typed) {
		parameters, returns := functionSigType(typed)
		list := make([]*datatype, len(parameters))
		funcSig := fnSigInit(me)
		for i, p := range parameters {
			list[i] = getdatatype(me, p)
			funcSig.args = append(funcSig.args, fnArgInit(typeToVarData(me, p).asVariable()))
		}
		funcSig.returns = typeToVarData(me, returns)
		return newdatafunction(funcSig, list, getdatatype(me, returns))
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

	var base string
	var glist []*datatype
	g := strings.Index(typed, "<")
	if g != -1 {
		graw := getdatatypegenerics(typed)
		base = typed[0:g]
		glist = make([]*datatype, len(graw))
		for i, r := range graw {
			glist[i] = getdatatype(me, r)
		}
	}

	d = strings.Index(typed, ".")
	if d != -1 {
		base = typed[0:d]
		if en, ok := module.enums[base]; ok {
			un := en.types[typed[d+1:]]
			return newdataenum(module, en, un, glist)
		}
		return newdataunknown(module, typed, glist)
	}

	if cl, ok := module.classes[typed]; ok {
		return newdataclass(module, cl, glist)
	} else if en, ok := module.enums[typed]; ok {
		return newdataenum(module, en, nil, glist)
	} else if base != typed {
		if cl, ok := module.classes[base]; ok {
			return newdataclass(module, cl, glist)
		}
	}

	return newdataunknown(module, typed, glist)
}

func getdatatypegenerics(typed string) []string {
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

func (me *datatype) typeSigOf(name string, mutable bool) string {
	code := ""
	if me.is == dataTypeFunction {
		code += fmtassignspace(me.returns.typeSig())
		code += "(*"
		if !mutable {
			code += "const "
		}
		code += name
		code += ")("
		for ix, arg := range me.parameters {
			if ix > 0 {
				code += ", "
			}
			code += arg.typeSig()
		}
		code += ")"
	} else {
		sig := fmtassignspace(me.typeSig())
		if mutable || me.noConst() {
			code += sig
		} else if me.postfixConst() {
			code += sig + "const "
		} else {
			code += "const " + sig
		}
		code += name
	}
	return code
}

func (me *datatype) typeSig() string {
	switch me.is {
	case dataTypeClass:
		{
			out := me.class.cname
			if me.heap && me.pointer {
				out += " *"
			}
			return out
		}
	case dataTypeEnum:
		{
			return me.enum.typeSig()
		}
	case dataTypeNone:
		fallthrough
	case dataTypeMaybe:
		{
			return me.member.typeSig()
		}
	case dataTypeSlice:
		fallthrough
	case dataTypeArray:
		{
			return fmtptr(me.member.typeSig())
		}
	case dataTypeUnknown:
		fallthrough
	case dataTypeString:
		fallthrough
	case dataTypePrimitive:
		{
			if c, ok := getCName(me.canonical); ok {
				return c
			}
			return me.canonical
		}
	default:
		panic("switch statement is missing data type \"" + me.nameIs() + "\"")
	}
}

func (me *datatype) noMallocTypeSig() string {
	switch me.is {
	case dataTypeClass:
		{
			return me.class.cname
		}
	case dataTypeEnum:
		{
			return me.enum.noMallocTypeSig()
		}
	case dataTypeNone:
		fallthrough
	case dataTypeMaybe:
		{
			return me.member.noMallocTypeSig()
		}
	case dataTypeSlice:
		fallthrough
	case dataTypeArray:
		{
			return fmtptr(me.member.noMallocTypeSig())
		}
	case dataTypeUnknown:
		fallthrough
	case dataTypeString:
		fallthrough
	case dataTypePrimitive:
		{
			if c, ok := getCName(me.canonical); ok {
				return c
			}
			return me.canonical
		}
	default:
		panic("switch statement is missing data type \"" + me.nameIs() + "\"")
	}
}
