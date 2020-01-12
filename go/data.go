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

type varData struct {
	hmlib      *hmlib
	module     *hmfile
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
	me.dtype = in.dtype.copy()
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

func (me *varData) print() string {
	return me.dtype.print()
}

func (me *hmfile) typeToVarDataWithAttributes(typed string, attributes map[string]string) *varData {
	data := typeToVarData(me, typed)
	if _, ok := attributes["stack"]; ok {
		data.onStack = true
		data.dtype.setIsOnStack(true)
	}
	return data
}

func (me *hmlib) literalType(typed string) *varData {
	data := &varData{}
	data.hmlib = me
	data.dtype = getdatatype(nil, typed)
	return data
}

func (me *varData) plain() *plainType {
	return me.dtype.plain()
}

func typeToVarData(module *hmfile, typed string) *varData {

	if module != nil && module.scope.fn != nil && module.scope.fn.aliasing != nil {
		if alias, ok := module.scope.fn.aliasing[typed]; ok {
			typed = alias
		}
	}

	data := &varData{}
	data.dtype = getdatatype(module, typed)
	data.module = module
	data.mutable = true
	data.isptr = true
	data.heap = true

	dot := strings.Split(typed, ".")
	if len(dot) != 1 {
		if newmodule, ok := module.imports[dot[0]]; ok {
			data.module = newmodule
			typed = strings.Join(dot[1:], ".")
		}
	}

	if checkIsPrimitive(typed) {
		if typed == TokenString {
			data.array = true
			data.isptr = true
			data.memberType = typeToVarData(module, TokenChar)
		} else {
			data.isptr = false
			data.onStack = true
		}
		return data
	}

	if strings.HasPrefix(typed, "maybe<") {
		data.maybe = true
		data.memberType = typeToVarData(module, typed[6:len(typed)-1])
		return data

	} else if strings.HasPrefix(typed, "none<") {
		if len(typed) > 4 {
			data.memberType = typeToVarData(module, typed[5:len(typed)-1])
			typed = "maybe" + typed[4:len(typed)]
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

	dot = strings.Split(typed, ".")
	if len(dot) != 1 {
		if _, ok := data.module.enums[dot[0]]; ok {
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
			me.dtype = newdataarray(strconv.Itoa(hint.size), me.dtype)
		} else {
			me.dtype = newdataslice(me.dtype)
		}
	}
	me.dtype.heap = me.heap
	me.dtype.pointer = me.isptr
}

func (me *varData) sizeOfArray() string {
	return me.dtype.arraySize()
}

func (me *varData) arrayToSlice() {
	me.array = false
	me.slice = true
	me.dtype.convertArrayToSlice()
}

func (me *varData) checkIsSomeOrNone() bool {
	return me.dtype.isSomeOrNone()
}

func (me *varData) checkIsString() bool {
	return me.dtype.isString()
}

func (me *varData) checkIsChar() bool {
	return me.dtype.isChar()
}

func (me *varData) checkIsArray() bool {
	return me.dtype.isArray()
}

func (me *varData) checkIsSlice() bool {
	return me.dtype.isSlice()
}

func (me *varData) checkIsArrayOrSlice() bool {
	return me.dtype.isArrayOrSlice()
}

func (me *varData) checkIsIndexable() bool {
	return me.dtype.isIndexable()
}

func (me *varData) checkIsPointerInC() bool {
	return me.dtype.isPointerInC()
}

func (me *varData) checkIsPrimitive() bool {
	return me.dtype.isPrimitive()
}

func (me *varData) checkIsClass() (*class, bool) {
	return me.dtype.isClass()
}

func (me *varData) checkIsEnum() (*enum, *union, bool) {
	return me.dtype.isEnum()
}

func (me *varData) getFunction(name string) (*function, bool) {
	return me.dtype.isFunction(name)
}

func (me *varData) postfixConst() bool {
	return me.dtype.postfixConst()
}

func (me *varData) noConst() bool {
	return me.dtype.noConst()
}

func (me *varData) setOnStackNotPointer() {
	me.isptr = false
	me.onStack = true
	me.dtype.setOnStackNotPointer()
}

func (me *varData) setIsPointer(flag bool) {
	me.isptr = flag
	me.dtype.setIsPointer(flag)
}

func (me *varData) memPtr() string {
	return me.dtype.memoryGet()
}

func (me *varData) typeSigOf(name string, mutable bool) string {
	return me.dtype.typeSigOf(name, mutable)
}

func (me *varData) typeSig() string {
	return me.dtype.typeSig()
}

func (me *varData) noMallocTypeSig() string {
	return me.dtype.noMallocTypeSig()
}

func (me *varData) typeEqual(b *varData) bool {
	return me.dtype.equals(b.dtype)
}

func (me *varData) notEqual(b *varData) bool {
	return me.dtype.notEquals(b.dtype)
}

func (me *varData) isQuestion() bool {
	return me.dtype.isQuestion()
}

func (me *varData) isVoid() bool {
	return me.dtype.isVoid()
}

func (me *varData) isPointer() bool {
	return me.dtype.isPointer()
}

func (me *varData) cname() string {
	return me.dtype.cname()
}

func (me *varData) isInt() bool {
	return me.dtype.isInt()
}
