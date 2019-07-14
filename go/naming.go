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

func capital(id string) string {
	head := strings.ToUpper(id[0:1])
	body := id[1:]
	return head + body
}

func headAndBody(s string) (string, string) {
	head := strings.ToUpper(s[0:1])
	body := strings.ToLower(s[1:])
	return head, body
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
	head, body := headAndBody(name)
	parts := strings.Split(body, "<")
	if len(parts) > 1 {
		impl := parts[1]
		impl = impl[0 : len(impl)-1]
		get := strings.Split(impl, ",")
		body = parts[0]
		for ix := range get {
			h, b := headAndBody(strings.Trim(get[ix], " "))
			body += h + b
		}
	}
	return globalClassPrefix + me.classPrefix + head + body
}

func (me *hmfile) enumNameSpace(id string) string {
	head, body := headAndBody(id)
	return globalEnumPrefix + me.enumPrefix + head + body
}

func (me *hmfile) unionNameSpace(id string) string {
	head, body := headAndBody(id)
	return globalUnionPrefix + me.unionPrefix + "Union" + head + body
}

func (me *hmfile) enumTypeName(base, name string) string {
	head, body := headAndBody(name)
	return base + head + body
}
