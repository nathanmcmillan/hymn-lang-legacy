package main

import "path/filepath"

type enum struct {
	module       *hmfile
	name         string
	cname        string
	ucname       string
	pathLocal    string
	pathGlobal   string
	simple       bool
	types        map[string]*union
	typesOrder   []*union
	generics     []string
	genericsDict map[string]int
	gmapper      map[string]string
	base         *enum
	impls        []*enum
	doNotDefine  bool
}

type union struct {
	name     string
	types    []*datatype
	generics []string
}

func unionInit(module *hmfile, en string, name string, types []string, generics []string) *union {
	u := &union{}
	u.name = name
	u.types = make([]*datatype, len(types))
	for i, t := range types {
		u.types[i] = getdatatype(module, t)
	}
	u.generics = generics
	return u
}

func (me *union) copy() *union {
	u := &union{}
	u.name = me.name
	u.types = make([]*datatype, len(me.types))
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
	e.pathLocal = e.enumFileName()
	if module != nil {
		e.cname = module.enumNameSpace(name)
		e.ucname = module.unionNameSpace(name)
		e.pathGlobal = filepath.Join(module.relativeOut, e.pathLocal)
	}
	e.simple = simple
	e.types = dict
	e.typesOrder = order
	e.generics = generics
	e.genericsDict = genericsDict
	if len(generics) > 0 {
		e.impls = make([]*enum, 0)
		e.doNotDefine = true
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
		return me.cname
	}
	return me.ucname + " *"
}

func (me *enum) noMallocTypeSig() string {
	return me.cname
}

func (me *enum) getGenerics() []string {
	return me.generics
}

func (me *enum) uid() string {
	return me.module.reference(me.name)
}
