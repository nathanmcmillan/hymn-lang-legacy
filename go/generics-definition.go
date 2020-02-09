package main

import (
	"fmt"
)

func (me *parser) defineEnumImplGeneric(base *enum, order []*datatype) *enum {

	base = base.baseEnum()

	unionList := make([]*union, len(base.types))
	unionDict := make(map[string]*union)
	for i, v := range base.typesOrder {
		cp := v.copy()
		unionList[i] = cp
		unionDict[cp.name] = cp
	}

	module := base.module

	implementation := base.name + genericslist(order)
	uid := base.uid() + genericslist(order)

	module.namespace[uid] = "enum"
	module.types[uid] = "enum"

	module.namespace[implementation] = "enum"
	module.types[implementation] = "enum"

	enumDef := enumInit(base.module, implementation, false, unionList, unionDict, nil, nil)
	enumDef.base = base
	base.impls = append(base.impls, enumDef)

	module.defineOrder = append(module.defineOrder, &defineType{enum: enumDef})

	module.enums[uid] = enumDef
	module.enums[implementation] = enumDef

	gmapper := make(map[string]string)
	for ix, gname := range base.generics {
		value := order[ix].getRaw()
		gmapper[gname] = value
		if gname == value {
			enumDef.doNotDefine = true
		}
	}
	enumDef.gmapper = gmapper

	for _, un := range unionList {
		for i, data := range un.types {
			// TODO:
			// un.types[i] = getdatatype(module, me.genericsReplacer(data, gmapper).print())
			un.types[i] = me.genericsReplacer(module, data, gmapper)
		}
	}

	return enumDef
}

func (me *parser) defineClassImplGeneric(base *class, order []*datatype) *class {

	base = base.baseClass()
	module := base.module

	implementation := base.name + genericslist(order)
	uid := base.uid() + genericslist(order)

	for k, v := range base.variables {
		fmt.Println(uid+" base class variables ::", k, "::", v.data().print())
	}

	fmt.Println("inserting new class ::", module.name, "::", implementation, "(", uid, ")", "::", base.name, "::", genericslist(order))

	module.namespace[uid] = "type"
	module.types[uid] = "class"

	module.namespace[implementation] = "type"
	module.types[implementation] = "class"

	classDef := classInit(module, implementation, nil, nil)
	classDef.base = base

	module.defineOrder = append(module.defineOrder, &defineType{class: classDef})

	module.classes[uid] = classDef
	module.classes[implementation] = classDef

	for k := range module.classes {
		fmt.Println("updated module class list ::", k)
	}

	base.implementations = append(base.implementations, classDef)

	gmapper := make(map[string]string)
	for ix, gname := range base.generics {
		from := order[ix]
		value := from.getRaw()
		fmt.Println(implementation, "generic map ::", gname, "<-", value)
		gmapper[gname] = value
		if gname == value {
			classDef.doNotDefine = true
		} else if from.isRecursiveUnknown() {
			fmt.Println("DO NOT DEFINE THIS CLASS BECAUSE OF ::", from.print())
			classDef.doNotDefine = true
		}
	}
	classDef.gmapper = gmapper

	if base.variables != nil {
		me.finishClassDefinition(classDef)
	}

	return classDef
}

func (me *parser) finishClassDefinition(classDef *class) {

	memberMap := make(map[string]*variable)
	for k, v := range classDef.base.variables {
		memberMap[k] = v.copy()
	}

	classDef.initMembers(classDef.base.variableOrder, memberMap)

	for k, v := range memberMap {
		fmt.Println(classDef.name, "member map ::", k, "::", v.data().print())
	}

	for _, mem := range memberMap {
		// TODO:
		// replace := me.genericsReplacer(mem.data(), classDef.gmapper).print()
		// data := getdatatype(classDef.module, replace)
		data := me.genericsReplacer(classDef.module, mem.data(), classDef.gmapper)
		mem._vdata = data
	}

	for _, fn := range classDef.base.functionOrder {
		remapClassFunctionImpl(classDef, fn)
	}
}
