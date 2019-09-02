package main

type funcArg struct {
	*variable
	defaultNode *node
}

func (me *hmfile) fnArgInit(typed, name string, mutable, isptr bool) *funcArg {
	fa := &funcArg{}
	fa.variable = me.varInit(typed, name, mutable, isptr)
	return fa
}

func fnArgInit(v *variable) *funcArg {
	f := &funcArg{}
	f.variable = v
	return f
}
