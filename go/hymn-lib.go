package main

type hmlib struct {
	fn []*function
}

func (me *hmfile) mkBuiltIn(name string, ret string) {
	fn := funcInit(me, name)
	fn.typed = me.typeToVarData(ret)
	fn.args = append(fn.args, me.fnArgInit("?", "s", false, false))
	me.functions[name] = fn
	me.types[name] = ""
}

func (me *hmfile) libFiles() {
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

	me.libFiles()

	for primitive := range primitives {
		me.types[primitive] = ""
	}
}
