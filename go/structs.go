package main

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
	is      string
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
	imports       map[string]bool
	classes       map[string]*class
	statics       []*node
	classOrder    []string
	functions     map[string]*function
	functionOrder []string
	types         map[string]bool
}

type parser struct {
	hmfile *hmfile
	tokens *tokenizer
	token  *token
	pos    int
	line   int
}

type cfile struct {
	funcPrefix    string
	classPrefix   string
	varPrefix     string
	imports       map[string]bool
	classes       map[string]*class
	rootScope     *scope
	scope         *scope
	functions     map[string]*function
	functionOrder []string
	types         map[string]bool
	depth         int
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

func classInit(name string, variableOrder []string, variables map[string]*variable) *class {
	c := &class{}
	c.name = name
	c.variableOrder = variableOrder
	c.variables = variables
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

func hymnFileInit(prog *program) *hmfile {
	hm := &hmfile{}
	hm.program = prog
	hm.rootScope = scopeInit(nil)
	hm.scope = hm.rootScope
	hm.types = make(map[string]bool)
	hm.imports = make(map[string]bool)
	hm.classes = make(map[string]*class, 0)
	hm.statics = make([]*node, 0)
	hm.classOrder = make([]string, 0)
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

func cFileInit() *cfile {
	c := &cfile{}
	c.imports = make(map[string]bool)
	c.rootScope = scopeInit(nil)
	c.scope = c.rootScope
	c.functions = make(map[string]*function)
	c.functionOrder = make([]string, 0)
	c.classes = make(map[string]*class, 0)
	return c
}

func (me *cfile) pushScope() {
	sc := scopeInit(me.scope)
	me.scope = sc
}

func (me *cfile) popScope() {
	me.scope = me.scope.root
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
	e.args = append(e.args, varInit("?", "s", false, false))
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

func varInit(is, name string, mutable, pointer bool) *variable {
	v := &variable{}
	v.is = is
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
