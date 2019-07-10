package main

import (
	"fmt"
	"strings"
)

type token struct {
	depth int
	is    string
	value string
}

type tokenizer struct {
	stream      *stream
	current     string
	tokens      []*token
	eof         *token
	size        int
	depth       int
	updateDepth bool
}

type node struct {
	is         string
	value      string
	typed      string
	attributes []string
	has        []*node
}

type variable struct {
	typed   string
	members map[string]string
	name    string
	mutable bool
	pointer bool
	cName   string
}

type scope struct {
	root      *scope
	fn        *function
	variables map[string]*variable
}

type function struct {
	name        string
	args        []*variable
	expressions []*node
	typed       string
}

type class struct {
	name          string
	variables     map[string]*variable
	variableOrder []string
	generics      []string
}

type enum struct {
	name       string
	simple     bool
	types      map[string]*union
	typesOrder []*union
	generics   []string
}

type union struct {
	name  string
	types []string
}

type program struct {
	out       string
	directory string
	hmfiles   map[string]*hmfile
	sources   map[string]string
}

type hmfile struct {
	program       *program
	name          string
	rootScope     *scope
	scope         *scope
	staticScope   map[string]*variable
	namespace     map[string]string
	imports       map[string]bool
	classes       map[string]*class
	enums         map[string]*enum
	statics       []*node
	defineOrder   []string
	functions     map[string]*function
	functionOrder []string
	types         map[string]bool
	funcPrefix    string
	classPrefix   string
	enumPrefix    string
	unionPrefix   string
	varPrefix     string
}

type parser struct {
	hmfile *hmfile
	tokens *tokenizer
	token  *token
	pos    int
	line   int
}

type cfile struct {
	hmfile    *hmfile
	rootScope *scope
	scope     *scope
	depth     int
}

type cnode struct {
	is    string
	value string
	has   []*cnode
	typed string
	code  string
}

var (
	primitives = map[string]bool{
		"int":    true,
		"string": true,
		"bool":   true,
		"float":  true,
	}
)

func unionInit(name string, types []string) *union {
	u := &union{}
	u.name = name
	u.types = types
	return u
}

func enumInit(name string, simple bool, order []*union, dict map[string]*union, generics []string) *enum {
	e := &enum{}
	e.name = name
	e.simple = simple
	e.types = dict
	e.typesOrder = order
	e.generics = generics
	return e
}

func classInit(name string, variableOrder []string, variables map[string]*variable, generics []string) *class {
	c := &class{}
	c.name = name
	c.variableOrder = variableOrder
	c.variables = variables
	c.generics = generics
	return c
}

func scopeInit(root *scope) *scope {
	s := &scope{}
	s.root = root
	s.variables = make(map[string]*variable)
	return s
}

func programInit() *program {
	prog := &program{}
	prog.hmfiles = make(map[string]*hmfile)
	prog.sources = make(map[string]string)
	return prog
}

func (prog *program) hymnFileInit() *hmfile {
	hm := &hmfile{}
	hm.program = prog
	hm.rootScope = scopeInit(nil)
	hm.scope = hm.rootScope
	hm.staticScope = make(map[string]*variable)
	hm.namespace = make(map[string]string)
	hm.types = make(map[string]bool)
	hm.imports = make(map[string]bool)
	hm.classes = make(map[string]*class, 0)
	hm.enums = make(map[string]*enum, 0)
	hm.statics = make([]*node, 0)
	hm.defineOrder = make([]string, 0)
	hm.functions = make(map[string]*function)
	hm.functionOrder = make([]string, 0)
	hm.libInit()
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
	return c
}

func (me *cfile) pushScope() {
	sc := scopeInit(me.scope)
	me.scope = sc
}

func (me *cfile) popScope() {
	me.scope = me.scope.root
}

func (me *hmfile) getstatic(name string) *variable {
	if s, ok := me.staticScope[name]; ok {
		return s
	}
	return nil
}

func (me *hmfile) getvar(name string) *variable {
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

func (me *cfile) getvar(name string) *variable {
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

func nodeInit(is string) *node {
	n := &node{}
	n.is = is
	n.has = make([]*node, 0)
	n.attributes = make([]string, 0)
	return n
}

func (me *node) push(n *node) {
	me.has = append(me.has, n)
}

func (me *node) attribute(find string) bool {
	for _, a := range me.attributes {
		if a == find {
			return true
		}
	}
	return false
}

func (me *node) pushAttribute(a string) {
	me.attributes = append(me.attributes, a)
}

func codeNode(is, value, typed, code string) *cnode {
	c := &cnode{}
	c.is = is
	c.value = value
	c.typed = typed
	c.code = code
	c.has = make([]*cnode, 0)
	return c
}

func (me *cnode) push(n *cnode) {
	me.has = append(me.has, n)
}

func (me *hmfile) libInit() {
	e := funcInit()
	e.typed = "void"
	e.args = append(e.args, me.varInit("?", "s", false, false))
	me.functions["echo"] = e

	for primitive := range primitives {
		me.types[primitive] = true
	}
}

func funcInit() *function {
	f := &function{}
	f.args = make([]*variable, 0)
	f.expressions = make([]*node, 0)
	return f
}

func (me *hmfile) varInit(typed, name string, mutable, pointer bool) *variable {
	v := &variable{}
	v.typed = typed
	v.name = name
	v.cName = name
	v.mutable = mutable
	v.pointer = pointer
	return v
}

func (me *variable) memget() string {
	if me.pointer {
		return "->"
	}
	return "."
}

func isNumber(t string) bool {
	return t == "int" || t == "float"
}

func (me *hmfile) varNameSpace(id string) string {
	return globalVarPrefix + me.varPrefix + capital(id)
}

func (me *hmfile) funcNameSpace(id string) string {
	return globalFuncPrefix + me.funcPrefix + id
}

func (me *hmfile) classNameSpace(id string) string {
	head := strings.ToUpper(id[0:1])
	body := strings.ToLower(id[1:])
	return globalClassPrefix + me.classPrefix + head + body
}

func (me *hmfile) enumNameSpace(id string) string {
	head := strings.ToUpper(id[0:1])
	body := strings.ToLower(id[1:])
	return globalEnumPrefix + me.enumPrefix + head + body
}

func (me *hmfile) unionNameSpace(id string) string {
	head := strings.ToUpper(id[0:1])
	body := strings.ToLower(id[1:])
	return globalUnionPrefix + me.unionPrefix + "Union" + head + body
}

func (me *hmfile) enumTypeName(base, name string) string {
	head := strings.ToUpper(name[0:1])
	body := strings.ToLower(name[1:])
	return base + head + body
}

func (me *hmfile) moduleAndName(name string) (*hmfile, string) {
	if checkIsArray(name) {
		name = typeOfArray(name)
	}
	get := strings.Split(name, ".")
	if len(get) == 1 {
		return me, get[0]
	}
	module := me.program.hmfiles[get[0]]
	return module, get[1]
}

func (me *hmfile) getclass(name string) (*class, string) {
	ix := strings.Index(name, "[")
	if ix == -1 {
		fmt.Println("getclass", name)
		cl, _ := me.classes[name]
		return cl, ""
	}
	get0 := name[0:ix]
	get1 := name[ix+1 : len(name)-1]
	fmt.Println("getclass", get0, "->", get1)
	cl, _ := me.classes[get0]
	return cl, get1
}
