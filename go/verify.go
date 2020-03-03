package main

func (me *parser) verifyFile() *parseError {
	module := me.hmfile
	for _, def := range module.defineOrder {
		cl := def.class
		if cl == nil {
			continue
		}
		for iname, in := range cl.selfInterfaces {
			for fname, infn := range in.functions {
				if clfn := cl.getFunction(fname); clfn != nil {
					sig := clfn.asSig()
					if !sig.equals(infn) {
						e := "Class '" + cl.name + "' with function '" + fname + sig.print() + "'"
						e += " does not match interface '" + iname + "' with '" + fname + infn.print() + "'"
						return err(me, ECodeClassInterfaceSignatureMismatch, e)
					}
				} else {
					e := "Class '" + cl.name + "' is missing function '" + fname + infn.print() + "' for interface '" + iname + "'"
					return err(me, ECodeClassMissingRequiredInterface, e)
				}
			}
		}
	}
	return nil
}
