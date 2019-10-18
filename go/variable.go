package main

type variable struct {
	name    string
	mutable bool
	isptr   bool
	cName   string
	_vdata  *varData
}

func (me *variable) data() *varData {
	return me._vdata
}

func (me *variable) copyData(data *varData) {
	me._vdata = data.copy()
}

func (me *hmfile) varInitFromData(data *varData, name string, mutable, isptr bool) *variable {
	v := &variable{}
	v.copyData(data)
	v.name = name
	v.cName = name
	v.mutable = mutable
	v.isptr = isptr
	v.data().isptr = v.isptr
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

func (me *variable) update(module *hmfile, typed string) {
	me.copyData(module.typeToVarData(typed))
	me.data().isptr = me.isptr
}

func (me *variable) updateFromVar(module *hmfile, data *varData) {
	me.copyData(data)
	me.data().isptr = me.isptr
}

func (me *variable) copy() *variable {
	v := &variable{}
	v.name = me.name
	v.cName = me.name
	v.mutable = me.mutable
	v.isptr = me.isptr
	v.copyData(me.data())
	return v
}
