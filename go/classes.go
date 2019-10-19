package main

type class struct {
	name          string
	variables     map[string]*variable
	variableOrder []string
	generics      []string
	genericsDict  map[string]int
	gmapper       map[string]string
	functions     map[string]*function
	base          *class
	impls         []*class
}

func classInit(name string, generics []string, genericsDict map[string]int) *class {
	c := &class{}
	c.name = name
	c.generics = generics
	c.genericsDict = genericsDict
	c.functions = make(map[string]*function)
	if len(generics) > 0 {
		c.impls = make([]*class, 0)
	}
	return c
}

func (me *class) initMembers(variableOrder []string, variables map[string]*variable) {
	me.variableOrder = variableOrder
	me.variables = variables
}

func (me *class) getGenerics() []string {
	return me.generics
}
