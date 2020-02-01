package main

type variable struct {
	name    string
	cname   string
	mutable bool
	_vdata  *datatype
}

type variableNode struct {
	n *node
	v *variable
}

func (me *variable) data() *datatype {
	return me._vdata
}

func (me *variable) copyData(data *datatype) {
	me._vdata = data.copy()
}

func (me *hmfile) varInit(typed, name string, mutable bool) *variable {
	v := &variable{}
	v.name = name
	v.cname = name
	v.mutable = mutable
	v._vdata = getdatatype(me, typed)
	return v
}

func (me *variable) copy() *variable {
	v := &variable{}
	v.name = me.name
	v.cname = me.name
	v.mutable = me.mutable
	v.copyData(me.data())
	return v
}
