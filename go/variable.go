package main

type variable struct {
	name    string
	mutable bool
	isptr   bool
	cName   string
	vdat    *varData
}

func (me *hmfile) varInitFromData(vdat *varData, name string, mutable, isptr bool) *variable {
	v := &variable{}
	v.vdat = vdat
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

func (me *variable) update(module *hmfile, typed string) {
	me.vdat = module.typeToVarData(typed)
	me.vdat.isptr = me.isptr
}

func (me *variable) updateFromVar(module *hmfile, data *varData) {
	me.vdat = data
	me.vdat.isptr = me.isptr
}

func (me *variable) copy() *variable {
	v := &variable{}
	v.name = me.name
	v.cName = me.name
	v.mutable = me.mutable
	v.isptr = me.isptr
	v.vdat = me.vdat
	return v
}
