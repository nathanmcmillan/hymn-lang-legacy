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
	name    string
	dfault  string
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
	argDict     map[string]int
	expressions []*node
	typed       string
}

type hasGenerics interface {
	getGenerics() []string
}

type class struct {
	name          string
	variables     map[string]*variable
	variableOrder []string
	generics      []string
	genericsDict  map[string]bool
}

type enum struct {
	name         string
	simple       bool
	types        map[string]*union
	typesOrder   []*union
	generics     []string
	genericsDict map[string]bool
}

type union struct {
	name     string
	types    []string
	generics []string
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
	hmfile             *hmfile
	headPrefix         string
	headIncludeSection string
	headTypeDefSection string
	headTypesSection   string
	headExternSection  string
	headFuncSection    string
	headSuffix         string
	codeFn             []string
	rootScope          *scope
	scope              *scope
	depth              int
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

func unionInit(name string, types []string, generics []string) *union {
	u := &union{}
	u.name = name
	u.types = types
	u.generics = generics
	return u
}

func (me *union) copy() *union {
	u := &union{}
	u.name = me.name
	u.types = make([]string, len(me.types))
	u.generics = make([]string, len(me.generics))
	copy(u.types, me.types)
	copy(u.generics, me.generics)
	return u
}

func enumInit(name string, simple bool, order []*union, dict map[string]*union, generics []string, genericsDict map[string]bool) *enum {
	e := &enum{}
	e.name = name
	e.simple = simple
	e.types = dict
	e.typesOrder = order
	e.generics = generics
	e.genericsDict = genericsDict
	return e
}

func classInit(name string, generics []string, genericsDict map[string]bool) *class {
	c := &class{}
	c.name = name
	c.generics = generics
	c.genericsDict = genericsDict
	return c
}

func (me *class) initMembers(variableOrder []string, variables map[string]*variable) {
	me.variableOrder = variableOrder
	me.variables = variables
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

func (prog *program) hymnFileInit(name string) *hmfile {
	hm := &hmfile{}
	hm.name = name
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
	c.codeFn = make([]string, 0)
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
	echo := funcInit()
	echo.typed = "void"
	echo.args = append(echo.args, me.varInit("?", "s", false, false))
	me.functions["echo"] = echo

	str := funcInit()
	str.typed = "string"
	str.args = append(str.args, me.varInit("?", "s", false, false))
	me.functions["string"] = str

	for primitive := range primitives {
		me.types[primitive] = true
	}
}

func funcInit() *function {
	f := &function{}
	f.args = make([]*variable, 0)
	f.argDict = make(map[string]int)
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

func (me *hmfile) varWithDefaultInit(typed, name string, mutable, pointer bool, dfault string) *variable {
	v := me.varInit(typed, name, mutable, pointer)
	v.dfault = dfault
	return v
}

func (me *variable) copy() *variable {
	v := &variable{}
	v.typed = me.typed
	v.name = me.name
	v.cName = me.name
	v.mutable = me.mutable
	v.pointer = me.pointer
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

func (me *cfile) head() string {
	head := ""
	head += me.headPrefix
	head += me.headIncludeSection
	if len(me.headTypeDefSection) != 0 {
		head += me.headTypeDefSection
		head += "\n"
	}
	head += me.headTypesSection
	head += me.headExternSection
	head += me.headFuncSection
	head += me.headSuffix
	return head
}

func (me *class) getGenerics() []string {
	return me.generics
}

func (me *enum) getGenerics() []string {
	return me.generics
}
