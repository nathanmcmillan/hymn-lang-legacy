package main

import (
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
	module      *hmfile
	typed       string
	full        string
	mutable     bool
	onStack     bool
	isptr       bool
	heap        bool
	array       bool
	none        bool
	maybe       bool
	some        *varData
	noneType    *varData
	typeInArray *varData
	en          *enum
	un          *union
	cl          *class
	fn          *fnSig
}

func (me *varData) copy() *varData {
	v := &varData{}
	v.module = me.module
	v.typed = me.typed
	v.full = me.full
	v.mutable = me.mutable
	v.onStack = me.onStack
	v.isptr = me.isptr
	v.heap = me.heap
	v.array = me.array
	v.none = me.none
	v.maybe = me.maybe
	v.some = me.some
	v.noneType = me.noneType
	v.typeInArray = me.typeInArray
	v.en = me.en
	v.un = me.un
	v.cl = me.cl
	v.fn = me.fn
	return v
}

func (me *hmfile) typeToVarDataWithAttributes(typed string, attributes map[string]string) *varData {
	data := me.typeToVarData(typed)

	if _, ok := attributes["use-stack"]; ok {
		data.onStack = true
	}

	return data
}

func (me *hmfile) typeToVarData(typed string) *varData {
	data := &varData{}
	data.full = typed
	data.mutable = true
	data.isptr = true
	data.heap = true

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
	if data.array {
		typed = typeOfArray(typed)
		data.typeInArray = me.typeToVarData(typed)
	}

	data.module = me
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
		typeInArray := me.copy()
		typeInArray.array = false

		me.typeInArray = typeInArray
		me.full = "[]" + typeInArray.full
		me.typed = "[]" + typeInArray.typed
	}
}

func (me *varData) checkIsArray() bool {
	return strings.HasPrefix(me.full, "[]")
}

func (me *varData) checkIsFunction() (*fnSig, bool) {
	return me.fn, me.fn != nil
}

func (me *varData) checkIsClass() (*class, bool) {
	cl, ok := me.module.classes[me.typed]
	return cl, ok
}

func (me *varData) checkIsEnum() (*enum, *union, bool) {
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
	if me.array {
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

func primitiveC(primitive string) string {
	switch primitive {
	case TokenFloat32:
		return "float"
	case TokenFloat64:
		return "double"
	case TokenString:
		return "char *"
	case TokenInt8:
		return "int8_t"
	case TokenInt16:
		return "int16_t"
	case TokenInt32:
		return "int32_t"
	case TokenInt64:
		return "int64_t"
	case TokenUInt:
		return "unsigned int"
	case TokenUInt8:
		return "uint8_t"
	case TokenUInt16:
		return "uint16_t"
	case TokenUInt32:
		return "uint32_t"
	case TokenUInt64:
		return "uint64_t"
	}
	return primitive
}

func (me *varData) typeSig() string {
	if me.array {
		return fmtptr(me.typeInArray.typeSig())
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
	return primitiveC(me.full)
}

func (me *varData) noMallocTypeSig() string {
	if me.array {
		return fmtptr(me.typeInArray.noMallocTypeSig())
	}
	if _, ok := me.checkIsClass(); ok {
		return me.module.classNameSpace(me.typed)
	} else if en, _, ok := me.checkIsEnum(); ok {
		return en.noMallocTypeSig()
	}
	return primitiveC(me.full)
}

func (me *varData) memPtr() string {
	if me.isptr {
		return "->"
	}
	return "."
}
