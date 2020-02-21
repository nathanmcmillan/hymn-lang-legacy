package main

func (me *parser) verifyFile() {
	module := me.hmfile
	// TODO: do something about duplicates
	for _, cl := range module.classes {
		for iname, in := range cl.interfaces {
			for fname, infn := range in.functions {
				if clfn, ok := cl.functions[fname]; ok {
					sig := clfn.asSig()
					if sig.equals(infn) {
						panic(me.fail() + "Class '" + cl.name + "' function signature '" + fname + "' " + sig.print() + " does not match interface '" + iname + "' " + infn.print())
					} else {
						panic(me.fail() + "Class '" + cl.name + "' Missing " + fname + " for interface '" + iname + "'")
					}
				}
			}
		}
	}
}
