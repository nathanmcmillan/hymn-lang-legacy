package main

import (
	"path/filepath"
)

type class struct {
	module          *hmfile
	name            string
	cname           string
	pathLocal       string
	pathGlobal      string
	variables       []*variable
	generics        []string
	mapping         map[string]*datatype
	functions       map[string]*function
	functionOrder   []*function
	base            *class
	implementations []*class
	doNotDefine     bool
	interfaces      map[string]*classInterface
}

func classInit(module *hmfile, name string, generics []string, interfaces map[string]*classInterface) *class {
	c := &class{}
	c.module = module
	c.name = name
	c.pathLocal = c.classFileName()
	if module != nil {
		c.cname = module.classNameSpace(name)
		c.pathGlobal = filepath.Join(module.relativeOut, c.pathLocal)
	}
	c.generics = generics
	c.functions = make(map[string]*function)
	c.functionOrder = make([]*function, 0)
	if len(generics) > 0 {
		c.implementations = make([]*class, 0)
		c.doNotDefine = true
	}
	c.interfaces = interfaces
	return c
}

func (me *class) initMembers(variables []*variable) {
	me.variables = variables
}

func (me *class) baseClass() *class {
	if me.base == nil {
		return me
	}
	return me.base.baseClass()
}

func (me *class) getGenerics() []string {
	return me.generics
}

func (me *class) uid() string {
	return me.module.reference(me.name)
}

func (me *class) getVariable(name string) *variable {
	return getVariable(me.variables, name)
}

func getVariable(variables []*variable, name string) *variable {
	for _, v := range variables {
		if name == v.name {
			return v
		}
	}
	return nil
}
