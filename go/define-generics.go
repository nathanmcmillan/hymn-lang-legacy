package main

import (
	"fmt"
)

func (me *parser) defineEnumImplGeneric(base *enum, order []*datatype) *enum {

	unionList := make([]*union, len(base.types))
	unionDict := make(map[string]*union)
	for i, v := range base.typesOrder {
		cp := v.copy()
		unionList[i] = cp
		unionDict[cp.name] = cp
	}

	implementation := base.name + genericslist(order)
	uid := base.uid() + genericslist(order)

	me.hmfile.namespace[uid] = "enum"
	me.hmfile.types[uid] = "enum"

	me.hmfile.namespace[implementation] = "enum"
	me.hmfile.types[implementation] = "enum"

	me.hmfile.defineOrder = append(me.hmfile.defineOrder, implementation+"_enum")

	enumDef := enumInit(base.module, implementation, false, unionList, unionDict, nil, nil)
	enumDef.base = base
	base.impls = append(base.impls, enumDef)

	me.hmfile.enums[uid] = enumDef
	me.hmfile.enums[implementation] = enumDef

	gmapper := make(map[string]string)
	for ix, gname := range base.generics {
		gmapper[gname] = order[ix].getRaw()
	}
	enumDef.gmapper = gmapper

	for _, un := range unionList {
		for i, data := range un.types {
			un.types[i] = getdatatype(me.hmfile, me.genericsReplacer(data, gmapper).print())
		}
	}

	return enumDef
}

func (me *parser) defineClassImplGeneric(base *class, order []*datatype) *class {
	memberMap := make(map[string]*variable)
	for k, v := range base.variables {
		memberMap[k] = v.copy()
	}

	module := base.module

	implementation := base.name + genericslist(order)
	uid := base.uid() + genericslist(order)

	fmt.Println("inserting new class ::", module.name, "::", implementation, "(", uid, ")", "::", base.name, "::", genericslist(order))

	module.namespace[uid] = "type"
	module.types[uid] = "class"

	module.namespace[implementation] = "type"
	module.types[implementation] = "class"

	module.defineOrder = append(module.defineOrder, implementation+"_type")

	classDef := classInit(module, implementation, nil, nil)
	classDef.base = base
	base.impls = append(base.impls, classDef)
	classDef.initMembers(base.variableOrder, memberMap)

	module.classes[uid] = classDef
	module.classes[implementation] = classDef

	for k := range module.classes {
		fmt.Println("update module class list ::", k)
	}

	gmapper := make(map[string]string)
	for ix, gname := range base.generics {
		fmt.Println(implementation, "generic map ::", gname, "<-", order[ix].getRaw())
		gmapper[gname] = order[ix].getRaw()
	}
	classDef.gmapper = gmapper

	for _, mem := range memberMap {
		replace := me.genericsReplacer(mem.data(), gmapper).print()
		data := getdatatype(module, replace)
		clname := ""
		cl, ok := data.isClass()
		if ok {
			clname = " | " + cl.name
		}
		fmt.Println(implementation, "replacing member ::", mem.name, "<-", data.print()+clname)
		mem._vdata = data
	}

	for _, fn := range base.functionOrder {
		remapClassFunctionImpl(classDef, fn)
	}

	return classDef
}
