package main

type classInterface struct {
	module          *hmfile
	name            string
	generics        []*datatype
	functions       map[string]*fnSig
	super           *classInterface
	implementations []*classInterface
}

func interfaceInit(module *hmfile, name string, generics []*datatype, functions map[string]*fnSig) *classInterface {
	i := &classInterface{}
	i.module = module
	i.name = name
	i.functions = functions
	if len(generics) > 0 {
		i.generics = generics
		i.implementations = make([]*classInterface, 0)
	}
	return i
}

func (me *classInterface) requiresGenerics() bool {
	return me.generics != nil
}

func (me *classInterface) getSuper() *classInterface {
	if me.super == nil {
		return me
	}
	return me.super.getSuper()
}
