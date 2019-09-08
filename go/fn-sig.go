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
		sig += arg.vdat.full
	}
	sig += ")"
	if me.typed.full != "void" {
		sig += " "
		sig += me.typed.full
	}
	return sig
}

func (me *fnSig) asVar() *varData {
	sig := me.print()
	d := &varData{}
	d.fn = me
	d.full = sig
	d.typed = sig
	d.module = me.module
	return d
}
