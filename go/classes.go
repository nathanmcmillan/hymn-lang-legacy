package main

type class struct {
	module        *hmfile
	name          string
	cname         string
	variables     map[string]*variable
	variableOrder []string
	generics      []string
	genericsDict  map[string]int
	gmapper       map[string]string
	functions     map[string]*function
	functionOrder []*function
	base          *class
	impls         []*class
}

func classInit(module *hmfile, name string, generics []string, genericsDict map[string]int) *class {
	c := &class{}
	c.module = module
	c.name = name
	c.cname = getdatatype(module, name).cname()
	c.generics = generics
	c.genericsDict = genericsDict
	c.functions = make(map[string]*function)
	c.functionOrder = make([]*function, 0)
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
