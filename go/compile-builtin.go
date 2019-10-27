package main

func (me *cfile) compileBuiltin(n *node, name string, parameters []*node) *codeblock {
	switch name {
	case libPush:
		param0 := me.eval(parameters[0])
		p := param0.data()
		if p.checkIsSlice() {
			uses := p.memberType
			param1 := me.eval(parameters[1])
			if uses.checkIsPointerInC() {
				cb := codeBlockOne(n, "hmlib_slice_push("+param0.pop()+", "+param1.pop()+")")
				cb.prepend(param0.pre)
				cb.prepend(param1.pre)
				return cb
			}
			cb := codeBlockOne(n, "hmlib_slice_push_"+uses.typeSig()+"("+param0.pop()+", "+param1.pop()+")")
			cb.prepend(param0.pre)
			cb.prepend(param1.pre)
			return cb
		}
		panic("argument for push was not an array \"" + p.full + "\"")
	case libLength:
		param := me.eval(parameters[0])
		switch param.getType() {
		case TokenRawString:
			return codeBlockMerge(n, "((int) strlen("+param.pop()+"))", param.pre)
		case TokenString:
			return codeBlockMerge(n, "hmlib_string_len_int("+param.pop()+")", param.pre)
		}
		p := param.data()
		if p.checkIsArray() {
			return codeBlockMerge(n, p.sizeOfArray(), param.pre)
		} else if p.checkIsSlice() {
			return codeBlockMerge(n, "hmlib_slice_len_int("+param.pop()+")", param.pre)
		}
		panic("argument for echo was " + param.string(0))
	case libOpen:
		param0 := me.eval(parameters[0])
		param1 := me.eval(parameters[1])
		cb := codeBlockOne(n, "fopen("+param0.pop()+", "+param1.pop()+")")
		cb.prepend(param0.pre)
		cb.prepend(param1.pre)
		return cb
	case libEcho:
		param := me.eval(parameters[0])
		switch param.getType() {
		case TokenChar:
			return codeBlockMerge(n, "printf(\"%c\\n\", "+param.pop()+")", param.pre)
		case TokenString:
			return codeBlockMerge(n, "printf(\"%s\\n\", "+param.pop()+")", param.pre)
		case TokenRawString:
			return codeBlockMerge(n, "printf(\"%s\\n\", "+param.pop()+")", param.pre)
		case TokenInt:
			return codeBlockMerge(n, "printf(\"%d\\n\", "+param.pop()+")", param.pre)
		case TokenInt8:
			return codeBlockMerge(n, "printf(\"%\" PRId8 \"\\n\", "+param.pop()+")", param.pre)
		case TokenInt16:
			return codeBlockMerge(n, "printf(\"%\" PRId16 \"\\n\", "+param.pop()+")", param.pre)
		case TokenInt32:
			return codeBlockMerge(n, "printf(\"%\" PRId32 \"\\n\", "+param.pop()+")", param.pre)
		case TokenInt64:
			return codeBlockMerge(n, "printf(\"%\" PRId64 \"\\n\", "+param.pop()+")", param.pre)
		case TokenUInt:
			return codeBlockMerge(n, "printf(\"%u\\n\", "+param.pop()+")", param.pre)
		case TokenUInt8:
			return codeBlockMerge(n, "printf(\"%\" PRId8 \"\\n\", "+param.pop()+")", param.pre)
		case TokenUInt16:
			return codeBlockMerge(n, "printf(\"%\" PRId16 \"\\n\", "+param.pop()+")", param.pre)
		case TokenUInt32:
			return codeBlockMerge(n, "printf(\"%\" PRId32 \"\\n\", "+param.pop()+")", param.pre)
		case TokenUInt64:
			return codeBlockMerge(n, "printf(\"%\" PRId64 \"\\n\", "+param.pop()+")", param.pre)
		case TokenFloat:
			return codeBlockMerge(n, "printf(\"%f\\n\", "+param.pop()+")", param.pre)
		case TokenFloat32:
			return codeBlockMerge(n, "printf(\"%f\\n\", "+param.pop()+")", param.pre)
		case TokenFloat64:
			return codeBlockMerge(n, "printf(\"%f\\n\", "+param.pop()+")", param.pre)
		case "bool":
			return codeBlockMerge(n, "printf(\"%s\\n\", "+param.pop()+" ? \"true\" : \"false\")", param.pre)
		case TokenLibSize:
			return codeBlockMerge(n, "printf(\"%zu\\n\", "+param.pop()+")", param.pre)
		}
		panic("argument for echo was " + param.string(0))
	case libToStr:
		param := me.eval(parameters[0])
		switch param.getType() {
		case TokenString:
			panic("redundant string cast")
		case TokenInt:
			return codeBlockMerge(n, "hmlib_int_to_string("+param.pop()+")", param.pre)
		case TokenInt8:
			return codeBlockMerge(n, "hmlib_int8_to_string("+param.pop()+")", param.pre)
		case TokenInt16:
			return codeBlockMerge(n, "hmlib_int16_to_string("+param.pop()+")", param.pre)
		case TokenInt32:
			return codeBlockMerge(n, "hmlib_int32_to_string("+param.pop()+")", param.pre)
		case TokenInt64:
			return codeBlockMerge(n, "hmlib_int64_to_string("+param.pop()+")", param.pre)
		case TokenUInt:
			return codeBlockMerge(n, "hmlib_uint_to_string("+param.pop()+")", param.pre)
		case TokenUInt8:
			return codeBlockMerge(n, "hmlib_uint8_to_string("+param.pop()+")", param.pre)
		case TokenUInt16:
			return codeBlockMerge(n, "hmlib_uint16_to_string("+param.pop()+")", param.pre)
		case TokenUInt32:
			return codeBlockMerge(n, "hmlib_uint32_to_string("+param.pop()+")", param.pre)
		case TokenUInt64:
			return codeBlockMerge(n, "hmlib_uint64_to_string("+param.pop()+")", param.pre)
		case TokenFloat:
			return codeBlockMerge(n, "hmlib_float_to_string("+param.pop()+")", param.pre)
		case TokenFloat32:
			return codeBlockMerge(n, "hmlib_float32_to_string("+param.pop()+")", param.pre)
		case TokenFloat64:
			return codeBlockMerge(n, "hmlib_float64_to_string("+param.pop()+")", param.pre)
		case TokenChar:
			return codeBlockMerge(n, "hmlib_char_to_string("+param.pop()+")", param.pre)
		case "bool":
			return codeBlockMerge(n, "("+param.pop()+" ? \"true\" : \"false\")", param.pre)
		}
		panic("argument for string cast was " + param.string(0))
	case libToInt:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_int("+param.pop()+")", param.pre)
		}
		panic("argument for conversion to int was " + param.string(0))
	case libToInt8:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_int8("+param.pop()+")", param.pre)
		}
		panic("argument for conversion to int8 was " + param.string(0))
	case libToInt16:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_int16("+param.pop()+")", param.pre)
		}
		panic("argument for conversion to int16 was " + param.string(0))
	case libToInt32:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_int32("+param.pop()+")", param.pre)
		}
		panic("argument for conversion to int32 was " + param.string(0))
	case libToInt64:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_int64("+param.pop()+")", param.pre)
		}
		panic("argument for conversion to int64 was " + param.string(0))
	case libToUInt:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_uint("+param.pop()+")", param.pre)
		}
		panic("argument for conversion to uint was " + param.string(0))
	case libToUInt8:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_uint8("+param.pop()+")", param.pre)
		}
		panic("argument for conversion to uint8 was " + param.string(0))
	case libToUInt16:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_uint16("+param.pop()+")", param.pre)
		}
		panic("argument for conversion to uint16 was " + param.string(0))
	case libToUInt32:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_uint32("+param.pop()+")", param.pre)
		}
		panic("argument for conversion to uint32 was " + param.string(0))
	case libToUInt64:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_uint64("+param.pop()+")", param.pre)
		}
		panic("argument for conversion to uint64 was " + param.string(0))
	case libToFloat:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_float("+param.pop()+")", param.pre)
		}
		panic("argument for conversion to float was " + param.string(0))
	case libToFloat32:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_float32("+param.pop()+")", param.pre)
		}
		panic("argument for conversion to float32 was " + param.string(0))
	case libToFloat64:
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_float64("+param.pop()+")", param.pre)
		}
		panic("argument for conversion to float64 was " + param.string(0))
	default:
		return nil
	}
}
