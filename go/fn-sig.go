package main

import "fmt"

type fnSig struct {
	args  []*funcArg
	typed *varData
}

func fnSigInit() *fnSig {
	f := &fnSig{}
	f.args = make([]*funcArg, 0)
	return f
}

func (me *fnSig) asVar() *varData {
	fmt.Println(":: SIG TO VAR")
	return nil
}
