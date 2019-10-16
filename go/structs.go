package main

type scope struct {
	root      *scope
	temp      int
	fn        *function
	variables map[string]*variable
}

type hasGenerics interface {
	getGenerics() []string
}

type program struct {
	out       string
	directory string
	libDir    string
	hmlib     *hmlib
	hmfiles   map[string]*hmfile
	hmorder   []*hmfile
	sources   map[string]string
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
		TokenInt:     true,
		TokenInt8:    true,
		TokenInt16:   true,
		TokenInt32:   true,
		TokenInt64:   true,
		TokenUInt:    true,
		TokenUInt8:   true,
		TokenUInt16:  true,
		TokenUInt32:  true,
		TokenUInt64:  true,
		TokenFloat:   true,
		TokenFloat32: true,
		TokenFloat64: true,
		TokenString:  true,
		TokenChar:    true,
		TokenBoolean: true,
	}
	typeToCName = map[string]string{
		TokenFloat32:   "float",
		TokenFloat64:   "double",
		TokenString:    "hmlib_string",
		TokenRawString: "char *",
		TokenInt8:      "int8_t",
		TokenInt16:     "int16_t",
		TokenInt32:     "int32_t",
		TokenInt64:     "int64_t",
		TokenUInt:      "unsigned int",
		TokenUInt8:     "uint8_t",
		TokenUInt16:    "uint16_t",
		TokenUInt32:    "uint32_t",
		TokenUInt64:    "uint64_t",
		TokenLibSize:   "size_t",
	}
	literals = map[string]string{
		TokenIntLiteral:     TokenInt,
		TokenFloatLiteral:   TokenFloat,
		TokenStringLiteral:  TokenString,
		TokenBooleanLiteral: TokenBoolean,
	}
	numbers = map[string]bool{
		TokenInt:     true,
		TokenInt8:    true,
		TokenInt16:   true,
		TokenInt32:   true,
		TokenInt64:   true,
		TokenUInt:    true,
		TokenUInt8:   true,
		TokenUInt16:  true,
		TokenUInt32:  true,
		TokenUInt64:  true,
		TokenFloat:   true,
		TokenFloat32: true,
		TokenFloat64: true,
	}
	integerTypes = map[string]bool{
		TokenInt:    true,
		TokenInt8:   true,
		TokenInt16:  true,
		TokenInt32:  true,
		TokenInt64:  true,
		TokenUInt:   true,
		TokenUInt8:  true,
		TokenUInt16: true,
		TokenUInt32: true,
		TokenUInt64: true,
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
	prog.hmorder = make([]*hmfile, 0)
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
	// TODO fix me
	// if v, ok := me.scope.variables[name]; ok {
	// 	return v
	// }
	// return me.hmfile.getStatic(name)

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

func codeNode(node *node, code string) *cnode {
	c := &cnode{}
	c.is = node.is
	c.value = node.value
	c.vdata = node.vdata
	c.code = code
	c.has = make([]*cnode, 0)
	return c
}

func (me *cnode) push(n *cnode) {
	me.has = append(me.has, n)
}

func isNumber(t string) bool {
	_, ok := numbers[t]
	return ok
}

func canCastToNumber(t string) bool {
	if t == TokenChar {
		return true
	}
	_, ok := numbers[t]
	return ok
}

func isInteger(t string) bool {
	_, ok := integerTypes[t]
	return ok
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
	stack bool
	array bool
	slice bool
}
