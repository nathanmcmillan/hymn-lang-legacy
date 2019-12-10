package main

import (
	"strconv"
	"strings"
)

type plainType struct {
	module *hmfile
	typed  string
}

func (me *plainType) print() string {
	return me.module.name + "." + me.typed
}

type idData struct {
	module *hmfile
	name   string
}

func (me *idData) copy() *idData {
	i := &idData{}
	i.module = me.module
	i.name = me.name
	return i
}

func (me *idData) string() string {
	return me.module.name + "." + me.name
}

type varData struct {
	hmlib      *hmlib
	module     *hmfile
	typed      string
	full       string
	dtype      *datatype
	mutable    bool
	onStack    bool
	isptr      bool
	heap       bool
	array      bool
	slice      bool
	none       bool
	maybe      bool
	memberType *varData
	en         *enum
	un         *union
	cl         *class
	fn         *fnSig
}

func (me *varData) set(in *varData) {
	me.module = in.module
	me.typed = in.typed
	me.full = in.full
	me.dtype = in.dtype
	me.mutable = in.mutable
	me.onStack = in.onStack
	me.isptr = in.isptr
	me.heap = in.heap
	me.array = in.array
	me.slice = in.slice
	me.none = in.none
	me.maybe = in.maybe
	if in.memberType != nil {
		me.memberType = in.memberType.copy()
	}
	me.en = in.en
	me.un = in.un
	me.cl = in.cl
	me.fn = in.fn
}

func (me *varData) copy() *varData {
	v := &varData{}
	v.set(me)
	return v
}

func (me *hmfile) typeToVarDataWithAttributes(typed string, attributes map[string]string) *varData {
	data := typeToVarData(me, typed)

	if _, ok := attributes["stack"]; ok {
		data.onStack = true
	}

	return data
}

func (me *hmlib) literalType(typed string) *varData {
	data := &varData{}
	data.full = typed
	data.typed = typed
	data.hmlib = me
	data.dtype = getdatatype(nil, typed)
	return data
}

func (me *varData) plain() *plainType {
	return &plainType{me.module, me.full}
}

func typeToVarData(module *hmfile, typed string) *varData {

	if module != nil && module.scope.fn != nil && module.scope.fn.aliasing != nil {
		if alias, ok := module.scope.fn.aliasing[typed]; ok {
			typed = alias
		}
	}

	data := &varData{}
	data.full = typed
	data.typed = typed
	data.mutable = true
	data.isptr = true
	data.heap = true
	data.module = module
	data.dtype = getdatatype(module, typed)

	if checkIsPrimitive(typed) {
		if typed == TokenString {
			data.array = true
			data.isptr = true
			data.memberType = typeToVarData(module, TokenChar)
			data.typed = TokenChar
		} else {
			data.isptr = false
			data.onStack = true
		}
		return data
	}

	if strings.HasPrefix(typed, "maybe") {
		data.maybe = true
		data.memberType = typeToVarData(module, typed[6:len(typed)-1])
		return data

	} else if strings.HasPrefix(typed, "none") {
		if len(typed) > 4 {
			data.memberType = typeToVarData(module, typed[5:len(typed)-1])
			typed = "maybe" + typed[4:len(typed)]
			data.full = typed
			data.typed = typed
			data.dtype = getdatatype(module, typed)
			data.maybe = true
		} else {
			data.none = true
			data.memberType = typeToVarData(module, "")
		}
		return data
	}

	data.array = checkIsArray(typed)
	data.slice = checkIsSlice(typed)
	if data.array || data.slice {
		data.isptr = true
		_, typed = typeOfArrayOrSlice(typed)
		data.memberType = typeToVarData(module, typed)
	}

	if checkIsFunction(typed) {
		args, ret := functionSigType(typed)
		fn := fnSigInit(module)
		for _, a := range args {
			t := typeToVarData(module, a)
			fn.args = append(fn.args, fnArgInit(t.asVariable()))
		}
		fn.returns = typeToVarData(module, ret)
		data.fn = fn
	}

	data.typed = typed

	dot := strings.Split(typed, ".")
	if len(dot) != 1 {
		if module2, ok := module.program.hmfiles[dot[0]]; ok {
			data.module = module2
			if len(dot) > 2 {
				if _, ok := module2.enums[dot[1]]; ok {
					data.typed = dot[1] + "." + dot[2]
				} else {
					panic("unknown type \"" + typed + "\"")
				}
			} else {
				data.typed = dot[1]
			}
		} else if _, ok := module.enums[dot[0]]; ok {
			data.typed = dot[0] + "." + dot[1]
		} else {
			panic("unknown type \"" + typed + "\"")
		}
	}

	return data
}

func (me *varData) asVariable() *variable {
	v := &variable{}
	v.copyData(me)
	return v
}

func (me *varData) merge(hint *allocData) {
	if hint == nil {
		return
	}
	me.array = hint.array
	me.slice = hint.slice
	if me.array || me.slice {
		me.isptr = true
	}
	me.heap = !hint.stack
	if me.array || me.slice {
		member := me.copy()
		member.array = false
		member.slice = false
		me.memberType = member
		if me.array {
			size := "[" + strconv.Itoa(hint.size) + "]"
			me.full = size + member.full
			me.typed = size + member.typed
		} else {
			me.full = "[]" + member.full
			me.typed = "[]" + member.typed
		}
	}
}

