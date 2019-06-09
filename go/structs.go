package main

type token struct {
	is    string
	value string
}

type tokenizer struct {
	stream  *stream
	current string
}

type node struct {
	is    string
	value string
	typed string
	has   []*node
}

type variable struct {
	is   string
	name string
}

type scope struct {
	root      *scope
	variables map[string]*variable
}

type function struct {
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
	rootScope     *scope
	scope         *scope
	imports       map[string]bool
	classes       map[string]*class
	classOrder    []string
	functions     map[string]*function
	functionOrder []string
}

type cfile struct {
	imports       map[string]bool
	classes       map[string]*class
	rootScope     *scope
	scope         *scope
	functions     map[string]*function
	functionOrder []string
}

type parser struct {
	tokens  []*token
	token   *token
	pos     int
	program *program
}

type cnode struct {
	is    string
	value string
	has   []*cnode
	typed string
	code  string
}

func (me *node) string(lv int) string {
	s := ""
	s += fmc(lv) + "{is:" + me.is
	if me.value != "" {
		s += ", value:" + me.value
	}
	s += ", typed:" + me.typed
	if len(me.has) > 0 {
		s += ", has[\n"
		lv++
		for ix, has := range me.has {
			if ix > 0 {
				s += "\n"
			}
			s += has.string(lv)
		}
		lv--
		s += "\n"
		s += fmc(lv) + "]"
	}
	s += "}"
	return s
}

func (me *cnode) string(lv int) string {
	s := ""
	s += fmc(lv) + "{is:" + me.is
	if me.value != "" {
		s += ", value:" + me.value
	}
	s += ", typed:" + me.typed
	s += ", code:" + me.code
	if len(me.has) > 0 {
		s += ", has[\n"
		lv++
		for ix, has := range me.has {
			if ix > 0 {
				s += ",\n"
			}
			s += has.string(lv)
		}
		lv--
		s += "\n"
		s += fmc(lv) + "]"
	}
	s += "}"
	return s
}

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
	p := &program{}
	p.rootScope = scopeInit(nil)
	p.scope = p.rootScope
	p.imports = make(map[string]bool)
	p.classes = make(map[string]*class, 0)
	p.classOrder = make([]string, 0)
	p.functions = make(map[string]*function)
	p.functionOrder = make([]string, 0)
	p.libInit()
	return p
}

func (me *program) pushScope() {
	sc := scopeInit(me.scope)
	me.scope = sc
}

func (me *program) popScope() {
	me.scope = me.scope.root
}

func cfileInit() *cfile {
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

func nodeInit(is string) *node {
	n := &node{}
	n.is = is
	n.has = make([]*node, 0)
	return n
}

func (me *node) push(n *node) {
	me.has = append(me.has, n)
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
