package main

//
//           The Hymn Compiler
// Copyright 2019 Nathan Michael McMillan
//

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	debug       = true
	debugTokens = false
	debugTree   = true
)

type flags struct {
	cc              string
	path            string
	hmlib           string
	writeTo         string
	help            bool
	format          bool
	library         bool
	analysis        bool
	memoryCheck     bool
	sanitizeAddress bool
	info            bool
	optimize        bool
	makefile        bool
	doNotCompile    bool
	variables       string
}

const (
	spaceChar = '\t'
	spaceFmc  = string(spaceChar)
)

func fmc(depth int) string {
	space := ""
	for i := 0; i < depth; i++ {
		space += spaceFmc
	}
	return space
}

func help() {
	fmt.Println("Hymn command line interface.")
	fmt.Println("")
	flag.Usage()
}

func helpExit() {
	help()
	os.Exit(0)
}

func main() {

	flags := &flags{}

	flag.StringVar(&flags.cc, "c", "gcc", "specify what compiler to use")
	flag.StringVar(&flags.path, "p", "", "path to main hymn file")
	flag.StringVar(&flags.hmlib, "d", "", "directory path of hmlib files")
	flag.StringVar(&flags.writeTo, "w", "out", "write generated files to this directory")
	flag.StringVar(&flags.variables, "v", "", "set of import expansion variables")
	flag.BoolVar(&flags.help, "h", false, "show usage")
	flag.BoolVar(&flags.format, "f", false, "format the given code")
	flag.BoolVar(&flags.analysis, "a", false, "run static analysis on the generated binary")
	flag.BoolVar(&flags.sanitizeAddress, "s", false, "includes memory analysis in the binary (sends -fsanitize=address to the compiler)")
	flag.BoolVar(&flags.memoryCheck, "m", false, "run dynamic memory analysis on the generated binary")
	flag.BoolVar(&flags.library, "l", false, "generate code for use as a library")
	flag.BoolVar(&flags.info, "i", false, "includes additional information in the binary (sends -g flag to the compiler)")
	flag.BoolVar(&flags.optimize, "o", false, "optimizes the binary (sends -O2 flag to the compiler)")
	flag.BoolVar(&flags.makefile, "g", false, "generate a makefile")
	flag.BoolVar(&flags.doNotCompile, "x", false, "do not compile")
	flag.Parse()

	if flags.help || flags.path == "" || flags.hmlib == "" {
		helpExit()
	}

	if flags.format {
		execFormat(flags.path)
	} else {
		execCompile(flags)
	}
}

func execCompile(flags *flags) (string, error) {
	out, err := filepath.Abs(flags.writeTo)
	if err != nil {
		panic(err.Error())
	}

	program := programInit()
	program.out = out
	program.libs = flags.hmlib
	program.directory = fileDir(flags.path)

	variableFlags(program.shellvar, os.Getenv("HYMN_MODULES"))
	variableFlags(program.shellvar, flags.variables)

	program.loadlibs(flags.hmlib)

	hmlib := &hmlib{}
	hmlib.libs()
	program.hmlib = hmlib

	program.parse(flags.writeTo, flags.path, flags.hmlib)
	program.compile()

	name := fileName(flags.path)
	fileOut := flags.writeTo + "/" + name
	if exists(fileOut) {
		os.Remove(fileOut)
	}
	if flags.doNotCompile {
		return "", nil
	}
	gcc(flags, program.sources, fileOut)
	return app(flags, name)
}

func (me *program) parse(out, path, libs string) *hmfile {
	uid := strconv.Itoa(me.moduleUID)
	me.moduleUID++

	path, _ = filepath.Abs(path)
	name := fileName(path)
	module := me.hymnFileInit(uid, name, out, path, libs)

	me.modules[uid] = module
	me.hmfiles[path] = module
	me.hmorder = append(me.hmorder, module)

	module.parse(out, path)
	return module
}

func (me *program) compile() {
	list := me.hmorder
	for x := len(list) - 1; x >= 0; x-- {
		module := list[x]
		os.MkdirAll(module.out, os.ModePerm)
		source := module.generateC(module)
		me.sources[module.path] = source
	}
}

func gcc(flags *flags, sources map[string]string, fileOut string) {
	command := flags.cc
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
	hmlibabs, _ := filepath.Abs(flags.hmlib)
	hmpathabs, _ := filepath.Abs(flags.writeTo)
	paramGcc = append(paramGcc, "-I"+hmlibabs)
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
	if debug {
		fmt.Println(command, strings.Join(paramGcc, " "))
	}

	if flags.makefile {
		var make strings.Builder
		make.WriteString("CC= gcc\n")
		make.WriteString("program:\n\t")
		make.WriteString(command + " ")
		make.WriteString(strings.Join(paramGcc, " "))
		write(flags.writeTo+"/makefile", make.String())
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

func app(flags *flags, name string) (string, error) {
	path := flags.writeTo + "/" + name
	if exists(path) {
		if debug {
			fmt.Println("=== run ===")
		}
		var stdout []byte
		var err error
		pwd, _ := os.Getwd()
		os.Chdir(flags.writeTo)
		bwd, _ := os.Getwd()
		binary := bwd + "/" + name
		if flags.memoryCheck {
			stdout, err = exec.Command("valgrind", "--track-origins=yes", binary).CombinedOutput()
		} else {
			stdout, err = exec.Command(binary).CombinedOutput()
		}
		os.Chdir(pwd)
		finalout := string(stdout)
		fmt.Println(finalout)
		return finalout, err
	}
	fmt.Println("===")
	return "", nil
}

func variableFlags(dict map[string]string, value string) {
	list := strings.Split(value, ":")
	for _, item := range list {
		eq := strings.Index(item, "=")
		if eq <= 0 {
			continue
		}
		key := item[0:eq]
		is := item[eq+1:]
		if key == "" || is == "" {
			continue
		}
		dict[key] = is
	}
}
