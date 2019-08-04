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

type scope struct {
	root      *scope
	temp      int
	fn        *function
	variables map[string]*variable
}

type function struct {
	name        string
	args        []*variable
	argDict     map[string]int
	expressions []*node
	typed       *varData
}

type hasGenerics interface {
	getGenerics() []string
}

type program struct {
	out       string
	directory string
	libDir    string
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
	echo.typed = me.typeToVarData("void")
	echo.args = append(echo.args, me.varInit("?", "s", false, false))
	me.functions["echo"] = echo

	str := funcInit()
	str.typed = me.typeToVarData("string")
	str.args = append(str.args, me.varInit("?", "s", false, false))
	me.functions["string"] = str

	intfn := funcInit()
	intfn.typed = me.typeToVarData("int")
	intfn.args = append(intfn.args, me.varInit("?", "s", false, false))
	me.functions["int"] = intfn

	floatfn := funcInit()
	floatfn.typed = me.typeToVarData("float")
	floatfn.args = append(floatfn.args, me.varInit("?", "s", false, false))
	me.functions["float"] = floatfn

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

func isNumber(t string) bool {
	return t == "int" || t == "float"
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
