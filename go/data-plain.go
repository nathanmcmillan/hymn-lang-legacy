package main

import (
	"strconv"
	"strings"
)

const (
	dataTypePrimitive  = 0
	dataTypeMaybe      = 1
	dataTypeArray      = 2
	dataTypeFunction   = 3
	dataTypeClass      = 4
	dataTypeEnum       = 5
	dataTypeUnknown    = 6
	dataTypeNone       = 7
	dataTypeSlice      = 8
	dataTypeString     = 9
	dataTypeVoid       = 10
	dataTypeAny        = 11
	dataTypeAnyPointer = 12
)

type datatype struct {
	origin     *hmfile
	hmlib      *hmlib
	module     *hmfile
	is         int
	canonical  string
	size       string
	member     *datatype
	parameters []*datatype
	variadic   *datatype
	returns    *datatype
	generics   []*datatype
	heap       bool
	pointer    bool
	class      *class
	enum       *enum
	union      *union
	funcSig    *fnSig
}

func (me *datatype) set(in *datatype) {
	me.origin = in.origin
	me.module = in.module
	me.is = in.is
	me.canonical = in.canonical
	me.size = in.size
	if in.member != nil {
		me.member = in.member.copy()
	}
	if in.parameters != nil {
		me.parameters = make([]*datatype, len(in.parameters))
		for i, p := range in.parameters {
			me.parameters[i] = p.copy()
		}
	}
	if in.variadic != nil {
		me.variadic = in.variadic.copy()
	}
	if in.returns != nil {
		me.returns = in.returns.copy()
	}
	if in.generics != nil {
		me.generics = make([]*datatype, len(in.generics))
		for i, g := range in.generics {
			me.generics[i] = g.copy()
		}
	}
	me.hmlib = in.hmlib
	me.heap = in.heap
	me.pointer = in.pointer
	me.class = in.class
	me.enum = in.enum
	me.union = in.union
	me.funcSig = in.funcSig
}

func (me *datatype) copy() *datatype {
	c := &datatype{}
	c.set(me)
	return c
}

