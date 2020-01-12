package main

type idData struct {
	module *hmfile
	name   string
}

func (me *idData) copy() *idData {
	i := &idData{}
	i.module = me.module
	i.name = me.name
	return i
}

func (me *idData) string() string {
	return me.module.name + "." + me.name
}
