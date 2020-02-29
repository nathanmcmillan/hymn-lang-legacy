package main

func (me *parser) verifyFile() {
	module := me.hmfile
	for _, def := range module.defineOrder {
		cl := def.class
		if cl == nil {
			continue
		}
		for iname, in := range cl.interfaces {
			for fname, infn := range in.functions {
				if clfn := cl.getFunction(fname); clfn != nil {
					sig := clfn.asSig()
					if !sig.equals(infn) {
						e := me.fail()
						e += "Class '" + cl.name + "' with function '" + fname + sig.print() + "'"
						e += " does not match interface '" + iname + "' with '" + fname + infn.print() + "'"
						panic(e)
					}
				} else {
					panic(me.fail() + "Class '" + cl.name + "' is missing function '" + fname + infn.print() + "' for interface '" + iname + "'")
				}
			}
		}
	}
}
