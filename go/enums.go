package main

import "strings"

type enum struct {
	module       *hmfile
	name         string
	location     string
	simple       bool
	types        map[string]*union
	typesOrder   []*union
	generics     []string
	genericsDict map[string]int
	base         *enum
	impls        []*enum
}

type union struct {
	name     string
	types    []*varData
	generics []string
}

func (me *hmfile) unionInit(name string, types []string, generics []string) *union {
	u := &union{}
	u.name = name
	u.types = make([]*varData, len(types))
	for i, t := range types {
		u.types[i] = typeToVarData(me, t)
	}
	u.generics = generics
	return u
}

func (me *union) copy() *union {
	u := &union{}
	u.name = me.name
	u.types = make([]*varData, len(me.types))
	for i, t := range me.types {
		u.types[i] = t.copy()
	}
	u.generics = make([]string, len(me.generics))
	for i, g := range me.generics {
		u.generics[i] = g
	}
	return u
}

func enumInit(module *hmfile, name string, simple bool, order []*union, dict map[string]*union, generics []string, genericsDict map[string]int) *enum {
	e := &enum{}
	e.module = module
	e.name = name
	e.location = e.getLocation()
	e.simple = simple
	e.types = dict
	e.typesOrder = order
	e.generics = generics
	e.genericsDict = genericsDict
	if len(generics) > 0 {
		e.impls = make([]*enum, 0)
	}
	return e
}

func (me *enum) baseEnum() *enum {
	if me.base == nil {
		return me
	}
	return me.base.baseEnum()
}

func (me *enum) typeSig() string {
	if me.simple {
		return me.module.enumNameSpace(me.name)
	}
	return me.module.unionNameSpace(me.name) + " *"
}

func (me *enum) noMallocTypeSig() string {
	if me.simple {
		return me.module.enumNameSpace(me.name)
	}
	return me.module.unionNameSpace(me.name)
}

func (me *enum) getGenerics() []string {
	return me.generics
}

func (me *enum) getLocation() string {
	// path := ""
	name := me.name
	// if strings.Index(name, "<") != -1 {
	// 	path = name[0:strings.Index(name, "<")]
	// } else {
	// 	path = name
	// }
	name = flatten(name)
	name = strings.ReplaceAll(name, "_", "-")
	return name // path + "/" + name
}
