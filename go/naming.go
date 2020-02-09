package main

import (
	"strings"
)

var (
	globalClassPrefix = "Hm"
	globalEnumPrefix  = globalClassPrefix
	globalUnionPrefix = globalClassPrefix
	globalFuncPrefix  = "hm_"
	globalVarPrefix   = "hm"
	definePrefix      = "HM_"
)

func simpleCapitalize(name string) string {
	return strings.ToUpper(name[0:1]) + name[1:]
}

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
	name = upperSplit(name, ".")
	name = upperSplit(name, "_")
	name = upperSplit(name, "And")
	return name
}

func flatten(name string) string {
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "")
	name = strings.ReplaceAll(name, ",", "_and_")
	name = strings.ReplaceAll(name, ".", "_")
	return name
}

func (me *hmfile) nameSpaceModuleUID(name string) string {
	module := strings.Index(name, "%")
	for module != -1 {
		dot := module + strings.Index(name[module:], ".")
		name = name[0:module] + name[dot+1:]
		module = strings.Index(name, "%")
	}
	return name
}

func (me *hmfile) headerFileGuard(root, name string) string {
	if root != "" {
		root = strings.ToUpper(root)
		root = strings.ReplaceAll(root, "-", "_")
		root += "_"
	}
	name = me.nameSpaceModuleUID(name)
	name = strings.ToUpper(name)
	name = strings.ReplaceAll(name, "-", "_")
	return definePrefix + root + name + "_H"
}

func enumTypeName(base, name string) string {
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

func (me *hmfile) varNameSpace(name string) string {
	return globalVarPrefix + me.varPrefix + capital(name)
}

func (me *hmfile) funcNameSpace(name string) string {
	name = me.nameSpaceModuleUID(name)
	name = flatten(name)
	if !strings.HasPrefix(name, me.funcPrefix) {
		name = me.funcPrefix + name
	}
	return globalFuncPrefix + name
}

func (me *hmfile) classNameSpace(name string) string {
	name = me.nameSpaceModuleUID(name)
	name = capital(name)
	if !strings.HasPrefix(name, me.classPrefix) {
		name = me.classPrefix + name
	}
	return globalClassPrefix + name
}

func (me *hmfile) enumNameSpace(name string) string {
	name = me.nameSpaceModuleUID(name)
	name = capital(name)
	if !strings.HasPrefix(name, me.enumPrefix) {
		name = me.enumPrefix + name
	}
	return globalEnumPrefix + name
}

func (me *hmfile) unionNameSpace(name string) string {
	name = me.nameSpaceModuleUID(name)
	return globalUnionPrefix + me.unionPrefix + capital(name)
}

func (me *hmfile) compileWithFileName(name string) string {
	name = me.nameSpaceModuleUID(name)
	name = flatten(name)
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, ".", "-")
	return name
}

func (me *class) classFileName() string {
	return me.module.compileWithFileName(me.name)
}

func (me *enum) enumFileName() string {
	return me.module.compileWithFileName(me.name)
}
