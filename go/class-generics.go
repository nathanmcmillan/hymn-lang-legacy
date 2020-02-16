package main

func (me *parser) defineClassImplGeneric(base *class, order []*datatype) *class {

	base = base.baseClass()
	module := base.module

	implementation := base.name + genericslist(order)
	uid := base.uid() + genericslist(order)

	module.namespace[uid] = "class"
	module.types[uid] = "class"

	module.namespace[implementation] = "class"
	module.types[implementation] = "class"

	classDef := classInit(module, implementation, nil, nil)
	classDef.base = base

	module.defineOrder = append(module.defineOrder, &defineType{class: classDef})

	module.classes[uid] = classDef
	module.classes[implementation] = classDef

	base.implementations = append(base.implementations, classDef)

	gmapper := make(map[string]string)
	for ix, gname := range base.generics {
		from := order[ix]
		value := from.getRaw()
		gmapper[gname] = value
		if gname == value || from.isRecursiveUnknown() {
			classDef.doNotDefine = true
		}
	}
	classDef.gmapper = gmapper

	if base.variables != nil && len(base.variables) > 0 {
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
