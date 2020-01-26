package main

type idData struct {
	module *hmfile
	name   string
	cname  string
	global bool
}

func (me *idData) copy() *idData {
	i := &idData{}
	i.module = me.module
	i.name = me.name
	i.cname = me.cname
	return i
}

func (me *idData) getcname() string {
	return me.cname
}

func newidvariable(module *hmfile, name string) *idData {
	i := &idData{}
	i.module = module
	i.name = name
	i.cname = module.varNameSpace(name)
	return i
}

func (me *idData) string() string {
	return me.module.name + "." + me.name
}

func (me *idData) isGlobal() bool {
	return me.global
}

func (me *idData) setGlobal(flag bool) {
	me.global = flag
}
