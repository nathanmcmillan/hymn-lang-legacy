package main

func (me *parser) defineInterfaceImplementation(in *classInterface, generics []*datatype) (*classInterface, *parseError) {

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

	var er *parseError

	for fname, superfn := range super.functions {
		functions[fname], er = superfn.genericsReplacer(me, mapping)
		if er != nil {
			return nil, er
		}
	}

	interfaceDef := interfaceInit(module, implementation, generics, functions)
	interfaceDef.super = super

	module.interfaces[uid] = interfaceDef
	module.interfaces[implementation] = interfaceDef

	me.program.interfaces[uid] = interfaceDef

	super.implementations = append(super.implementations, interfaceDef)

	return interfaceDef, nil
}
