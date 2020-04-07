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
	libToSizeT   = "sizet"
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
	libPop       = "pop"
	libExit      = "exit"
	libChdir     = "chdir"
	libSubstring = "substring"
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

func (me *hmlib) initPop() *parseError {
	var er *parseError
	fn := funcInit(nil, libPop, nil)
	fn.returns = newdataany()
	a, er := me.fnArgInit("?", "a", false)
	if er != nil {
		return er
	}
	fn.args = append(fn.args, a)
	me.functions[libPop] = fn
	me.types[libPop] = ""
	return nil
}

func (me *hmlib) libs() *parseError {
	me.types = make(map[string]string)
	me.classes = make(map[string]*class)
	me.functions = make(map[string]*function)

	var er *parseError

	if er = me.newLibSimpleIn(libExit, TokenInt, "void"); er != nil {
		return er
	}

	if er = me.newLibSimpleIn(libChdir, TokenString, "void"); er != nil {
		return er
	}

	if er = me.newLibSimpleVardiac(libEcho, "void"); er != nil {
		return er
	}

	if er = me.newLibSimpleVardiac(libFormat, TokenString); er != nil {
		return er
	}

	if er = me.newLibSimplePrint(libPrintf, "void"); er != nil {
		return er
	}

	if er = me.newLibSimplePrint(libPrintln, "void"); er != nil {
		return er
	}

	if er = me.newLibSimplePrint(libSprintf, TokenString); er != nil {
		return er
	}

	if er = me.newLibSimplePrint(libSprintln, TokenString); er != nil {
		return er
	}

	if er = me.newLibSimple(libCat, TokenString); er != nil {
		return er
	}

	if er = me.newLibSimple(libSystem, TokenString); er != nil {
		return er
	}

	if er = me.newLibSimple(libToStr, TokenString); er != nil {
		return er
	}

	if er = me.newLibSimple(libToInt, TokenInt); er != nil {
		return er
	}

	if er = me.newLibSimple(libToInt8, TokenInt8); er != nil {
		return er
	}

	if er = me.newLibSimple(libToInt16, TokenInt16); er != nil {
		return er
	}

	if er = me.newLibSimple(libToInt32, TokenInt32); er != nil {
		return er
	}

	if er = me.newLibSimple(libToInt64, TokenInt64); er != nil {
		return er
	}

	if er = me.newLibSimple(libToSizeT, TokenSizeT); er != nil {
		return er
	}

	if er = me.newLibSimple(libToUInt, TokenUInt); er != nil {
		return er
	}

	if er = me.newLibSimple(libToUInt8, TokenUInt8); er != nil {
		return er
	}

	if er = me.newLibSimple(libToUInt16, TokenUInt16); er != nil {
		return er
	}

	if er = me.newLibSimple(libToUInt32, TokenUInt32); er != nil {
		return er
	}

	if er = me.newLibSimple(libToUInt64, TokenUInt64); er != nil {
		return er
	}

	if er = me.newLibSimple(libToFloat, TokenFloat); er != nil {
		return er
	}

	if er = me.newLibSimple(libToFloat32, TokenFloat32); er != nil {
		return er
	}

	if er = me.newLibSimple(libToFloat64, TokenFloat64); er != nil {
		return er
	}

	if er = me.newLibSimple(libLength, TokenInt); er != nil {
		return er
	}

	if er = me.newLibSimple(libCapacity, TokenInt); er != nil {
		return er
	}

	if er = me.newLibRegular(libWrite, "void", TokenString, TokenString); er != nil {
		return er
	}

	if er = me.newLibRegular(libSubstring, TokenString, TokenString, TokenInt, TokenInt); er != nil {
		return er
	}

	me.initPush()
	me.initPop()

	for primitive := range primitives {
		me.types[primitive] = ""
	}

	return nil
}
