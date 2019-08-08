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

func (me *cfile) pushScope() {
	sc := scopeInit(me.scope)
	me.scope = sc
}

func (me *cfile) popScope() {
	me.scope = me.scope.root
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

func codeNode(is, value, typed string, vdata *varData, code string) *cnode {
	c := &cnode{}
	c.is = is
	c.value = value
	c.typed = typed
	c.vdata = vdata
	c.code = code
	c.has = make([]*cnode, 0)
	return c
}

func (me *cnode) push(n *cnode) {
	me.has = append(me.has, n)
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

type allocData struct {
	useStack bool
	isArray  bool
}
