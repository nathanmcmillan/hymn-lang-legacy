package main

type funcArg struct {
	*variable
	defaultNode *node
}

func (me *hmfile) fnArgInit(typed, name string, mutable, isptr bool) *funcArg {
	f := &funcArg{}
	f.variable = me.varInit(typed, name, mutable, isptr)
	return f
}

func fnArgInit(v *variable) *funcArg {
	f := &funcArg{}
	f.variable = v
	return f
}

func (me *hmlib) fnArgInit(typed, name string, mutable, isptr bool) *funcArg {
	f := &funcArg{}
	v := &variable{}
	v.name = name
	v.cName = name
	v.mutable = mutable
	v.isptr = isptr
	v.vdat = me.literalType(typed)
	v.vdat.isptr = isptr
	f.variable = v
	return f
}