func getdatatype(me *hmfile, typed string) (*datatype, *parseError) {

	if me != nil {
		typed = me.alias(typed)
	}

	if typed == "void" {
		return newdatavoid(), nil
	}

	if typed == "?" {
		return newdataany(), nil
	}

	if typed == "*" {
		return newdataanypointer(), nil
	}

	if typed == TokenString {
		return newdatastring(), nil
	}

	if checkIsPrimitive(typed) {
		return newdataprimitive(typed), nil
	}

	if strings.HasPrefix(typed, "maybe<") {
		member, er := getdatatype(me, typed[6:len(typed)-1])
		if er != nil {
			return nil, er
		}
		if !member.isPointer() {
			return nil, err(me.parser, ECodeMaybeTypeRequiresPointer, "Maybe type requires a pointer found: "+member.print())
		}
		return newdatamaybe(member), nil

	} else if typed == "none" {
		return newdatanone(), nil

	} else if strings.HasPrefix(typed, "none<") {
		member, er := getdatatype(me, typed[5:len(typed)-1])
		if er != nil {
			return nil, er
		}
		if !member.isPointer() {
			return nil, err(me.parser, ECodeNoneTypeRequiresPointer, "None type requires a pointer found: "+member.print())
		}
		return newdatamaybe(member), nil
	}

	if checkIsArray(typed) {
		size, member := typeOfArrayOrSlice(typed)
		mem, er := getdatatype(me, member)
		if er != nil {
			return nil, er
		}
		return newdataarray(size, mem), nil
	}

	if checkIsSlice(typed) {
		_, member := typeOfArrayOrSlice(typed)
		mem, er := getdatatype(me, member)
		if er != nil {
			return nil, er
		}
		return newdataslice(mem), nil
	}

	if checkIsFunction(typed) {
		parameters, returns := functionSigType(typed)
		list := make([]*datatype, len(parameters))
		funcSig := fnSigInit(me)
		for i, p := range parameters {
			ls, er := getdatatype(me, p)
			if er != nil {
				return nil, er
			}
			list[i] = ls
			a, er := getdatatype(me, p)
			if er != nil {
				return nil, er
			}
			funcSig.args = append(funcSig.args, a.tofnarg())
		}
		var er *parseError
		funcSig.returns, er = getdatatype(me, returns)
		if er != nil {
			return nil, er
		}
		ret, er := getdatatype(me, returns)
		if er != nil {
			return nil, er
		}
		return newdatafunction(funcSig, list, nil, ret), nil
	}

	if me == nil {
		return newdataunknown(nil, nil, typed, nil), nil
	}

	origin := me
	module := me

	d := strings.Index(typed, ".")
	g := strings.Index(typed, "<")

	// fmt.Println("DEBUG:: ", typed)

	if d != -1 && (g == -1 || d < g) {
		if strings.HasPrefix(typed, "%") {
			uid := typed[1:d]
			if search, ok := me.program.modules[uid]; ok {
				module = search
				typed = typed[d+1:]
				g = strings.Index(typed, "<")
			} else {
				panic("Module UID \"" + uid + "\" not found.")
			}
		} else {
			base := typed[0:d]
			if search, ok := me.imports[base]; ok {
				module = search
				typed = typed[d+1:]
				g = strings.Index(typed, "<")
			}
		}
	}

	base := typed
	var glist []*datatype
	if g != -1 {
		graw := getdatatypegenerics(typed)
		// fmt.Println("GENERICS:: ", graw)
		base = typed[0:g]
		gt := strings.LastIndex(typed, ">") + 1
		remainder := typed[gt:]
		glist = make([]*datatype, len(graw))
		for i, r := range graw {
			var er *parseError
			glist[i], er = getdatatype(me, r)
			// fmt.Println("TYPE:: ", glist[i].error())
			if er != nil {
				return nil, er
			}
		}

		d = strings.Index(remainder, ".")
		if d != -1 {
			base = typed[0:gt]
			// fmt.Println("BASE:: ", base)
			// for k := range module.enums {
			// fmt.Println("ENUMS::", module.name, "::", k)
			// }
			if en, ok := module.enums[base]; ok {
				un := en.getType(remainder[d+1:])
				return newdataenum(origin, en, un, glist), nil
			}
			return newdataunknown(origin, module, typed, glist), nil
		}
	} else {
		d = strings.Index(base, ".")
		if d != -1 {
			base = typed[0:d]
			if en, ok := module.enums[base]; ok {
				un := en.getType(typed[d+1:])
				return newdataenum(origin, en, un, glist), nil
			}
			return newdataunknown(origin, module, typed, glist), nil
		}
	}

	if cl, ok := module.classes[typed]; ok {
		return newdataclass(origin, cl, glist), nil
	} else if en, ok := module.enums[typed]; ok {
		return newdataenum(origin, en, nil, glist), nil
	} else if base != typed {
		if cl, ok := module.classes[base]; ok {
			if cl.module != module {
				if cl, ok := cl.module.classes[typed]; ok {
					return newdataclass(origin, cl, glist), nil
				}
			}
			return newdataclass(origin, cl, glist), nil
		} else if en, ok := module.enums[base]; ok {
			if en.module != module {
				if en, ok := en.module.enums[typed]; ok {
					return newdataenum(origin, en, nil, glist), nil
				}
			}
			return newdataenum(origin, en, nil, glist), nil
		}
	}

	return newdataunknown(origin, module, typed, glist), nil
}

func (me *datatype) missingCase() bool {
	panic("Switch statement is missing data type \"" + me.nameIs() + "\".")
}

func (me *datatype) getmodule() *hmfile {
	return me.module
}

func (me *datatype) getmember() *datatype {
	return me.member
}

func (me *datatype) isOnStack() bool {
	return !me.heap
}

func (me *datatype) isSome() bool {
	return me.is == dataTypeMaybe
}

