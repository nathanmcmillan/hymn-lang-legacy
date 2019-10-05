package main

func (me *cfile) builtin(name string, parameters []*node) string {
	switch name {
	case libPush:
		param0 := me.eval(parameters[0])
		p := param0.asVar(me.hmfile)
		if p.checkIsSlice() {
			uses := p.memberType
			param1 := me.eval(parameters[1])
			if uses.checkIsPointerInC() {
				return "hmlib_slice_push(" + param0.code + ", " + param1.code + ")"
			}
			return "hmlib_slice_push_" + uses.typeSig() + "(" + param0.code + ", " + param1.code + ")"
		}
		panic("argument for push was not an array \"" + p.full + "\"")
	case libLength:
		param := me.eval(parameters[0])
		switch param.getType() {
		case TokenRawString:
			return "((int) strlen(" + param.code + "))"
		case TokenString:
			return "hmlib_string_len_int(" + param.code + ")"
		}
		p := param.asVar(me.hmfile)
		if p.checkIsArray() {
			panic("TODO SIZE OF ARRAY")
		} else if p.checkIsSlice() {
			return "hmlib_slice_len_int(" + param.code + ")"
		}
		panic("argument for echo was " + param.string(0))
	case libOpen:
		param0 := me.eval(parameters[0])
		param1 := me.eval(parameters[1])
		return "fopen(" + param0.code + ", " + param1.code + ")"
	case libEcho:
		param := me.eval(parameters[0])
		switch param.getType() {
		case TokenChar:
			return "printf(\"%c\\n\", " + param.code + ")"
		case TokenString:
			return "printf(\"%s\\n\", " + param.code + ")"
		case TokenRawString:
			return "printf(\"%s\\n\", " + param.code + ")"
		case TokenInt:
			return "printf(\"%d\\n\", " + param.code + ")"
		case TokenInt8:
			return "printf(\"%\" PRId8 \"\\n\", " + param.code + ")"
		case TokenInt16:
			return "printf(\"%\" PRId16 \"\\n\", " + param.code + ")"
		case TokenInt32:
			return "printf(\"%\" PRId32 \"\\n\", " + param.code + ")"
		case TokenInt64:
			return "printf(\"%\" PRId64 \"\\n\", " + param.code + ")"
		case TokenUInt:
			return "printf(\"%u\\n\", " + param.code + ")"
		case TokenUInt8:
			return "printf(\"%\" PRId8 \"\\n\", " + param.code + ")"
		case TokenUInt16:
			return "printf(\"%\" PRId16 \"\\n\", " + param.code + ")"
		case TokenUInt32:
			return "printf(\"%\" PRId32 \"\\n\", " + param.code + ")"
		case TokenUInt64:
			return "printf(\"%\" PRId64 \"\\n\", " + param.code + ")"
		case TokenFloat:
			return "printf(\"%f\\n\", " + param.code + ")"
		case TokenFloat32:
			return "printf(\"%f\\n\", " + param.code + ")"
		case TokenFloat64:
			return "printf(\"%f\\n\", " + param.code + ")"
		case "bool":
			return "printf(\"%s\\n\", " + param.code + " ? \"true\" : \"false\")"
		case TokenLibSize:
			return "printf(\"%zu\\n\", " + param.code + ")"
		}
		panic("argument for echo was " + param.string(0))
	case libToStr:
		param := me.eval(parameters[0])
		switch param.getType() {
		case TokenString:
			panic("redundant string cast")
		case TokenInt:
			return "hmlib_int_to_string(" + param.code + ")"
		case TokenInt8:
			return "hmlib_int8_to_string(" + param.code + ")"
		case TokenInt16:
			return "hmlib_int16_to_string(" + param.code + ")"
		case TokenInt32:
			return "hmlib_int32_to_string(" + param.code + ")"
		case TokenInt64:
			return "hmlib_int64_to_string(" + param.code + ")"
		case TokenUInt:
			return "hmlib_uint_to_string(" + param.code + ")"
		case TokenUInt8:
			return "hmlib_uint8_to_string(" + param.code + ")"
		case TokenUInt16:
			return "hmlib_uint16_to_string(" + param.code + ")"
		case TokenUInt32:
			return "hmlib_uint32_to_string(" + param.code + ")"
		case TokenUInt64:
			return "hmlib_uint64_to_string(" + param.code + ")"
		case TokenFloat:
			return "hmlib_float_to_string(" + param.code + ")"
		case TokenFloat32:
			return "hmlib_float32_to_string(" + param.code + ")"
		case TokenFloat64:
			return "hmlib_float64_to_string(" + param.code + ")"
		case TokenChar:
			return "hmlib_char_to_string(" + param.code + ")"
		case "bool":
			return "(" + param.code + " ? \"true\" : \"false\")"
		}
		panic("argument for string cast was " + param.string(0))
	case libToInt:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_int(" + param.code + ")"
		}
		panic("argument for conversion to int was " + param.string(0))
	case libToInt8:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_int8(" + param.code + ")"
		}
		panic("argument for conversion to int8 was " + param.string(0))
	case libToInt16:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_int16(" + param.code + ")"
		}
		panic("argument for conversion to int16 was " + param.string(0))
	case libToInt32:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_int32(" + param.code + ")"
		}
		panic("argument for conversion to int32 was " + param.string(0))
	case libToInt64:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_int64(" + param.code + ")"
		}
		panic("argument for conversion to int64 was " + param.string(0))
	case libToUInt:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_uint(" + param.code + ")"
		}
		panic("argument for conversion to uint was " + param.string(0))
	case libToUInt8:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_uint8(" + param.code + ")"
		}
		panic("argument for conversion to uint8 was " + param.string(0))
	case libToUInt16:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_uint16(" + param.code + ")"
		}
		panic("argument for conversion to uint16 was " + param.string(0))
	case libToUInt32:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_uint32(" + param.code + ")"
		}
		panic("argument for conversion to uint32 was " + param.string(0))
	case libToUInt64:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_uint64(" + param.code + ")"
		}
		panic("argument for conversion to uint64 was " + param.string(0))
	case libToFloat:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_float(" + param.code + ")"
		}
		panic("argument for conversion to float was " + param.string(0))
	case libToFloat32:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_float32(" + param.code + ")"
		}
		panic("argument for conversion to float32 was " + param.string(0))
	case libToFloat64:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return "hmlib_string_to_float64(" + param.code + ")"
		}
		panic("argument for conversion to float64 was " + param.string(0))
	default:
		return ""
	}
}
