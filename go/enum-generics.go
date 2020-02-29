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

	mapping := make(map[string]*datatype)
	for ix, gname := range base.generics {
		from := order[ix]
		value := from.getRaw()
		mapping[gname] = from
		if gname == value || from.isRecursiveUnknown() {
			enumDef.doNotDefine = true
		}
	}
	enumDef.mapping = mapping

	if len(base.interfaces) > 0 {
		for _, g := range base.generics {
			i, ok := base.interfaces[g]
			if !ok {
				continue
			}
			m := mapping[g]
			if cl, ok := m.isClass(); ok {
				for _, t := range i {
					if _, ok := cl.interfaces[t.name]; !ok {
						panic(me.fail() + "Class '" + cl.name + "' for enum '" + implementation + "' requires interface '" + t.name + "'")
					}
				}
			} else {
				panic(me.fail() + "Enum '" + implementation + "' requires interface implementation but type was " + m.error())
			}
		}
	}

	if base.types != nil && len(base.types) > 0 {
		me.finishEnumGenericDefinition(enumDef)
	}

	return enumDef
}

func (me *parser) finishEnumGenericDefinition(enumDef *enum) {

	unionList := make([]*union, len(enumDef.base.types))
	for i, v := range enumDef.base.types {
		cp := v.copy()
		unionList[i] = cp
	}

	mapping := make(map[string]string)
	for k, m := range enumDef.mapping {
		mapping[k] = m.getRaw()
	}

	for _, un := range unionList {
		for _, dataKey := range un.types.order {
			data := un.types.table[dataKey]
			un.types.table[dataKey] = me.genericsReplacer(enumDef.module, data, mapping)
		}
	}

	enumDef.finishInit(false, unionList, nil, nil)
}