func (me *datatype) isNone() bool {
	return me.is == dataTypeNone
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
	if me.is == dataTypeString {
		return true
	} else if me.isPrimitive() {
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

func (me *datatype) isUnknown() bool {
	return me.is == dataTypeUnknown
}

func (me *datatype) isRecursiveUnknown() bool {
	if me.isUnknown() {
		return true
	}
	if me.generics != nil {
		for _, g := range me.generics {
			if g.isRecursiveUnknown() {
				return true
			}
		}
	}
	if me.parameters != nil {
		for _, p := range me.parameters {
			if p.isRecursiveUnknown() {
				return true
			}
		}
	}
	if me.variadic != nil && me.variadic.isRecursiveUnknown() {
		return true
	}
	if me.returns != nil && me.returns.isRecursiveUnknown() {
		return true
	}
	if me.member != nil && me.member.isRecursiveUnknown() {
		return true
	}
	return false
}

func (me *datatype) isAnyType() bool {
	return me.is == dataTypeAny
}

func (me *datatype) isAnyPointerType() bool {
	return me.is == dataTypeAnyPointer
}

func (me *datatype) isVoid() bool {
	return me.is == dataTypeVoid
}

func (me *datatype) isFunction() bool {
	return me.is == dataTypeFunction
}

func (me *datatype) functionSignature() *fnSig {
	return me.funcSig
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

func (me *datatype) standardEquals(b *datatype) bool {
	for b.is == dataTypeMaybe {
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
	return true
}

func (me *datatype) equals(b *datatype) bool {
	if me.isAnyType() || b.isAnyType() {
		return true
	} else if b.is == dataTypeAnyPointer {
		if me.is == dataTypeAnyPointer {
			return true
		}
		return me.isPointer()
	}
	switch me.is {
	case dataTypeAnyPointer:
		if b.is == dataTypeAnyPointer {
			return true
		}
		return b.isPointer()
	case dataTypeVoid:
		return b.is == dataTypeVoid
	case dataTypeClass:
		{
			for b.is == dataTypeMaybe {
				b = b.member
			}
			if me.class != b.class {
				return false
			}
		}
	case dataTypeEnum:
		{
			for b.is == dataTypeMaybe {
				b = b.member
			}
			if me.enum != b.enum {
				return false
			}
			if me.union != nil && b.union != nil && me.union != b.union {
				return false
			}
		}
	case dataTypeString:
		fallthrough
	case dataTypeUnknown:
		fallthrough
	case dataTypePrimitive:
		{
			return me.standardEquals(b)
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
			if me.variadic != nil || b.variadic != nil {
				if me.variadic == nil || b.variadic == nil || me.variadic.notEquals(b.variadic) {
					return false
				}
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
		me.missingCase()
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
	case dataTypeVoid:
		return "void"
	case dataTypeAny:
		return "?"
	case dataTypeAnyPointer:
		return "*"
	}
	panic("Missing data type " + strconv.Itoa(me.is))
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
			if me.variadic != nil {
				if len(me.parameters) > 0 {
					f += ","
				}
				f += "..." + me.variadic.cname()
			}
			f += ") " + me.returns.cname()
			return f
		}
	case dataTypeClass:
		{
			return me.class.cname
		}
	default:
		me.missingCase()
	}
	return ""
}

func (me *datatype) getRaw() string {
	return me.print()
}

func (me *datatype) error() string {
	e := me.print()
	e += " " + me.string(0)
	return e
}

func (me *datatype) print() string {
	switch me.is {
	case dataTypeUnknown:
		fallthrough
	case dataTypeString:
		fallthrough
	case dataTypeAny:
		fallthrough
	case dataTypeAnyPointer:
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
	case dataTypeVoid:
		{
			return "void"
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
			if me.variadic != nil {
				if len(me.parameters) > 0 {
					f += ","
				}
				f += "..." + me.variadic.print()
			}
			f += ") " + me.returns.print()
			return f
		}
	case dataTypeClass:
		{
			f := me.module.reference(me.class.baseClass().name)
			if len(me.generics) > 0 {
				f += genericslist(me.generics)
			}
			return f
		}
	case dataTypeEnum:
		{
			f := me.module.reference(me.enum.baseEnum().name)
			if len(me.generics) > 0 {
				f += genericslist(me.generics)
			}
			if me.union != nil {
				f += "." + me.union.name
			}
			return f
		}
	default:
		me.missingCase()
	}
	return ""
}

func newdatatype(is int) *datatype {
	d := &datatype{}
	d.is = is
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

func newdatavoid() *datatype {
	return newdatatype(dataTypeVoid)
}

func newdatastring() *datatype {
	d := newdatatype(dataTypeString)
	d.canonical = TokenString
	d.member = newdataprimitive(TokenChar)
	return d
}

func newdataprimitive(canonical string) *datatype {
	d := newdatatype(dataTypePrimitive)
	d.canonical = canonical
	d.pointer = false
	d.heap = false
	return d
}

func newdataclass(origin *hmfile, class *class, generics []*datatype) *datatype {
	d := newdatatype(dataTypeClass)
	d.origin = origin
	d.module = class.module
	d.class = class
	if generics != nil && len(generics) > 0 {
		d.generics = generics
	}
	return d
}

func newdataenum(origin *hmfile, enum *enum, union *union, generics []*datatype) *datatype {
	d := newdatatype(dataTypeEnum)
	d.origin = origin
	d.module = enum.module
	d.enum = enum
	d.union = union
	if generics != nil && len(generics) > 0 {
		d.generics = generics
	}
	return d
}

func newdataany() *datatype {
	d := newdatatype(dataTypeAny)
	d.canonical = "?"
	return d
}

func newdataanypointer() *datatype {
	d := newdatatype(dataTypeAny)
	d.canonical = "*"
	return d
}

func newdataunknown(origin *hmfile, module *hmfile, canonical string, generics []*datatype) *datatype {
	// fmt.Println("UNKNOWN:: ", canonical)
	d := newdatatype(dataTypeUnknown)
	d.origin = origin
	d.module = module
	d.canonical = canonical
	if generics != nil && len(generics) > 0 {
		d.generics = generics
	}
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

func newdatafunction(funcSig *fnSig, parameters []*datatype, variadic *datatype, returns *datatype) *datatype {
	d := newdatatype(dataTypeFunction)
	d.funcSig = funcSig
	d.parameters = parameters
	d.variadic = variadic
	d.returns = returns
	return d
}

func (me *datatype) typeSigOf(cfile *cfile, name string, mutable bool) string {
	code := ""
	switch me.is {
	case dataTypeFunction:
		code += fmtassignspace(me.returns.typeSig(cfile))
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
			code += arg.typeSig(cfile)
		}
		if me.variadic != nil {
			if len(me.parameters) > 0 {
				code += ", "
			}
			code += "..." + me.variadic.typeSig(cfile)
		}
		code += ")"
	default:
		sig := fmtassignspace(me.typeSig(cfile))
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

func (me *datatype) typeSig(cfile *cfile) string {
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
			return me.member.typeSig(cfile)
		}
	case dataTypeSlice:
		fallthrough
	case dataTypeArray:
		{
			return fmtptr(me.member.typeSig(cfile))
		}
	case dataTypeUnknown:
		fallthrough
	case dataTypeString:
		fallthrough
	case dataTypePrimitive:
		{
			if c, ok := getCName(me.canonical); ok {
				if lib, ok := typeToStd[me.canonical]; ok {
					cfile.stdReq.add(lib)
				}
				return c
			}
			return me.canonical
		}
	case dataTypeVoid:
		{
			return "void"
		}
	default:
		me.missingCase()
	}
	return ""
}

func (me *datatype) noMallocTypeSig(cfile *cfile) string {
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
			return me.member.noMallocTypeSig(cfile)
		}
	case dataTypeSlice:
		fallthrough
	case dataTypeArray:
		{
			return fmtptr(me.member.noMallocTypeSig(cfile))
		}
	case dataTypeUnknown:
		fallthrough
	case dataTypeString:
		fallthrough
	case dataTypePrimitive:
		{
			if c, ok := getCName(me.canonical); ok {
				if lib, ok := typeToStd[me.canonical]; ok {
					cfile.stdReq.add(lib)
				}
				return c
			}
			return me.canonical
		}
	default:
		me.missingCase()
	}
	return ""
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
				pop := current.name + genericsliststr(current.order)
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

func (me *datatype) merge(hint *allocHint) *datatype {
	if hint == nil {
		return me
	}
	if hint.array || hint.slice {
		me.pointer = true
	}
	me.heap = !hint.stack
	if hint.array {
		return newdataarray(strconv.Itoa(hint.size), me)
	} else if hint.slice {
		return newdataslice(me)
	}
	return me
}

func (me *datatype) getvariable() *variable {
	v := &variable{}
	v.copyData(me)
	return v
}

func (me *datatype) getnamedvariable(name string, mutable bool) *variable {
	v := me.getvariable()
	v.name = name
	v.cname = name
	v.mutable = mutable
	return v
}

func genericslist(list []*datatype) string {
	f := "<"
	for i, g := range list {
		if i > 0 {
			f += ","
		}
		f += g.print()
	}
	f += ">"
	return f
}

func genericsmap(dict map[string]*datatype) string {
	out := ""
	for k, v := range dict {
		if out != "" {
			out += ", "
		}
		out += k + ":" + v.print()
	}
	return "{" + out + "}"
}

func genericsliststr(list []string) string {
	return "<" + strings.Join(list, ",") + ">"
}

func datatypels(data []*datatype) []string {
	if data == nil {
		return nil
	}
	ls := make([]string, len(data))
	for i, d := range data {
		ls[i] = d.print()
	}
	return ls
}

func listofgenerics(module *hmfile, generics []string) ([]*datatype, *parseError) {
	order := make([]*datatype, len(generics))
	for i, g := range generics {
		var er *parseError
		order[i], er = getdatatype(module, g)
		if er != nil {
			return nil, er
		}
	}
	return order, nil
}

func copydatalist(generics []*datatype) []*datatype {
	order := make([]*datatype, len(generics))
	for i, g := range generics {
		order[i] = g.copy()
	}
	return order
}

func (me *datatype) tofnarg() *funcArg {
	return fnArgInit(me.getvariable())
}
