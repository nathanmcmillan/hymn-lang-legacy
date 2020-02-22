package main

import "strconv"

const (
	libEcho      = "echo"
	libFormat    = "format"
	libPrintf    = "printf"
	libPrintln   = "println"
	libSprintf   = "sprintf"
	libSprintln  = "sprintln"
	libSystem    = "system"
	libToStr     = "str"
	libToInt     = "integer"
	libToInt8    = "i8"
	libToInt16   = "i16"
	libToInt32   = "i32"
	libToInt64   = "i64"
	libToUInt    = "uinteger"
	libToUInt8   = "u8"
	libToUInt16  = "u16"
	libToUInt32  = "u32"
	libToUInt64  = "u64"
	libToFloat   = "floating"
	libToFloat32 = "f32"
	libToFloat64 = "f64"
	libOpen      = "open"
	libCat       = "cat"
	libWrite     = "write"
	libLength    = "len"
	libCapacity  = "cap"
	libPush      = "push"
	libExit      = "exit"
	libChdir     = "chdir"
	libSubstring = "substring"
)

// library tokens
const (
	TokenLibSize = "SIZE"
	TokenLibFile = "FILE"
)

type hmlib struct {
	fn        []*function
	types     map[string]string
	classes   map[string]*class
	functions map[string]*function
}

func (me *hmlib) newLibSimple(name string, ret string, params ...string) {
	fn := funcInit(nil, name, nil)
	fn.returns = getdatatype(nil, ret)
	fn.args = append(fn.args, me.fnArgInit("?", "s", false))
	if params != nil {
		fn.args = append(fn.args, me.fnArgInit("?", "s", false))
	}
	me.functions[name] = fn
	me.types[name] = ""
}

func (me *hmlib) newLibRegular(name string, ret string, params ...string) {
	fn := funcInit(nil, name, nil)
	fn.returns = getdatatype(nil, ret)
	if params != nil {
		for ix, p := range params {
			fn.args = append(fn.args, me.fnArgInit(p, "p"+strconv.Itoa(ix), false))
		}
	}
	me.functions[name] = fn
	me.types[name] = ""
}

func (me *hmlib) newLibSimpleIn(name string, in string, ret string) {
	fn := funcInit(nil, name, nil)
	fn.returns = getdatatype(nil, ret)
	fn.args = append(fn.args, me.fnArgInit(in, "s", false))
	me.functions[name] = fn
	me.types[name] = ""
}

func (me *hmlib) newLibSimpleVardiac(name string, ret string) {
	fn := funcInit(nil, name, nil)
	fn.returns = getdatatype(nil, ret)
	fn.argVariadic = me.fnArgInit("?", "a", false)
	me.functions[name] = fn
	me.types[name] = ""
}

func (me *hmlib) newLibSimplePrint(name string, ret string) {
	fn := funcInit(nil, name, nil)
	fn.returns = getdatatype(nil, ret)
	fn.args = append(fn.args, me.fnArgInit(TokenString, "a", false))
	fn.argVariadic = me.fnArgInit("?", "b", false)
	me.functions[name] = fn
	me.types[name] = ""
}

func (me *hmlib) initPush() {
	fn := funcInit(nil, libPush, nil)
	fn.returns = newdataany()
	fn.args = append(fn.args, me.fnArgInit("?", "a", false))
	fn.args = append(fn.args, me.fnArgInit("?", "v", false))
	me.functions[libPush] = fn
	me.types[libPush] = ""
}

func (me *hmlib) initIO() {
	me.types[TokenLibFile] = ""
	order := make([]string, 0)
	dict := make(map[string]int)
	classDef := classInit(nil, TokenLibFile, order, dict, nil)
	me.classes[TokenLibFile] = classDef

	fn := funcInit(nil, libOpen, nil)
	fn.returns = getdatatype(nil, TokenLibFile)
	fn.args = append(fn.args, me.fnArgInit(TokenString, "path", false))
	fn.args = append(fn.args, me.fnArgInit(TokenString, "mode", false))
	me.functions[libOpen] = fn
	me.types[libOpen] = ""

	fnName := "read"
	fn = funcInit(nil, fnName, classDef)
	fn.returns = getdatatype(nil, TokenInt)
	fn.args = append(fn.args, me.fnArgInit(classDef.name, "self", false))
	me.functions[fn.getname()] = fn
	me.types[fn.getname()] = ""

	fnName = "read_line"
	fn = funcInit(nil, fnName, classDef)
	fn.returns = getdatatype(nil, TokenString)
	fn.args = append(fn.args, me.fnArgInit(classDef.name, "self", false))
	me.functions[fn.getname()] = fn
	me.types[fn.getname()] = ""

	fnName = "close"
	fn = funcInit(nil, fnName, classDef)
	fn.returns = newdatavoid()
	fn.args = append(fn.args, me.fnArgInit(classDef.name, "self", false))
	me.functions[fn.getname()] = fn
	me.types[fn.getname()] = ""
}

func (me *hmlib) libs() {
	me.types = make(map[string]string)
	me.classes = make(map[string]*class)
	me.functions = make(map[string]*function)

	me.newLibSimpleIn(libExit, TokenInt, "void")
	me.newLibSimpleIn(libChdir, TokenString, "void")

	me.newLibSimpleVardiac(libEcho, "void")
	me.newLibSimpleVardiac(libFormat, TokenString)

	me.newLibSimplePrint(libPrintf, "void")
	me.newLibSimplePrint(libPrintln, "void")
	me.newLibSimplePrint(libSprintf, TokenString)
	me.newLibSimplePrint(libSprintln, TokenString)

	me.newLibSimple(libCat, TokenString)
	me.newLibSimple(libSystem, TokenString)
	me.newLibSimple(libToStr, TokenString)

	me.newLibSimple(libToInt, TokenInt)
	me.newLibSimple(libToInt8, TokenInt8)
	me.newLibSimple(libToInt16, TokenInt16)
	me.newLibSimple(libToInt32, TokenInt32)
	me.newLibSimple(libToInt64, TokenInt64)

	me.newLibSimple(libToUInt, TokenUInt)
	me.newLibSimple(libToUInt8, TokenUInt8)
	me.newLibSimple(libToUInt16, TokenUInt16)
	me.newLibSimple(libToUInt32, TokenUInt32)
	me.newLibSimple(libToUInt64, TokenUInt64)

	me.newLibSimple(libToFloat, TokenFloat)
	me.newLibSimple(libToFloat32, TokenFloat32)
	me.newLibSimple(libToFloat64, TokenFloat64)

	me.newLibSimple(libLength, TokenInt)
	me.newLibSimple(libCapacity, TokenInt)

	me.newLibRegular(libWrite, "void", TokenString, TokenString)
	me.newLibRegular(libSubstring, TokenString, TokenString, TokenInt, TokenInt)

	me.initIO()
	me.initPush()

	for primitive := range primitives {
		me.types[primitive] = ""
	}
}
