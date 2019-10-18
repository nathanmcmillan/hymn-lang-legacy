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
	libLength    = "len"
	libPush      = "push"
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

func (me *hmlib) simple(name string, ret string) {
	fn := funcInit(nil, name)
	fn.typed = me.literalType(ret)
	fn.args = append(fn.args, me.fnArgInit("?", "s", false))
	me.functions[name] = fn
	me.types[name] = ""
}

func (me *hmlib) initPush() {
	fn := funcInit(nil, libPush)
	fn.typed = me.literalType("?")
	fn.args = append(fn.args, me.fnArgInit("?", "a", false))
	fn.args = append(fn.args, me.fnArgInit("?", "v", false))
	me.functions[libPush] = fn
	me.types[libPush] = ""
}

func (me *hmlib) initIO() {
	me.types[TokenLibFile] = ""
	order := make([]string, 0)
	dict := make(map[string]int)
	classDef := classInit(TokenLibFile, order, dict)
	me.classes[TokenLibFile] = classDef

	fn := funcInit(nil, libOpen)
	fn.typed = me.literalType(TokenLibFile)
	fn.args = append(fn.args, me.fnArgInit(TokenString, "path", false))
	fn.args = append(fn.args, me.fnArgInit(TokenString, "mode", false))
	me.functions[libOpen] = fn
	me.types[libOpen] = ""

	fnName := "read"
	fn = funcInit(nil, fnName)
	fn.typed = me.literalType(TokenInt)
	fn.args = append(fn.args, me.fnArgInit(classDef.name, "self", false))
	fn.forClass = classDef
	name := nameOfClassFunc(TokenLibFile, fnName)
	me.functions[name] = fn
	me.types[name] = ""

	fnName = "read_line"
	fn = funcInit(nil, fnName)
	fn.typed = me.literalType(TokenString)
	fn.args = append(fn.args, me.fnArgInit(classDef.name, "self", false))
	fn.forClass = classDef
	name = nameOfClassFunc(TokenLibFile, fnName)
	me.functions[name] = fn
	me.types[name] = ""

	fnName = "close"
	fn = funcInit(nil, fnName)
	fn.typed = me.literalType("void")
	fn.args = append(fn.args, me.fnArgInit(classDef.name, "self", false))
	fn.forClass = classDef
	name = nameOfClassFunc(TokenLibFile, fnName)
	me.functions[name] = fn
	me.types[name] = ""
}

func (me *hmlib) libs() {
	me.types = make(map[string]string)
	me.classes = make(map[string]*class)
	me.functions = make(map[string]*function)

	me.simple(libEcho, "void")

	me.simple(libToStr, TokenString)

	me.simple(libToInt, TokenInt)
	me.simple(libToInt8, TokenInt8)
	me.simple(libToInt16, TokenInt16)
	me.simple(libToInt32, TokenInt32)
	me.simple(libToInt64, TokenInt64)

	me.simple(libToUInt, TokenUInt)
	me.simple(libToUInt8, TokenUInt8)
	me.simple(libToUInt16, TokenUInt16)
	me.simple(libToUInt32, TokenUInt32)
	me.simple(libToUInt64, TokenUInt64)

	me.simple(libToFloat, TokenFloat)
	me.simple(libToFloat32, TokenFloat32)
	me.simple(libToFloat64, TokenFloat64)

	me.simple(libLength, TokenInt)

	me.initIO()
	me.initPush()

	for primitive := range primitives {
		me.types[primitive] = ""
	}
}
