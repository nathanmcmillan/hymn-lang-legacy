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

func (me *hmlib) newLibSimple(name string, ret string, params ...string) *parseError {
	var er *parseError
	fn := funcInit(nil, name, nil)
	fn.returns, er = getdatatype(nil, ret)
	if er != nil {
		return er
	}
	a, er := me.fnArgInit("?", "s", false)
	if er != nil {
		return er
	}
	fn.args = append(fn.args, a)
	if params != nil {
		a, er := me.fnArgInit("?", "s", false)
		if er != nil {
			return er
		}
		fn.args = append(fn.args, a)
	}
	me.functions[name] = fn
	me.types[name] = ""
	return nil
}

func (me *hmlib) newLibRegular(name string, ret string, params ...string) *parseError {
	var er *parseError
	fn := funcInit(nil, name, nil)
	fn.returns, er = getdatatype(nil, ret)
	if er != nil {
		return er
	}
	if params != nil {
		for ix, p := range params {
			a, er := me.fnArgInit(p, "p"+strconv.Itoa(ix), false)
			if er != nil {
				return er
			}
			fn.args = append(fn.args, a)
		}
	}
	me.functions[name] = fn
	me.types[name] = ""
	return nil
}

func (me *hmlib) newLibSimpleIn(name string, in string, ret string) *parseError {
	var er *parseError
	fn := funcInit(nil, name, nil)
	fn.returns, er = getdatatype(nil, ret)
	if er != nil {
		return er
	}
	a, er := me.fnArgInit(in, "s", false)
	if er != nil {
		return er
	}
	fn.args = append(fn.args, a)
	me.functions[name] = fn
	me.types[name] = ""
	return nil
}

func (me *hmlib) newLibSimpleVardiac(name string, ret string) *parseError {
	var er *parseError
	fn := funcInit(nil, name, nil)
	fn.returns, er = getdatatype(nil, ret)
	if er != nil {
		return er
	}
	a, er := me.fnArgInit("?", "a", false)
	fn.argVariadic = a
	me.functions[name] = fn
	me.types[name] = ""
	return nil
}

func (me *hmlib) newLibSimplePrint(name string, ret string) *parseError {
	var er *parseError
	fn := funcInit(nil, name, nil)
	fn.returns, er = getdatatype(nil, ret)
	if er != nil {
		return er
	}
	a, er := me.fnArgInit(TokenString, "a", false)
	if er != nil {
		return er
	}
	fn.args = append(fn.args, a)
	b, er := me.fnArgInit("?", "b", false)
	if er != nil {
		return er
	}
	fn.argVariadic = b
	me.functions[name] = fn
	me.types[name] = ""
	return nil
}

func (me *hmlib) initPush() *parseError {
	var er *parseError
	fn := funcInit(nil, libPush, nil)
	fn.returns = newdataany()
	a, er := me.fnArgInit("?", "a", false)
	if er != nil {
		return er
	}
	fn.args = append(fn.args, a)
	v, er := me.fnArgInit("?", "v", false)
	if er != nil {
		return er
	}
	fn.args = append(fn.args, v)
	me.functions[libPush] = fn
	me.types[libPush] = ""
	return nil
}

func (me *hmlib) initIO() *parseError {
	var er *parseError

	me.types[TokenLibFile] = ""
	classDef := classInit(nil, TokenLibFile, make([]string, 0), make(map[string][]*classInterface), nil)
	me.classes[TokenLibFile] = classDef

	fn := funcInit(nil, libOpen, nil)
	fn.returns, er = getdatatype(nil, TokenLibFile)
	if er != nil {
		return er
	}
	a, er := me.fnArgInit(TokenString, "path", false)
	if er != nil {
		return er
	}
	fn.args = append(fn.args, a)
	m, er := me.fnArgInit(TokenString, "mode", false)
	if er != nil {
		return er
	}
	fn.args = append(fn.args, m)
	me.functions[libOpen] = fn
	me.types[libOpen] = ""

	fnName := "read"
	fn = funcInit(nil, fnName, classDef)
	fn.returns, er = getdatatype(nil, TokenInt)
	if er != nil {
		return er
	}
	s, er := me.fnArgInit(classDef.name, "self", false)
	if er != nil {
		return er
	}
	fn.args = append(fn.args, s)
	me.functions[fn.getname()] = fn
	me.types[fn.getname()] = ""

	fnName = "read_line"
	fn = funcInit(nil, fnName, classDef)
	fn.returns, er = getdatatype(nil, TokenString)
	if er != nil {
		return er
	}
	s, er = me.fnArgInit(classDef.name, "self", false)
	if er != nil {
		return er
	}
	fn.args = append(fn.args, s)
	me.functions[fn.getname()] = fn
	me.types[fn.getname()] = ""

	fnName = "close"
	fn = funcInit(nil, fnName, classDef)
	fn.returns = newdatavoid()
	a, er = me.fnArgInit(classDef.name, "self", false)
	if er != nil {
		return er
	}
	fn.args = append(fn.args, a)
	me.functions[fn.getname()] = fn
	me.types[fn.getname()] = ""

	return nil
}

