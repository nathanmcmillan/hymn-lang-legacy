package main

type funcArg struct {
	*variable
	defaultNode *node
}

func (me *funcArg) copy() *funcArg {
	a := &funcArg{}
	a.variable = me.variable.copy()
	a.defaultNode = me.defaultNode.copy()
	return a
}

func (me *hmfile) fnArgInit(typed, name string, mutable bool) *funcArg {
	f := &funcArg{}
	f.variable = me.varInit(typed, name, mutable)
	return f
}

func fnArgInit(v *variable) *funcArg {
	f := &funcArg{}
	f.variable = v
	return f
}

func (me *hmlib) fnArgInit(typed, name string, mutable bool) *funcArg {
	f := &funcArg{}
	v := &variable{}
	v.name = name
	v.cName = name
	v.mutable = mutable
	v.copyData(me.literalType(typed))
	f.variable = v
	return f
}
