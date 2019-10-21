package main

type variable struct {
	name    string
	cName   string
	mutable bool
	_vdata  *varData
}

func (me *variable) data() *varData {
	return me._vdata
}

func (me *variable) copyData(data *varData) {
	me._vdata = data.copy()
}

func (me *hmfile) varInitFromData(data *varData, name string, mutable bool) *variable {
	v := &variable{}
	v.copyData(data)
	v.name = name
	v.cName = name
	v.mutable = mutable
	return v
}

func (me *hmfile) varInit(typed, name string, mutable bool) *variable {
	v := &variable{}
	v.name = name
	v.cName = name
	v.mutable = mutable
	v.update(me, typed)
	return v
}

func (me *variable) update(module *hmfile, typed string) {
	me.copyData(module.typeToVarData(typed))
}

func (me *variable) copy() *variable {
	v := &variable{}
	v.name = me.name
	v.cName = me.name
	v.mutable = me.mutable
	v.copyData(me.data())
	return v
}
