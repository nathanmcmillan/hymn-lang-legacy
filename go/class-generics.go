package main

func (me *parser) defineClassImplGeneric(super *class, order []*datatype) *class {

	super = super.baseClass()
	module := super.module

	implementation := super.name + genericslist(order)
	uid := super.uid() + genericslist(order)

	module.namespace[uid] = "class"
	module.types[uid] = "class"

	module.namespace[implementation] = "class"
	module.types[implementation] = "class"

	classDef := classInit(module, implementation, nil, nil, nil)
	classDef.base = super

	module.defineOrder = append(module.defineOrder, &defineType{class: classDef})

	module.classes[uid] = classDef
	module.classes[implementation] = classDef

	super.implementations = append(super.implementations, classDef)

	mapping := make(map[string]*datatype)
	for ix, gname := range super.generics {
		from := order[ix]
		value := from.getRaw()
		mapping[gname] = from
		if gname == value || from.isRecursiveUnknown() {
			classDef.doNotDefine = true
		}
	}
	classDef.mapping = mapping

	if len(super.genericsInterfaces) > 0 {
		for _, g := range super.generics {
			i, ok := super.genericsInterfaces[g]
			if !ok {
				continue
			}
			m := mapping[g]
			if cl, ok := m.isClass(); ok {
				for _, t := range i {
					if _, ok := cl.selfInterfaces[t.uid()]; !ok {
						panic(me.fail() + "Class '" + cl.name + "' for '" + implementation + "' requires interface '" + t.name + "'")
					}
				}
			} else {
				panic(me.fail() + "Class '" + implementation + "' requires interface implementation but type was " + m.error())
			}
		}
	}

	classDef.selfInterfaces = make(map[string]*classInterface)
	for key, in := range super.selfInterfaces {
		if !in.requiresGenerics() {
			classDef.selfInterfaces[key] = in
			continue
		}
		super := in.getSuper()
		generics := make([]*datatype, len(in.generics))
		for i := 0; i < len(generics); i++ {
			if gn, ok := mapping[in.generics[i].getRaw()]; ok {
				generics[i] = gn
			} else {
				generics[i] = in.generics[i]
			}
		}
		intname := super.name + genericslist(generics)
		if gotInterface, ok := module.interfaces[intname]; ok {
			in = gotInterface
		} else {
			in = me.defineInterfaceImplementation(in, generics)
		}
		classDef.selfInterfaces[key] = in
	}

	if super.variables != nil && len(super.variables) > 0 {
		me.finishClassGenericDefinition(classDef)
	}

	return classDef
}

func (me *parser) finishClassGenericDefinition(classDef *class) {

	members := make([]*variable, len(classDef.base.variables))
	for i, v := range classDef.base.variables {
		members[i] = v.copy()
	}

	classDef.initMembers(members)

	mapping := make(map[string]string)
	for k, m := range classDef.mapping {
		mapping[k] = m.getRaw()
	}

	for _, mem := range members {
		data := me.genericsReplacer(classDef.module, mem.data(), mapping)
		mem._vdata = data
	}

	for _, fn := range classDef.base.functions {
		remapClassFunctionImpl(classDef, fn)
	}
}