func (me *hmlib) libs() *parseError {
	me.types = make(map[string]string)
	me.classes = make(map[string]*class)
	me.functions = make(map[string]*function)

	var er *parseError

	er = me.newLibSimpleIn(libExit, TokenInt, "void")
	if er != nil {
		return er
	}

	er = me.newLibSimpleIn(libChdir, TokenString, "void")
	if er != nil {
		return er
	}

	er = me.newLibSimpleVardiac(libEcho, "void")
	if er != nil {
		return er
	}

	er = me.newLibSimpleVardiac(libFormat, TokenString)
	if er != nil {
		return er
	}

	er = me.newLibSimplePrint(libPrintf, "void")
	if er != nil {
		return er
	}

	er = me.newLibSimplePrint(libPrintln, "void")
	if er != nil {
		return er
	}

	er = me.newLibSimplePrint(libSprintf, TokenString)
	if er != nil {
		return er
	}

	er = me.newLibSimplePrint(libSprintln, TokenString)
	if er != nil {
		return er
	}

	er = me.newLibSimple(libCat, TokenString)
	if er != nil {
		return er
	}

	er = me.newLibSimple(libSystem, TokenString)
	if er != nil {
		return er
	}

	er = me.newLibSimple(libToStr, TokenString)
	if er != nil {
		return er
	}

	er = me.newLibSimple(libToInt, TokenInt)
	if er != nil {
		return er
	}

	er = me.newLibSimple(libToInt8, TokenInt8)
	if er != nil {
		return er
	}

	er = me.newLibSimple(libToInt16, TokenInt16)
	if er != nil {
		return er
	}

	er = me.newLibSimple(libToInt32, TokenInt32)
	if er != nil {
		return er
	}

	er = me.newLibSimple(libToInt64, TokenInt64)
	if er != nil {
		return er
	}

	er = me.newLibSimple(libToUInt, TokenUInt)
	if er != nil {
		return er
	}

	er = me.newLibSimple(libToUInt8, TokenUInt8)
	if er != nil {
		return er
	}

	er = me.newLibSimple(libToUInt16, TokenUInt16)
	if er != nil {
		return er
	}

	er = me.newLibSimple(libToUInt32, TokenUInt32)
	if er != nil {
		return er
	}

	er = me.newLibSimple(libToUInt64, TokenUInt64)
	if er != nil {
		return er
	}

	er = me.newLibSimple(libToFloat, TokenFloat)
	if er != nil {
		return er
	}

	er = me.newLibSimple(libToFloat32, TokenFloat32)
	if er != nil {
		return er
	}

	er = me.newLibSimple(libToFloat64, TokenFloat64)
	if er != nil {
		return er
	}

	er = me.newLibSimple(libLength, TokenInt)
	if er != nil {
		return er
	}

	er = me.newLibSimple(libCapacity, TokenInt)
	if er != nil {
		return er
	}

	er = me.newLibRegular(libWrite, "void", TokenString, TokenString)
	if er != nil {
		return er
	}

	er = me.newLibRegular(libSubstring, TokenString, TokenString, TokenInt, TokenInt)
	if er != nil {
		return er
	}

	me.initIO()
	me.initPush()

	for primitive := range primitives {
		me.types[primitive] = ""
	}

	return nil
}
