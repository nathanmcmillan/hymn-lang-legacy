package main

type variable struct {
	name    string
	cname   string
	mutable bool
	used    bool
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
	v := me.declareVar(name, mutable)
	var er *parseError
	v._vdata, er = getdatatype(me, typed)
	if er != nil {
		return nil, er
	}
	return v, nil
}

func (me *hmfile) varInitWithData(data *datatype, name string, mutable bool) *variable {
	v := me.declareVar(name, mutable)
	v._vdata = data
	return v
}

func (me *hmfile) declareVar(name string, mutable bool) *variable {
	v := &variable{}
	v.name = name
	v.cname = name
	v.mutable = mutable
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
