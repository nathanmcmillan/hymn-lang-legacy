package main

import (
	"strings"
)

type variable struct {
	typed   string
	name    string
	dfault  string
	mutable bool
	isptr   bool
	cName   string
	vdat    *varData
}

func (me *hmfile) varInitFromData(vdat *varData, name string, mutable, isptr bool) *variable {
	v := &variable{}
	v.vdat = vdat
	v.typed = vdat.full
	v.name = name
	v.cName = name
	v.mutable = mutable
	v.isptr = isptr
	v.vdat.isptr = v.isptr
	return v
}

func (me *hmfile) varInit(typed, name string, mutable, isptr bool) *variable {
	v := &variable{}
	v.name = name
	v.cName = name
	v.mutable = mutable
	v.isptr = isptr
	v.update(me, typed)
	return v
}

func (me *hmfile) varWithDefaultInit(typed, name string, mutable, isptr bool, dfault string) *variable {
	v := me.varInit(typed, name, mutable, isptr)
	v.dfault = dfault
	return v
}

func (me *variable) update(module *hmfile, typed string) {
	me.typed = typed
	me.vdat = module.typeToVarData(typed)
	me.vdat.isptr = me.isptr
}

func (me *variable) copy() *variable {
	v := &variable{}
	v.typed = me.typed
	v.name = me.name
	v.cName = me.name
	v.mutable = me.mutable
	v.isptr = me.isptr
	v.vdat = me.vdat
	return v
}

type idData struct {
	module *hmfile
	name   string
}

type callData struct {
	module *hmfile
	name   string
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
	some        string
	noneType    string
	typeInArray *varData
	en          *enum
	un          *union
	cl          *class
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
	return v
}

func dataInit(module *hmfile, typed string, mutable, isptr, heap bool) *varData {
	d := &varData{}
	d.module = module
	d.typed = typed
	d.mutable = mutable
	d.isptr = isptr
	d.heap = heap
	return d
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
		data.some = typed[6 : len(typed)-1]

	} else if strings.HasPrefix(typed, "none") {
		data.none = true
		data.noneType = typed[5 : len(typed)-1]
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

func (me *varData) checkIsUnion() (*enum, bool) {
	dot := strings.Split(me.typed, ".")
	if len(dot) != 1 {
		en, ok := me.module.enums[dot[0]]
		if ok && en.simple {
			return en, true
		}
		return nil, false
	}
	en, ok := me.module.enums[me.typed]
	if ok && en.simple {
		return en, true
	}
	return nil, false
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
			if me.some == other.some {
				return true
			}
		} else if other.none {
			if me.some == other.noneType {
				return true
			}
		} else if me.some == other.full {
			return true
		}

	} else if me.none {
		if other.maybe {
			if me.noneType == other.some {
				return true
			}
		} else if other.none {
			if me.noneType == other.noneType {
				return true
			}
		} else if me.noneType == other.full {
			return true
		}

	} else if other.maybe {
		if other.some == me.full {
			return true
		}
	} else if other.none {
		if other.noneType == me.full {
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
		return me.module.typeToVarData(me.some).postfixConst()
	}
	if me.none {
		return me.module.typeToVarData(me.noneType).postfixConst()
	}
	if _, ok := me.checkIsClass(); ok {
		return true
	}
	if _, ok := me.checkIsUnion(); ok {
		return true
	}
	return false
}

func (me *varData) typeSig() string {
	if me.array {
		return fmtptr(me.typeInArray.typeSig())
	}
	if me.maybe {
		return me.module.typeToVarData(me.some).typeSig()
	}
	if me.none {
		return me.module.typeToVarData(me.noneType).typeSig()
	}
	if _, ok := me.checkIsClass(); ok {
		sig := me.module.classNameSpace(me.typed)
		if !me.onStack && me.isptr {
			sig += " *"
		}
		return sig
	} else if en, _, ok := me.checkIsEnum(); ok {
		return en.typeSig()
	} else if me.full == "string" {
		return "char *"
	}
	return me.full
}

func (me *varData) noMallocTypeSig() string {
	if me.array {
		return fmtptr(me.typeInArray.noMallocTypeSig())
	}
	if _, ok := me.checkIsClass(); ok {
		return me.module.classNameSpace(me.typed)
	} else if en, _, ok := me.checkIsEnum(); ok {
		return en.noMallocTypeSig()
	} else if me.full == "string" {
		return "char *"
	}
	return me.full
}

func (me *varData) memPtr() string {
	if me.isptr {
		return "->"
	}
	return "."
}
