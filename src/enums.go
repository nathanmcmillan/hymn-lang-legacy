package main

import "path/filepath"

type enum struct {
	module          *hmfile
	name            string
	cname           string
	ucname          string
	pathLocal       string
	pathGlobal      string
	simple          bool
	types           []*union
	generics        []string
	mapping         map[string]*datatype
	interfaces      map[string][]*classInterface
	base            *enum
	implementations []*enum
	doNotDefine     bool
}

type union struct {
	name  string
	types *ordereddata
}

func unionInit(module *hmfile, en string, name string, types *ordereddata) *union {
	u := &union{}
	u.name = name
	u.types = types
	return u
}

func (me *union) copy() *union {
	u := &union{}
	u.name = me.name
	u.types = newordereddata()
	for _, t := range me.types.order {
		u.types.push(t, me.types.table[t])
	}
	return u
}

func enumInit(module *hmfile, name string) *enum {
	e := &enum{}
	e.module = module
	e.name = name
	e.pathLocal = e.enumFileName()
	if module != nil {
		e.cname = module.enumNameSpace(name)
		e.ucname = module.unionNameSpace(name)
		e.pathGlobal = filepath.Join(module.includes, e.pathLocal)
	}
	return e
}

func (me *enum) finishInit(simple bool, types []*union, generics []string, interfaces map[string][]*classInterface) {
	me.simple = simple
	me.types = types
	me.generics = generics
	me.interfaces = interfaces
	if len(generics) > 0 {
		me.implementations = make([]*enum, 0)
		me.doNotDefine = true
	}
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

func (me *enum) join(un *union) string {
	return me.name + "." + un.name
}

func (me *enum) getType(name string) *union {
	return getUnionType(me.types, name)
}

func getUnionType(unions []*union, name string) *union {
	for _, v := range unions {
		if name == v.name {
			return v
		}
	}
	return nil
}

func (me *enum) getuniondata(module *hmfile, union string) (*datatype, *parseError) {
	d, er := getdatatype(module, me.module.reference(me.name)+"."+union)
	if er != nil {
		return nil, er
	}
	return d, nil
}
