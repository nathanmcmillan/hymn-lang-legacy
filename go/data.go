package main

import (
	"fmt"
	"strings"
)

type idData struct {
	module *hmfile
	name   string
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
	some       *varData
	noneType   *varData
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
	me.none = in.none
	me.maybe = in.maybe
	me.some = in.some
	me.noneType = in.noneType
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

	if _, ok := attributes["use-stack"]; ok {
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

	if checkIsPrimitive(typed) {
		if typed == TokenString {
			data.isptr = true
			data.array = true
			typed = TokenChar
			data.memberType = me.typeToVarData(typed)
		} else {
			data.isptr = false
		}
		data.typed = typed
		return data
	}

	if strings.HasPrefix(typed, "maybe") {
		data.maybe = true
		data.some = me.typeToVarData(typed[6 : len(typed)-1])

	} else if strings.HasPrefix(typed, "none") {
		data.none = true
		if len(typed) > 4 {
			data.noneType = me.typeToVarData(typed[5 : len(typed)-1])
		} else {
			data.noneType = me.typeToVarData("")
		}
	}

	data.array = checkIsArray(typed)
	data.slice = checkIsSlice(typed)
	if data.array || data.slice {
		_, typed = typeOfArrayOrSlice(typed)
		data.memberType = me.typeToVarData(typed)
	}

	data.typed = typed

	dot := strings.Split(typed, ".")
	if len(dot) != 1 {
		if module, ok := me.program.hmfiles[dot[0]]; ok {
			data.module = module
			if len(dot) > 2 {
				if _, ok := me.enums[dot[1]]; ok {
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
	v.vdat = me
	return v
}

func (me *varData) merge(alloc *allocData) {
	if alloc == nil {
		return
	}

	me.array = alloc.isArray
	me.heap = !alloc.useStack

	if me.array {
		mt := me.copy()
		mt.array = false
		mt.slice = false

		me.memberType = mt
		me.full = "[]" + mt.full
		me.typed = "[]" + mt.typed
	}
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

func (me *varData) checkIsFunction() (*fnSig, bool) {
	return me.fn, me.fn != nil
}

func (me *varData) checkIsClass() (*class, bool) {
	if me.module == nil {
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

func (me *varData) equal(other *varData) bool {
	if me.full == other.full {
		return true
	}

	if en, _, ok := me.checkIsEnum(); ok {
		if en2, _, ok2 := other.checkIsEnum(); ok2 {
			if en.name == en2.name {
				return true
			}
		}

	} else if me.maybe {
		if other.maybe {
			if me.some.equal(other.some) {
				return true
			}
		} else if other.none {
			if me.some.equal(other.noneType) {
				return true
			}
		} else if me.some.equal(other) {
			return true
		}

	} else if me.none {
		if other.maybe {
			if me.noneType.equal(other.some) {
				return true
			}
		} else if other.none {
			if me.noneType == other.noneType {
				return true
			}
		} else if me.noneType.equal(other) {
			return true
		}

	} else if other.maybe {
		if other.some.equal(me) {
			return true
		}
	} else if other.none {
		if other.noneType.equal(me) {
			return true
		}
	}

	return false
}

func (me *varData) notEqual(other *varData) bool {
	return !me.equal(other)
}

func (me *varData) postfixConst() bool {
	if me.array || me.slice {
		return true
	}
	if me.maybe {
		return me.some.postfixConst()
	}
	if me.none {
		return me.noneType.postfixConst()
	}
	if _, ok := me.checkIsClass(); ok {
		return true
	}
	if _, _, ok := me.checkIsEnum(); ok {
		return true
	}
	return false
}

func (me *varData) typeSigOf(name string, mutable bool) string {
	code := ""
	if _, ok := me.checkIsFunction(); ok {
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
			code += arg.vdat.typeSig()
		}
		code += ")"

	} else {
		sig := fmtassignspace(me.typeSig())
		if mutable {
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
	if me.array || me.slice {
		return fmtptr(me.memberType.typeSig())
	}
	if me.maybe {
		return me.some.typeSig()
	}
	if me.none {
		return me.noneType.typeSig()
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

func (me *varData) replaceAny(any map[string]string) {
	f := me.full

	if m, ok := any[f]; ok {
		fmt.Println("IMPL FUNCTION REMAP NODE ::", f, "|", m)
		me.set(me.module.typeToVarData(m))
	}

	if me.array || me.slice {
	}

	if me.maybe {

	}

	if me.none {

	}
}
