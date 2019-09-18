package main

const (
	libEcho    = "echo"
	libToStr   = "to_str"
	libToInt   = "to_int"
	libToFloat = "to_float"
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

func (me *hmfile) libInit() {
	echo := funcInit(me, libEcho)
	echo.typed = me.typeToVarData("void")
	echo.args = append(echo.args, me.fnArgInit("?", "s", false, false))
	me.functions[libEcho] = echo
	me.types[libEcho] = ""

	str := funcInit(me, libToStr)
	str.typed = me.typeToVarData("string")
	str.args = append(str.args, me.fnArgInit("?", "s", false, false))
	me.functions[libToStr] = str
	me.types[libToStr] = ""

	intfn := funcInit(me, libToInt)
	intfn.typed = me.typeToVarData("int")
	intfn.args = append(intfn.args, me.fnArgInit("?", "s", false, false))
	me.functions[libToInt] = intfn
	me.types[libToInt] = ""

	floatfn := funcInit(me, libToFloat)
	floatfn.typed = me.typeToVarData("float")
	floatfn.args = append(floatfn.args, me.fnArgInit("?", "s", false, false))
	me.functions[libToFloat] = floatfn
	me.types[libToFloat] = ""

	for primitive := range primitives {
		me.types[primitive] = ""
	}
}
