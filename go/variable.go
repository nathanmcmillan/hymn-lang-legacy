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

func (me *hmfile) varInit(typed, name string, mutable bool) (*variable, *parseError) {
	v := &variable{}
	v.name = name
	v.cname = name
	v.mutable = mutable
	var er *parseError
	v._vdata, er = getdatatype(me, typed)
	if er != nil {
		return nil, er
	}
	return v, nil
}

func (me *variable) copy() *variable {
	v := &variable{}
	v.name = me.name
	v.cname = me.name
	v.mutable = me.mutable
	v.copyData(me.data())
	return v
}
