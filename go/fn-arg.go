package main

type funcArg struct {
	*variable
	defaultNode *node
	used        bool
}

func (me *funcArg) copy() *funcArg {
	a := &funcArg{}
	a.variable = me.variable.copy()
	if me.defaultNode != nil {
		a.defaultNode = me.defaultNode.copy()
	}
	return a
}

func (me *hmfile) fnArgInit(typed, name string, mutable bool) (*funcArg, *parseError) {
	f := &funcArg{}
	var er *parseError
	f.variable, er = me.varInit(typed, name, mutable)
	if er != nil {
		return nil, er
	}
	return f, nil
}

func fnArgInit(v *variable) *funcArg {
	f := &funcArg{}
	f.variable = v
	return f
}

func (me *hmlib) fnArgInit(typed, name string, mutable bool) (*funcArg, *parseError) {
	f := &funcArg{}
	v := &variable{}
	v.name = name
	v.cname = name
	v.mutable = mutable
	data, er := getdatatype(nil, typed)
	if er != nil {
		return nil, er
	}
	v.copyData(data)
	f.variable = v
	return f, nil
}
