package main

type fnSig struct {
	module      *hmfile
	args        []*funcArg
	argVariadic *funcArg
	returns     *varData
}

func fnSigInit(module *hmfile) *fnSig {
	f := &fnSig{}
	f.module = module
	f.args = make([]*funcArg, 0)
	return f
}

func (me *fnSig) print() string {
	sig := "("
	for ix, arg := range me.args {
		if ix > 0 {
			sig += ", "
		}
		sig += arg.data().getRaw()
	}
	if me.argVariadic != nil {
		if len(me.args) > 0 {
			sig += ", "
		}
		sig += "..." + me.argVariadic.data().getRaw()
	}
	sig += ")"
	if !me.returns.isVoid() {
		sig += " "
		sig += me.returns.getRaw()
	}
	return sig
}
