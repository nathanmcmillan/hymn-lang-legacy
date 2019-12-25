package main

import "strings"

type hmfile struct {
	out             string
	path            string
	libs            string
	parser          *parser
	program         *program
	hmlib           *hmlib
	name            string
	rootScope       *scope
	scope           *scope
	staticScope     map[string]*variable
	namespace       map[string]string
	imports         map[string]*hmfile
	importOrder     []string
	classes         map[string]*class
	enums           map[string]*enum
	defs            map[string]*node
	statics         []*node
	defineOrder     []string
	functions       map[string]*function
	functionOrder   []string
	types           map[string]string
	funcPrefix      string
	classPrefix     string
	enumPrefix      string
	unionPrefix     string
	varPrefix       string
	needInit        bool
	assignmentStack []*node
	enumIsStack     []*variableNode
}

func (program *program) hymnFileInit(name string) *hmfile {
	hm := &hmfile{}
	hm.name = name
	hm.program = program
	hm.hmlib = program.hmlib
	hm.rootScope = scopeInit(nil)
	hm.scope = hm.rootScope
	hm.staticScope = make(map[string]*variable)
	hm.namespace = make(map[string]string)
	hm.types = make(map[string]string)
	hm.imports = make(map[string]*hmfile)
	hm.importOrder = make([]string, 0)
	hm.classes = make(map[string]*class)
	hm.enums = make(map[string]*enum)
	hm.defs = make(map[string]*node)
	hm.statics = make([]*node, 0)
	hm.defineOrder = make([]string, 0)
	hm.functions = make(map[string]*function)
	hm.functionOrder = make([]string, 0)
	hm.assignmentStack = make([]*node, 0)
	hm.enumIsStack = make([]*variableNode, 0)
	hm.prefixes(name)
	return hm
}

func (me *hmfile) pushScope() {
	sc := scopeInit(me.scope)
	me.scope = sc
}

func (me *hmfile) popScope() {
	me.scope = me.scope.root
}

func (me *hmfile) cFileInit() *cfile {
	c := &cfile{}
	c.hmfile = me
	c.rootScope = scopeInit(nil)
	c.scope = c.rootScope
	c.codeFn = make([]strings.Builder, 0)
	return c
}

func (me *hmfile) getStatic(name string) *variable {
	if s, ok := me.staticScope[name]; ok {
		return s
	}
	return nil
}

func (me *hmfile) getvar(name string) *variable {
	scope := me.scope
	for {
		if v, ok := scope.variables[name]; ok {
			fn := me.scope.fn
			if fn != nil {
				if id, ok := fn.argDict[name]; ok {
					fn.args[id].used = true
				}
			}
			return v
		}
		if scope.root == nil {
			return nil
		}
		scope = scope.root
	}
}

func (me *hmfile) getType(name string) (string, bool) {
	if x, ok := me.types[name]; ok {
		return x, true
	}
	if x, ok := me.hmlib.types[name]; ok {
		return x, true
	}
	return "", false
}

func (me *hmfile) getFunction(name string) (*function, bool) {
	if x, ok := me.functions[name]; ok {
		return x, true
	}
	if x, ok := me.hmlib.functions[name]; ok {
		return x, true
	}
	return nil, false
}

func (me *hmfile) getClass(name string) (*class, bool) {
	if x, ok := me.classes[name]; ok {
		return x, true
	}
	if x, ok := me.hmlib.classes[name]; ok {
		return x, true
	}
	return nil, false
}

func (me *hmfile) getEnum(name string) (*enum, bool) {
	if x, ok := me.enums[name]; ok {
		return x, true
	}
	return nil, false
}
