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
	typeToStd = map[string]string{
		TokenBoolean: CStdBool,
		TokenInt8:    CStdIntTypes,
		TokenInt16:   CStdIntTypes,
		TokenInt32:   CStdIntTypes,
		TokenInt64:   CStdIntTypes,
		TokenUInt8:   CStdIntTypes,
		TokenUInt16:  CStdIntTypes,
		TokenUInt32:  CStdIntTypes,
		TokenUInt64:  CStdIntTypes,
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
	outputDirectory string
	directory       string
	libs            string
	hmlibmap        map[string]string
	hmlib           *hmlib
	hmfiles         map[string]*hmfile
	modules         map[string]*hmfile
	hmorder         []*hmfile
	sources         map[string]string
	packages        map[string]string
	moduleUID       int
	remapStack      []string
}

func programInit() *program {
	prog := &program{}
	prog.hmlibmap = make(map[string]string)
	prog.hmfiles = make(map[string]*hmfile)
	prog.modules = make(map[string]*hmfile)
	prog.hmorder = make([]*hmfile, 0)
	prog.sources = make(map[string]string)
	prog.packages = make(map[string]string)
	prog.remapStack = make([]string, 0)
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

func (me *program) pushRemapStack(name string) {
	me.remapStack = append(me.remapStack, name)
}

func (me *program) popRemapStack() {
	me.remapStack = me.remapStack[0 : len(me.remapStack)-1]
}

func (me *program) peekRemapStack() string {
	if len(me.remapStack) == 0 {
		return ""
	}
	return me.remapStack[len(me.remapStack)-1]
}
