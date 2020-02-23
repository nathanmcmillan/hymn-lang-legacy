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

	gmapper := make(map[string]string)
	for ix, gname := range super.generics {
		from := order[ix]
		value := from.getRaw()
		gmapper[gname] = value
		if gname == value || from.isRecursiveUnknown() {
			classDef.doNotDefine = true
		}
	}
	classDef.gmapper = gmapper

	classDef.interfaces = make(map[string]*classInterface)
	for key, in := range super.interfaces {
		if !in.requiresGenerics() {
			classDef.interfaces[key] = in
			continue
		}
		super := in.getSuper()
		generics := make([]*datatype, len(in.generics))
		for i := 0; i < len(generics); i++ {
			if gn, ok := gmapper[in.generics[i].getRaw()]; ok {
				generics[i] = getdatatype(classDef.module, gn)
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
		classDef.interfaces[key] = in
	}

	if super.variables != nil && len(super.variables) > 0 {
		me.finishClassGenericDefinition(classDef)
	}

	return classDef
}

func (me *parser) finishClassGenericDefinition(classDef *class) {

	memberMap := make(map[string]*variable)
	for k, v := range classDef.base.variables {
		memberMap[k] = v.copy()
	}

	classDef.initMembers(classDef.base.variableOrder, memberMap)

	for _, mem := range memberMap {
		data := me.genericsReplacer(classDef.module, mem.data(), classDef.gmapper)
		mem._vdata = data
	}

	for _, fn := range classDef.base.functionOrder {
		remapClassFunctionImpl(classDef, fn)
	}
}
