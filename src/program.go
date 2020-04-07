package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

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
		TokenSizeT:   true,
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
		TokenSizeT:   CStdIntTypes,
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
		TokenSizeT:   true,
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
		TokenSizeT:  true,
	}
)

type program struct {
	outsourcedir string
	directory    string
	libc         string
	hmlibmap     map[string]string
	hmlib        *hmlib
	hmfiles      map[string]*hmfile
	modules      map[string]*hmfile
	hmorder      []*hmfile
	sources      map[string]string
	packages     map[string]string
	moduleUID    int
	remapStack   []string
	testing      bool
	classes      map[string]*class
	interfaces   map[string]*classInterface
	enums        map[string]*enum
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
	prog.classes = make(map[string]*class)
	prog.interfaces = make(map[string]*classInterface)
	prog.enums = make(map[string]*enum)
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

func (me *program) read(hymnPackage []string, hymnFile string) (*hmfile, *parseError) {
	uid := strconv.Itoa(me.moduleUID)
	me.moduleUID++

	hymnFile, _ = filepath.Abs(hymnFile)
	name := fileName(hymnFile)

	module := me.hymnFileInit(uid, name, hymnPackage, hymnFile)

	me.modules[uid] = module
	me.hmfiles[hymnFile] = module
	me.hmorder = append(me.hmorder, module)

	if er := module.parse(); er != nil {
		return nil, er
	}

	return module, nil
}

func (me *program) compile(cc string) {
	list := me.hmorder
	for x := len(list) - 1; x >= 0; x-- {
		module := list[x]
		os.MkdirAll(module.destination, os.ModePerm)
		file := module.generateC()
		if file != "" {
			me.sources[module.path] = file
		}
	}
	if me.testing {
		me.sources["{{unit test}}"] = me.generateUnitTestsC()
	}
}

func (me *program) gcc(flags *flags, fileOut string) {
	command := flags.cc
	sources := me.sources

	if debug {
		fmt.Println("=== " + command + " ===")
	}

	paramGcc := make([]string, 0)
	if flags.analysis {
		paramGcc = append(paramGcc, "-v")
		paramGcc = append(paramGcc, "-o")
		paramGcc = append(paramGcc, flags.writeTo)
		paramGcc = append(paramGcc, command)
		command = "scan-build"
	}
	if flags.info {
		paramGcc = append(paramGcc, "-g")
	}
	if flags.sanitizeAddress {
		paramGcc = append(paramGcc, "-fsanitize=address")
	}
	if flags.optimize {
		paramGcc = append(paramGcc, "-O2")
	}
	paramGcc = append(paramGcc, "-Wall")
	paramGcc = append(paramGcc, "-Wextra")
	paramGcc = append(paramGcc, "-Werror")
	paramGcc = append(paramGcc, "-pedantic")
	paramGcc = append(paramGcc, "-std=c11")
	hmpathabs, _ := filepath.Abs(filepath.Join(flags.writeTo, "src"))
	paramGcc = append(paramGcc, "-I"+me.libc)
	paramGcc = append(paramGcc, "-I"+hmpathabs)
	for _, src := range sources {
		paramGcc = append(paramGcc, src)
	}
	paramGcc = append(paramGcc, "-o")
	if flags.library {
		fileOut += ".o"
		paramGcc = append(paramGcc, fileOut)
		paramGcc = append(paramGcc, "-c")
	} else {
		paramGcc = append(paramGcc, fileOut)
	}

	if debugCommand {
		fmt.Println(command, strings.Join(paramGcc, " "))
	}

	if flags.script {
		name := fileName(flags.path)
		var script strings.Builder
		script.WriteString("#!/bin/sh -e\n\n")
		script.WriteString(command)
		for _, line := range paramGcc {
			script.WriteString(" \\\n")
			script.WriteString(line)
		}
		script.WriteString("\n\n")
		script.WriteString("./" + fileOut)
		script.WriteString("\n")
		sh := filepath.Join(flags.writeTo, name+".sh")
		write(sh, script.String())
		cmd := exec.Command("chmod", "+x", sh)
		err := cmd.Run()
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	if flags.makefile {
		var make strings.Builder
		make.WriteString("CC= gcc\n")
		make.WriteString("program:\n\t")
		make.WriteString(command + " ")
		make.WriteString(strings.Join(paramGcc, " "))
		write(filepath.Join(flags.writeTo, "makefile"), make.String())
	}

	cmd := exec.Command(command, paramGcc...)
	stdout, err := cmd.CombinedOutput()
	std := string(stdout)
	if std != "" {
		fmt.Println(std)
	}
	if err != nil {
		fmt.Println(err)
	}
}
