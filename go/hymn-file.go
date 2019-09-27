package main

const (
	libEcho      = "echo"
	libToStr     = "to_str"
	libToInt     = "to_int"
	libToInt8    = "to_int8"
	libToInt16   = "to_int16"
	libToInt32   = "to_int32"
	libToInt64   = "to_int64"
	libToUInt    = "to_uint"
	libToUInt8   = "to_uint8"
	libToUInt16  = "to_uint16"
	libToUInt32  = "to_uint32"
	libToUInt64  = "to_uint64"
	libToFloat   = "to_float"
	libToFloat32 = "to_float32"
	libToFloat64 = "to_float64"
	libOpen      = "open"
)

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
	defs          map[string]*node
	statics       []*node
	defineOrder   []string
	functions     map[string]*function
	functionOrder []string
	types         map[string]string
	funcPrefix    string
	classPrefix   string
	enumPrefix    string
	unionPrefix   string
	varPrefix     string
}

func (prog *program) hymnFileInit(name string) *hmfile {
	hm := &hmfile{}
	hm.name = name
	hm.program = prog
	hm.rootScope = scopeInit(nil)
	hm.scope = hm.rootScope
	hm.staticScope = make(map[string]*variable)
	hm.namespace = make(map[string]string)
	hm.types = make(map[string]string)
	hm.imports = make(map[string]bool)
	hm.classes = make(map[string]*class)
	hm.enums = make(map[string]*enum)
	hm.defs = make(map[string]*node)
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

func (me *hmfile) mkBuiltIn(name string, ret string) {
	fn := funcInit(me, name)
	fn.typed = me.typeToVarData(ret)
	fn.args = append(fn.args, me.fnArgInit("?", "s", false, false))
	me.functions[name] = fn
	me.types[name] = ""
}

func (me *hmfile) libInitOpenFiles() {
	const LibFileType = "FILE"
	me.namespace[LibFileType] = "type"
	me.types[LibFileType] = ""
	me.defineOrder = append(me.defineOrder, LibFileType+"_type")
	order := make([]string, 0)
	dict := make(map[string]bool, 0)
	classDef := classInit(LibFileType, order, dict)
	me.classes[LibFileType] = classDef

	fn := funcInit(me, libOpen)
	fn.typed = me.literalType(LibFileType)
	fn.args = append(fn.args, me.fnArgInit(TokenString, "path", false, false))
	fn.args = append(fn.args, me.fnArgInit(TokenString, "mode", false, false))
	me.functions[libOpen] = fn
	me.types[libOpen] = ""

	fnName := "read"
	fn = funcInit(me, fnName)
	fn.typed = me.literalType(TokenInt)
	fn.args = append(fn.args, me.fnArgInit(classDef.name, "self", false, true))
	fn.forClass = classDef
	name := nameOfClassFunc(LibFileType, fnName)
	me.functionOrder = append(me.functionOrder, name)
	me.functions[name] = fn
	me.types[name] = ""

	fnName = "read_line"
	fn = funcInit(me, fnName)
	fn.typed = me.literalType(TokenString)
	fn.args = append(fn.args, me.fnArgInit(classDef.name, "self", false, true))
	fn.forClass = classDef
	name = nameOfClassFunc(LibFileType, fnName)
	me.functionOrder = append(me.functionOrder, name)
	me.functions[name] = fn
	me.types[name] = ""

	fnName = "close"
	fn = funcInit(me, fnName)
	fn.typed = me.literalType("void")
	fn.args = append(fn.args, me.fnArgInit(classDef.name, "self", false, true))
	fn.forClass = classDef
	name = nameOfClassFunc(LibFileType, fnName)
	me.functionOrder = append(me.functionOrder, name)
	me.functions[name] = fn
	me.types[name] = ""
}

func (me *hmfile) libInit() {
	me.mkBuiltIn(libEcho, "void")

	me.mkBuiltIn(libToStr, TokenString)

	me.mkBuiltIn(libToInt, TokenInt)
	me.mkBuiltIn(libToInt8, TokenInt8)
	me.mkBuiltIn(libToInt16, TokenInt16)
	me.mkBuiltIn(libToInt32, TokenInt32)
	me.mkBuiltIn(libToInt64, TokenInt64)

	me.mkBuiltIn(libToUInt, TokenUInt)
	me.mkBuiltIn(libToUInt8, TokenUInt8)
	me.mkBuiltIn(libToUInt16, TokenUInt16)
	me.mkBuiltIn(libToUInt32, TokenUInt32)
	me.mkBuiltIn(libToUInt64, TokenUInt64)

	me.mkBuiltIn(libToFloat, TokenFloat)
	me.mkBuiltIn(libToFloat32, TokenFloat32)
	me.mkBuiltIn(libToFloat64, TokenFloat64)

	me.libInitOpenFiles()

	for primitive := range primitives {
		me.types[primitive] = ""
	}
}
