package main

import "strings"

var (
	globalClassPrefix = "Hm"
	globalEnumPrefix  = globalClassPrefix
	globalUnionPrefix = globalClassPrefix
	globalFuncPrefix  = "hm_"
	globalVarPrefix   = "hm"
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

func flatten(name string) string {
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "")
	name = strings.ReplaceAll(name, ",", "_and_")
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

func (me *hmfile) enumMaybeImplNameSpace(name string) (bool, string) {
	impl := false
	i := strings.Index(name, "<")
	if i != -1 {
		impl = true
		name = name[:i]
	}
	return impl, me.enumNameSpace(name)
}

func (me *hmfile) enumNameSpace(name string) string {
	return globalEnumPrefix + me.enumPrefix + capital(name)
}

func (me *hmfile) unionNameSpace(name string) string {
	return globalUnionPrefix + me.unionPrefix + capital(name)
}

func (me *hmfile) unionFnNameSpace(en *enum, un *union) string {
	return globalFuncPrefix + me.funcPrefix + "new_" + flatten(en.name) + "_" + un.name
}

func (me *hmfile) enumTypeName(base, name string) string {
	return base + capital(name)
}

func (me *hmfile) prefixes(name string) {
	name = strings.ReplaceAll(name, "-", "_")

	me.funcPrefix = name + "_"
	me.classPrefix = capital(name)
	me.enumPrefix = me.classPrefix
	me.unionPrefix = me.classPrefix + "Union"
	me.varPrefix = me.classPrefix
}
