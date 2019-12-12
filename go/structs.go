package main

import "strings"

type scope struct {
	root      *scope
	temp      int
	fn        *function
	variables map[string]*variable
	renaming  map[string]string
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
	hmfile                   *hmfile
	headPrefix               strings.Builder
	headIncludeSection       strings.Builder
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
	s.renaming = make(map[string]string)
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

func codeNode(node *node, code string) *cnode {
	c := &cnode{}
	c.is = node.is
	c.value = node.value
	c.copyData(node.data())
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
	var head strings.Builder
	head.WriteString(me.headPrefix.String())
	head.WriteString(me.headIncludeSection.String())
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
	head.WriteString(me.headExternSection.String())
	head.WriteString(me.headFuncSection.String())
	head.WriteString(me.headSuffix.String())
	return head.String()
}

type allocData struct {
	stack bool
	array bool
	slice bool
	size  int
}
