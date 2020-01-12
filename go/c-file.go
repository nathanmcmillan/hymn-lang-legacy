package main

import (
	"strings"
)

type cfile struct {
	location                 string
	hmfile                   *hmfile
	stdReq                   *OrderedSet
	libReq                   *OrderedSet
	dependencyReq            *OrderedSet
	structReq                *OrderedSet
	enumReq                  *OrderedSet
	headStdIncludeSection    strings.Builder
	headLibIncludeSection    strings.Builder
	headReqIncludeSection    strings.Builder
	headEnumTypeDefSection   strings.Builder
	headEnumSection          strings.Builder
	headStructTypeDefSection strings.Builder
	headStructSection        strings.Builder
	headExternSection        strings.Builder
	headFuncSection          strings.Builder
	headSuffix               strings.Builder
	codeFn                   []strings.Builder
	rootScope                *scope
	scope                    *scope
	depth                    int
}

func (me *cfile) pushScope() {
	sc := scopeInit(me.scope)
	me.scope = sc
}

func (me *cfile) popScope() {
	me.scope = me.scope.root
}

func (me *cfile) getvar(name string) *variable {
	// TODO fix scoping rules

	if alias, ok := me.scope.renaming[name]; ok {
		name = alias
	}

	scope := me.scope
	for {
		if v, ok := scope.variables[name]; ok {
			return v
		}
		if scope.root == nil {
			return nil
		}
		scope = scope.root
	}
}

func (me *cfile) head() string {
	me.includeLibs()
	var head strings.Builder
	if me.headStdIncludeSection.Len() != 0 {
		head.WriteString(me.headStdIncludeSection.String())
		head.WriteString("\n")
	}
	if me.headLibIncludeSection.Len() != 0 {
		head.WriteString(me.headLibIncludeSection.String())
		head.WriteString("\n")
	}
	if me.headReqIncludeSection.Len() != 0 {
		head.WriteString(me.headReqIncludeSection.String())
		head.WriteString("\n")
	}
	head.WriteString(me.headEnumSection.String())
	if me.headEnumTypeDefSection.Len() != 0 {
		head.WriteString(me.headEnumTypeDefSection.String())
		head.WriteString("\n")
	}
	if me.headStructTypeDefSection.Len() != 0 {
		head.WriteString(me.headStructTypeDefSection.String())
		head.WriteString("\n")
	}
	head.WriteString(me.headStructSection.String())
	if me.headExternSection.Len() != 0 {
		head.WriteString(me.headExternSection.String())
		head.WriteString("\n")
	}
	head.WriteString(me.headFuncSection.String())
	head.WriteString(me.headSuffix.String())
	return head.String()
}

func (me *cfile) includeLibs() {
	for _, name := range me.stdReq.order {
		me.headStdIncludeSection.WriteString("\n#include <" + name + ".h>")
	}
	for _, name := range me.libReq.order {
		location := me.hmfile.program.hmlibmap[name]
		me.hmfile.program.sources[name] = location
		me.headLibIncludeSection.WriteString("\n#include \"" + name + ".h\"")
	}
	me.dependencyReq.delete(me.location)
	for _, name := range me.dependencyReq.order {
		me.headReqIncludeSection.WriteString("\n#include \"" + name + ".h\"")
	}
}
