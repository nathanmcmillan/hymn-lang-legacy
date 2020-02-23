package main

func (me *parser) defineInterfaceImplementation(in *classInterface, generics []*datatype) *classInterface {

	super := in.getSuper()
	module := super.module

	implementation := super.name + genericslist(generics)
	uid := module.reference(implementation)

	module.namespace[uid] = "interface"
	module.types[uid] = "interface"

	module.namespace[implementation] = "interface"
	module.types[implementation] = "interface"

	mapping := make(map[string]string)
	for ix, gname := range super.generics {
		from := generics[ix]
		value := from.getRaw()
		mapping[gname.getRaw()] = value
	}

	functions := make(map[string]*fnSig)

	for fname, superfn := range super.functions {
		functions[fname] = superfn.genericsReplacer(me, mapping)
	}

	interfaceDef := interfaceInit(module, implementation, nil, functions)
	interfaceDef.super = super

	module.interfaces[uid] = interfaceDef
	module.interfaces[implementation] = interfaceDef

	super.implementations = append(super.implementations, interfaceDef)

	return interfaceDef
}
