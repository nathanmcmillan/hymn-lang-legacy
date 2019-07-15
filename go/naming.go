package main

import "strings"

var (
	globalClassPrefix = "Hm"
	globalEnumPrefix  = globalClassPrefix
	globalUnionPrefix = globalClassPrefix
	globalFuncPrefix  = "hm_"
	globalVarPrefix   = "me"
	definePrefix      = "HM_"
)

func upperSplit(name, repl string) string {
	full := ""
	parts := strings.Split(name, repl)
	for _, part := range parts {
		head := strings.ToUpper(part[0:1])
		body := part[1:]
		full += head + body
	}
	return full
}

func capital(name string) string {
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "")
	name = strings.ReplaceAll(name, ",", "And")
	name = upperSplit(name, "_")
	name = upperSplit(name, "And")
	return name
}

func (me *hmfile) defNameSpace(name string) string {
	name = strings.ToUpper(name)
	name = strings.ReplaceAll(name, "-", "_")
	return definePrefix + name + "_H"
}

func (me *hmfile) varNameSpace(id string) string {
	return globalVarPrefix + me.varPrefix + capital(id)
}

func (me *hmfile) funcNameSpace(name string) string {
	return globalFuncPrefix + me.funcPrefix + name
}

func (me *hmfile) classNameSpace(name string) string {
	return globalClassPrefix + me.classPrefix + capital(name)
}

func (me *hmfile) enumNameSpace(name string) string {
	return globalEnumPrefix + me.enumPrefix + capital(name)
}

func (me *hmfile) unionNameSpace(name string) string {
	return globalUnionPrefix + me.unionPrefix + "Union" + capital(name)
}

func (me *hmfile) enumTypeName(base, name string) string {
	return base + capital(name)
}

func (me *hmfile) prefixes(name string) {
	name = strings.ReplaceAll(name, "-", "_")

	me.funcPrefix = name + "_"
	me.classPrefix = capital(name)
	me.enumPrefix = me.classPrefix
	me.unionPrefix = me.classPrefix
	me.varPrefix = me.classPrefix
}
