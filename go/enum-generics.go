package main

func (me *parser) defineEnumImplGeneric(base *enum, order []*datatype) *enum {

	base = base.baseEnum()
	module := base.module

	implementation := base.name + genericslist(order)
	uid := base.uid() + genericslist(order)

	module.namespace[uid] = "enum"
	module.types[uid] = "enum"

	module.namespace[implementation] = "enum"
	module.types[implementation] = "enum"

	enumDef := enumInit(base.module, implementation)
	enumDef.base = base

	module.defineOrder = append(module.defineOrder, &defineType{enum: enumDef})

	module.enums[uid] = enumDef
	module.enums[implementation] = enumDef

	base.implementations = append(base.implementations, enumDef)

	gmapper := make(map[string]string)
	for ix, gname := range base.generics {
		from := order[ix]
		value := from.getRaw()
		gmapper[gname] = value
		if gname == value || from.isRecursiveUnknown() {
			enumDef.doNotDefine = true
		}
	}
	enumDef.gmapper = gmapper

	if base.types != nil && len(base.types) > 0 {
		me.finishEnumGenericDefinition(enumDef)
	}

	return enumDef
}

func (me *parser) finishEnumGenericDefinition(enumDef *enum) {

	unionList := make([]*union, len(enumDef.base.types))
	unionDict := make(map[string]*union)
	for i, v := range enumDef.base.typesOrder {
		cp := v.copy()
		unionList[i] = cp
		unionDict[cp.name] = cp
	}

	for _, un := range unionList {
		for _, dataKey := range un.types.order {
			data := un.types.table[dataKey]
			un.types.table[dataKey] = me.genericsReplacer(enumDef.module, data, enumDef.gmapper)
		}
	}

	enumDef.finishInit(false, unionList, unionDict, nil, nil)
}
