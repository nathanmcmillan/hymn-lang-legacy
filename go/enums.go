package main

type enum struct {
	module       *hmfile
	name         string
	simple       bool
	types        map[string]*union
	typesOrder   []*union
	generics     []string
	genericsDict map[string]int
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
		u.types[i] = me.typeToVarData(t)
	}
	u.generics = generics
	return u
}

func (me *union) copy() *union {
	u := &union{}
	u.name = me.name
	u.types = make([]*varData, len(me.types))
	u.generics = make([]string, len(me.generics))
	copy(u.types, me.types)
	copy(u.generics, me.generics)
	return u
}

func enumInit(module *hmfile, name string, simple bool, order []*union, dict map[string]*union, generics []string, genericsDict map[string]int) *enum {
	e := &enum{}
	e.module = module
	e.name = name
	e.simple = simple
	e.types = dict
	e.typesOrder = order
	e.generics = generics
	e.genericsDict = genericsDict
	return e
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
