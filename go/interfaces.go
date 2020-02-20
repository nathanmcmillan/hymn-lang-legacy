package main

type classInterface struct {
	module    *hmfile
	name      string
	functions map[string]*fnSig
}

func interfaceInit(module *hmfile, name string, functions map[string]*fnSig) *classInterface {
	i := &classInterface{}
	i.module = module
	i.name = name
	i.functions = functions
	return i
}
