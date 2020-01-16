package main

import (
	"strconv"
	"strings"
)

type varData struct {
	module     *hmfile
	dtype      *datatype
	memberType *varData
}

func (me *varData) set(in *varData) {
	me.module = in.module
	me.dtype = in.dtype.copy()
	if in.memberType != nil {
		me.memberType = in.memberType.copy()
	}
}

func (me *varData) copy() *varData {
	v := &varData{}
	v.set(me)
	return v
}

func (me *varData) print() string {
	return me.dtype.print()
}

func (me *varData) getRaw() string {
	return me.print()
}

func functionSigToVarData(fsig *fnSig) *varData {
	sig := fsig.print()
	d := &varData{}
	d.module = fsig.module
	d.dtype = getdatatype(nil, sig)
	d.dtype.module = fsig.module
	return d
}

func (me *hmfile) typeToVarDataWithAttributes(typed string, attributes map[string]string) *varData {
	data := typeToVarData(me, typed)
	if _, ok := attributes["stack"]; ok {
		data.dtype.setIsOnStack(true)
	}
	return data
}

func (me *hmlib) literalType(typed string) *varData {
	data := &varData{}
	data.dtype = getdatatype(nil, typed)
	return data
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

	dot := strings.Split(typed, ".")
	if len(dot) != 1 {
		if newmodule, ok := module.imports[dot[0]]; ok {
			data.module = newmodule
			typed = strings.Join(dot[1:], ".")
		}
	}

	if checkIsPrimitive(typed) {
		if typed == TokenString {
			data.memberType = typeToVarData(module, TokenChar)
		}
		return data
	}

	if strings.HasPrefix(typed, "maybe<") {
		data.memberType = typeToVarData(module, typed[6:len(typed)-1])
		return data

	} else if strings.HasPrefix(typed, "none<") {
		if len(typed) > 4 {
			data.memberType = typeToVarData(module, typed[5:len(typed)-1])
			typed = "maybe" + typed[4:len(typed)]
			data.dtype = getdatatype(module, typed)
		} else {
			data.memberType = typeToVarData(module, "")
		}
		return data
	}

	array := checkIsArray(typed)
	slice := checkIsSlice(typed)
	if array || slice {
		_, typed = typeOfArrayOrSlice(typed)
		data.memberType = typeToVarData(module, typed)
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
	if hint.array || hint.slice {
		me.dtype.pointer = true
	}
	me.dtype.heap = !hint.stack
	if hint.array || hint.slice {
		member := me.copy()
		me.memberType = member
		if hint.array {
			me.dtype = newdataarray(strconv.Itoa(hint.size), me.dtype)
		} else {
			me.dtype = newdataslice(me.dtype)
		}
	}
}

func (me *varData) sizeOfArray() string {
	return me.dtype.arraySize()
}

func (me *varData) arrayToSlice() {
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

func (me *varData) isArray() bool {
	return me.dtype.isArray()
}

func (me *varData) checkIsSlice() bool {
	return me.dtype.isSlice()
}

func (me *varData) checkIsArrayOrSlice() bool {
	return me.dtype.isArrayOrSlice()
}

func (me *varData) isArrayOrSlice() bool {
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
	return me.dtype.getFunction(name)
}

func (me *varData) postfixConst() bool {
	return me.dtype.postfixConst()
}

func (me *varData) noConst() bool {
	return me.dtype.noConst()
}

func (me *varData) setOnStackNotPointer() {
	me.dtype.setOnStackNotPointer()
}

func (me *varData) setIsPointer(flag bool) {
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

func (me *varData) isAnyIntegerType() bool {
	return me.dtype.isAnyIntegerType()
}

func (me *varData) isInt() bool {
	return me.dtype.isInt()
}

func (me *varData) isNumber() bool {
	return me.dtype.isNumber()
}

func (me *varData) isBoolean() bool {
	return me.dtype.isBoolean()
}

func (me *varData) isSome() bool {
	return me.dtype.isSome()
}

func (me *varData) isNone() bool {
	return me.dtype.isNone()
}

func (me *varData) isOnStack() bool {
	return me.dtype.isOnStack()
}

func (me *varData) equals(b *varData) bool {
	return me.dtype.equals(b.dtype)
}

func (me *varData) functionSignature() *fnSig {
	return me.dtype.functionSignature()
}

func (me *varData) getmember() *varData {
	return me.memberType
}

func (me *varData) getmodule() *hmfile {
	// return me.module
	return me.dtype.module
}
