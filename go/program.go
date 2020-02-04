package main

import "strings"

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
		TokenChar:    true,
		TokenBoolean: true,
		TokenString:  true,
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
		TokenCharLiteral:    TokenChar,
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

type program struct {
	out       string
	directory string
	libs      string
	hmlibmap  map[string]string
	hmlib     *hmlib
	hmfiles   map[string]*hmfile
	modules   map[string]*hmfile
	hmorder   []*hmfile
	sources   map[string]string
	shellvar  map[string]string
	moduleUID int
}

func programInit() *program {
	prog := &program{}
	prog.hmlibmap = make(map[string]string)
	prog.hmfiles = make(map[string]*hmfile)
	prog.modules = make(map[string]*hmfile)
	prog.hmorder = make([]*hmfile, 0)
	prog.sources = make(map[string]string)
	prog.shellvar = make(map[string]string)
	return prog
}

func (me *program) loadlibs(hmlibs string) {
	hmlibls := scan(hmlibs)
	for _, f := range hmlibls {
		name := f.Name()
		if strings.HasSuffix(name, ".c") {
			base := name[0:strings.LastIndex(name, ".c")]
			me.hmlibmap[base] = hmlibs + "/" + name
		}
	}
}
