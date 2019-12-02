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
		sig += arg.data().full
	}
	if me.argVariadic != nil {
		if len(me.args) > 0 {
			sig += ", "
		}
		sig += "..." + me.argVariadic.data().full
	}
	sig += ")"
	if me.returns.full != "void" {
		sig += " "
		sig += me.returns.full
	}
	return sig
}

func (me *fnSig) data() *varData {
	sig := me.print()
	d := &varData{}
	d.fn = me
	d.full = sig
	d.typed = sig
	d.module = me.module
	d.dtype = getdatatype(nil, sig)
	return d
}
