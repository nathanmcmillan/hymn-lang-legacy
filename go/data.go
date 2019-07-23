package main

import "strings"

type varData struct {
	module   *hmfile
	typed    string
	longType string
	mutable  bool
	pointer  bool
	heap     bool
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
	data.longType = typed
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

	dot := strings.Split(typed, ".")
	if len(dot) == 1 {
		data.module = me
		data.typed = typed
	} else {
		module := me.program.hmfiles[dot[0]]
		data.module = module
		data.typed = dot[1]
	}

	return data
}

// func (me *hmfile) moduleAndName(name string) (*hmfile, string) {
// 	if checkIsArray(name) {
// 		name = typeOfArray(name)
// 	}
// 	get := strings.Split(name, ".")
// 	if len(get) == 1 {
// 		return me, get[0]
// 	}
// 	module := me.program.hmfiles[get[0]]
// 	return module, get[1]
// }