func (me *varData) sizeOfArray() string {
	i := strings.Index(me.full, "]")
	return me.full[1:i]
}

func (me *varData) arrayToSlice() {
	me.array = false
	me.slice = true
	index := strings.Index(me.full, "]")
	me.full = "[]" + me.full[index+1:]
	me.typed = me.full
}

func (me *varData) checkIsSomeOrNone() bool {
	return me.maybe || me.none
}

func (me *varData) checkIsString() bool {
	return me.full == TokenString || me.full == "[]char"
}

func (me *varData) checkIsChar() bool {
	return me.full == TokenChar
}

func (me *varData) checkIsArray() bool {
	return checkIsArray(me.full)
}

func (me *varData) checkIsSlice() bool {
	return checkIsSlice(me.full)
}

func (me *varData) checkIsArrayOrSlice() bool {
	return me.array || me.slice
}

func (me *varData) checkIsPointerInC() bool {
	if me.checkIsPrimitive() {
		return false
	}
	return me.isptr
}

func checkIsPrimitive(t string) bool {
	_, ok := primitives[t]
	return ok
}

func (me *varData) checkIsPrimitive() bool {
	return checkIsPrimitive(me.full)
}

func (me *varData) checkIsClass() (*class, bool) {
	if me.module == nil {
		if me.hmlib == nil {
			return nil, false
		}
		cl, ok := me.hmlib.classes[me.typed]
		return cl, ok
	}
	cl, ok := me.module.classes[me.typed]
	return cl, ok
}

func (me *varData) checkIsEnum() (*enum, *union, bool) {
	if me.module == nil {
		return nil, nil, false
	}
	dot := strings.Split(me.typed, ".")
	if len(dot) != 1 {
		en, ok := me.module.enums[dot[0]]
		un, _ := en.types[dot[1]]
		return en, un, ok
	}
	en, ok := me.module.enums[me.typed]
	return en, nil, ok
}

func (me *varData) postfixConst() bool {
	if me.checkIsArrayOrSlice() {
		return true
	}
	if me.checkIsSomeOrNone() {
		return me.memberType.postfixConst()
	}
	if _, ok := me.checkIsClass(); ok {
		return true
	}
	if _, _, ok := me.checkIsEnum(); ok {
		return true
	}
	return false
}

func (me *varData) noConst() bool {
	if !me.checkIsPrimitive() {
		if me.onStack || !me.isptr {
			return true
		}
	}
	return false
}

func (me *varData) typeSigOf(name string, mutable bool) string {
	code := ""
	if me.fn != nil {
		sig := me.fn
		code += fmtassignspace(sig.returns.typeSig())
		code += "(*"
		if !mutable {
			code += "const "
		}
		code += name
		code += ")("
		for ix, arg := range sig.args {
			if ix > 0 {
				code += ", "
			}
			code += arg.data().typeSig()
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

func getCName(primitive string) (string, bool) {
	if name, ok := typeToCName[primitive]; ok {
		return name, true
	}
	return primitive, false
}

func (me *varData) typeSig() string {
	if cname, ok := getCName(me.full); ok {
		return cname
	}
	if me.checkIsArrayOrSlice() {
		return fmtptr(me.memberType.typeSig())
	}
	if me.checkIsSomeOrNone() {
		return me.memberType.typeSig()
	}
	if _, ok := me.checkIsClass(); ok {
		sig := me.module.classNameSpace(me.dtype.cname())
		if !me.onStack && me.isptr {
			sig += " *"
		}
		return sig
	} else if en, _, ok := me.checkIsEnum(); ok {
		return en.typeSig()
	}
	return me.full
}

func (me *varData) noMallocTypeSig() string {
	if cname, ok := getCName(me.full); ok {
		return cname
	}
	if me.checkIsArrayOrSlice() {
		return fmtptr(me.memberType.noMallocTypeSig())
	}
	if _, ok := me.checkIsClass(); ok {
		return me.module.classNameSpace(me.dtype.cname())
	} else if en, _, ok := me.checkIsEnum(); ok {
		return en.noMallocTypeSig()
	}
	return me.full
}

func (me *varData) memPtr() string {
	if me.isptr {
		return "->"
	}
	return "."
}

func (me *varData) getFunction(name string) (*function, bool) {
	if me.module != nil {
		f, ok := me.module.getFunction(name)
		return f, ok
	}
	f, ok := me.hmlib.functions[name]
	return f, ok
}

func (me *varData) typeEqual(b *varData) bool {
	if me.full == b.full {
		return true
	}
	if strings.HasPrefix(me.full, "maybe<") && b.full == "none" {
		return true
	}
	if strings.HasPrefix(b.full, "maybe<") && me.full == "none" {
		return true
	}
	return me.dtype.standard() == b.dtype.standard()
}

func (me *varData) notEqual(other *varData) bool {
	return !me.typeEqual(other)
}

func (me *parser) typeEqual(one, two string) bool {
	if one == two {
		return true
	}
	if strings.HasPrefix(one, "maybe<") && two == "none" {
		return true
	}
	if strings.HasPrefix(two, "maybe<") && one == "none" {
		return true
	}
	return getdatatype(me.hmfile, one).standard() == getdatatype(me.hmfile, two).standard()
}
