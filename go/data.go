package main

import (
	"fmt"
	"strings"
)

type varData struct {
	module  *hmfile
	typed   string
	full    string
	mutable bool
	pointer bool
	heap    bool
}

func dataInit(module *hmfile, typed string, mutable, pointer, heap bool) *varData {
	d := &varData{}
	d.module = module
	d.typed = typed
	d.mutable = mutable
	d.pointer = pointer
	d.heap = heap
	return d
}

func (me *hmfile) typeToVarData(typed string) *varData {
	data := &varData{}
	data.full = typed
	data.mutable = true
	data.pointer = true
	data.heap = true

	if checkIsArray(typed) {
		typed = typeOfArray(typed)
	}

	if typed[0] == '$' {
		data.pointer = false
		typed = typed[1:]
	} else if typed[0] == '\\' {
		data.heap = false
		typed = typed[1:]
	}

	data.module = me
	data.typed = typed

	dot := strings.Split(typed, ".")
	if len(dot) != 1 {
		fmt.Println("COMPARE::", typed)
		if module, ok := me.program.hmfiles[dot[0]]; ok {
			data.module = module
			if len(dot) > 2 {
				if _, ok := me.enums[dot[1]]; ok {
					data.typed = dot[1] + dot[2]
				} else {
					panic("unknown type \"" + typed + "\"")
				}
			} else {
				data.typed = dot[1]
			}
		} else if _, ok := me.enums[dot[0]]; ok {
			data.typed = dot[0] + dot[1]
		} else {
			panic("unknown type \"" + typed + "\"")
		}
	}

	return data
}

func (me *varData) checkIsArray() bool {
	return strings.HasPrefix(me.full, "[]")
}

func (me *varData) checkIsClass() (*class, bool) {
	cl, ok := me.module.classes[me.typed]
	return cl, ok
}

func (me *varData) checkIsEnum() (*enum, bool) {
	en, ok := me.module.enums[me.typed]
	return en, ok
}

func (me *varData) checkIsUnion() (*enum, bool) {
	en, ok := me.module.enums[me.typed]
	if ok && en.simple {
		return en, true
	}
	return nil, false
}

func (me *varData) postfixConst() bool {
	if me.checkIsArray() {
		return true
	}
	if _, ok := me.checkIsClass(); ok {
		return true
	}
	if _, ok := me.checkIsUnion(); ok {
		return true
	}
	return false
}
