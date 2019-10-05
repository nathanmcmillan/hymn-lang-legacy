package main

import (
	"strings"
)

type class struct {
	name          string
	variables     map[string]*variable
	variableOrder []string
	generics      []string
	genericsDict  map[string]int
}

func classInit(name string, generics []string, genericsDict map[string]int) *class {
	c := &class{}
	c.name = name
	c.generics = generics
	c.genericsDict = genericsDict
	return c
}

func (me *class) initMembers(variableOrder []string, variables map[string]*variable) {
	me.variableOrder = variableOrder
	me.variables = variables
}

func (me *hmfile) getclass(name string) (*class, string) {
	ix := strings.Index(name, "[")
	if ix == -1 {
		cl, _ := me.classes[name]
		return cl, ""
	}
	get0 := name[0:ix]
	get1 := name[ix+1 : len(name)-1]
	cl, _ := me.classes[get0]
	return cl, get1
}

func (me *class) getGenerics() []string {
	return me.generics
}
