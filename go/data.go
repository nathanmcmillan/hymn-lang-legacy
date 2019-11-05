package main

import (
	"fmt"
	"strconv"
	"strings"
)

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
	me.mutable = in.mutable
	me.onStack = in.onStack
	me.isptr = in.isptr
	me.heap = in.heap
	me.array = in.array
	me.slice = in.slice
	me.none = in.none
	me.maybe = in.maybe
	me.memberType = in.memberType
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
	data := me.typeToVarData(typed)

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
	return data
}

func (me *hmfile) typeToVarData(typed string) *varData {
	data := &varData{}
	data.full = typed
	data.mutable = true
	data.isptr = true
	data.heap = true
	data.module = me

	dtype := me.getdatatype(typed)
	fmt.Println("\"" + typed + "\" USE INSTEAD OF FULL :: \"" + dtype.print() + "\"")

	if checkIsPrimitive(typed) {
		if typed == TokenString {
			data.array = true
			data.isptr = true
			typed = TokenChar
			data.memberType = me.typeToVarData(typed)
		} else {
			data.isptr = false
			data.onStack = true
		}
		data.typed = typed
		return data
	}

	if strings.HasPrefix(typed, "maybe") {
		data.maybe = true
		data.memberType = me.typeToVarData(typed[6 : len(typed)-1])

	} else if strings.HasPrefix(typed, "none") {
		data.none = true
		if len(typed) > 4 {
			data.memberType = me.typeToVarData(typed[5 : len(typed)-1])
		} else {
			data.memberType = me.typeToVarData("")
		}
	}

	data.array = checkIsArray(typed)
	data.slice = checkIsSlice(typed)
	if data.array || data.slice {
		data.isptr = true
		_, typed = typeOfArrayOrSlice(typed)
		data.memberType = me.typeToVarData(typed)
	}

	if checkIsFunction(typed) {
		args, ret := functionSigType(typed)
		fn := fnSigInit(me)
		for _, a := range args {
			t := me.typeToVarData(a)
			fn.args = append(fn.args, fnArgInit(t.asVariable()))
		}
		fn.typed = me.typeToVarData(ret)
		data.fn = fn
	}

	data.typed = typed

	dot := strings.Split(typed, ".")
	if len(dot) != 1 {
		if module, ok := me.program.hmfiles[dot[0]]; ok {
			data.module = module
			if len(dot) > 2 {
				if _, ok := module.enums[dot[1]]; ok {
					data.typed = dot[1] + "." + dot[2]
				} else {
					panic("unknown type \"" + typed + "\"")
				}
			} else {
				data.typed = dot[1]
			}
		} else if _, ok := me.enums[dot[0]]; ok {
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
		code += fmtassignspace(sig.typed.typeSig())
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

func getCName(primitive string) string {
	if name, ok := typeToCName[primitive]; ok {
		return name
	}
	return primitive
}

func (me *varData) typeSig() string {
	if me.checkIsArrayOrSlice() {
		return fmtptr(me.memberType.typeSig())
	}
	if me.checkIsSomeOrNone() {
		return me.memberType.typeSig()
	}
	if _, ok := me.checkIsClass(); ok {
		sig := me.module.classNameSpace(me.typed)
		if !me.onStack && me.isptr {
			sig += " *"
		}
		return sig
	} else if en, _, ok := me.checkIsEnum(); ok {
		return en.typeSig()
	}
	return getCName(me.full)
}

func (me *varData) noMallocTypeSig() string {
	if me.array || me.slice {
		return fmtptr(me.memberType.noMallocTypeSig())
	}
	if _, ok := me.checkIsClass(); ok {
		return me.module.classNameSpace(me.typed)
	} else if en, _, ok := me.checkIsEnum(); ok {
		return en.noMallocTypeSig()
	}
	return getCName(me.full)
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

func (me *varData) genericReplace(any map[string]string) {
	f := me.full

	if m, ok := any[f]; ok {
		me.set(me.module.typeToVarData(m))
		return
	}

	if me.checkIsArrayOrSlice() {
		me.memberType.genericReplace(any)
		return
	}

	if me.checkIsSomeOrNone() {
		me.memberType.genericReplace(any)
		return
	}

	if me.fn != nil {
		for _, a := range me.fn.args {
			a.data().genericReplace(any)
		}
		me.fn.typed.genericReplace(any)
		return
	}
}

func (me *varData) typeEqual(other *varData) bool {
	if me.full == other.full {
		return true
	}
	a := me.module.typeStandard(me.full)
	b := other.module.typeStandard(other.full)
	return a == b
}

func (me *varData) notEqual(other *varData) bool {
	return !me.typeEqual(other)
}

func (me *parser) typeEqual(one, two string) bool {
	if one == two {
		return true
	}
	return me.typeStandard(one) == me.typeStandard(two)
}

func (me *parser) typeStandard(typed string) string {
	return me.hmfile.typeStandard(typed)
}

func (me *hmfile) typeStandard(typed string) string {

	if checkIsPrimitive(typed) {
		return typed
	}

	if strings.HasPrefix(typed, "maybe") {
		return me.typeStandard(typed[6 : len(typed)-1])
	} else if strings.HasPrefix(typed, "none") {
		return me.typeStandard(typed[5 : len(typed)-1])
	}

	if checkIsArray(typed) || checkIsSlice(typed) {
		size, member := typeOfArrayOrSlice(typed)
		return "[" + size + "]" + me.typeStandard(member)
	}

	if checkIsFunction(typed) {
		args, ret := functionSigType(typed)
		for i, a := range args {
			args[i] = me.typeStandard(a)
		}
		return "(" + strings.Join(args, ", ") + ") " + me.typeStandard(ret)
	}

	dot := strings.Split(typed, ".")
	if len(dot) != 1 {
		if _, ok := me.imports[dot[0]]; ok {
			fmt.Println("FIRST DOT IS FROM A CLASS")
		} else {
			fmt.Println("FIRST DOT IS FOR AN ENUM")
		}
	}

	g := strings.Index(typed, "<")
	if g != -1 {
		return me.typeStandardBrackets(typed)
	}

	// baseClass, okc := me.hmfile.classes[base]
	// baseEnum, oke := me.hmfile.enums[base]

	return typed
}

func (me *hmfile) typeStandardBrackets(typed string) string {

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

	fmt.Println("BRACKETS ::", order)
	return ""
}
