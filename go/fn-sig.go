package main

type fnSig struct {
	module *hmfile
	args   []*funcArg
	typed  *varData
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
	sig += ")"
	if me.typed.full != "void" {
		sig += " "
		sig += me.typed.full
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
