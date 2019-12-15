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
	"strings"
)

var (
	debug       = true
	debugTokens = false
	debugTree   = false
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
	flag.BoolVar(&flags.help, "h", false, "show usage")
	flag.BoolVar(&flags.format, "f", false, "format the given code")
	flag.BoolVar(&flags.analysis, "a", false, "run static analysis on the generated binary")
	flag.BoolVar(&flags.sanitizeAddress, "s", false, "includes memory analysis in the binary (sends -fsanitize=address to the compiler)")
	flag.BoolVar(&flags.memoryCheck, "m", false, "run dynamic memory analysis on the generated binary")
	flag.BoolVar(&flags.library, "l", false, "generate code for use as a library")
	flag.BoolVar(&flags.info, "i", false, "includes additional information in the binary (sends -g flag to the compiler)")
	flag.BoolVar(&flags.optimize, "o", false, "optimizes the binary (sends -O2 flag to the compiler)")
	flag.Parse()

	if flags.help || flags.path == "" || flags.hmlib == "" {
		helpExit()
	}

	if flags.format {
		execFormat(flags.path)
	} else {
		execCompile(flags, flags.writeTo, flags.path, flags.hmlib)
	}
}

func execCompile(flags *flags, out, path, libs string) string {
	program := programInit()
	program.out = out
	program.libs = libs
	program.directory = fileDir(path)

	hmlib := &hmlib{}
	hmlib.libs()
	program.hmlib = hmlib

	program.parse(out, path, libs)
	program.compile()

	name := fileName(path)
	fileOut := out + "/" + name
	if exists(fileOut) {
		os.Remove(fileOut)
	}
	gcc(flags, program.sources, fileOut)
	return app(flags, out, name)
}

func (me *program) parse(out, path, libs string) {
	name := fileName(path)
	hymn := me.hymnFileInit(name)
	hymn.out = out
	hymn.path = path
	hymn.libs = libs
	me.hmfiles[name] = hymn
	me.hmorder = append(me.hmorder, hymn)
	hymn.parse(out, path)
}

func (me *program) compile() {
	for _, module := range me.hmorder {
		os.MkdirAll(module.out, os.ModePerm)
		source := module.generateC(module.out, module.path, module.libs)
		me.sources[module.name] = source
	}
}

func gcc(flags *flags, sources map[string]string, fileOut string) {
	command := flags.cc
	fmt.Println("=== " + command + " ===")
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
	fmt.Println(command, strings.Join(paramGcc, " "))
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

func app(flags *flags, folder, name string) string {
	path := folder + "/" + name
	if exists(path) {
		fmt.Println("=== run ===")
		var stdout []byte
		pwd, _ := os.Getwd()
		os.Chdir(folder)
		bwd, _ := os.Getwd()
		binary := bwd + "/" + name
		if flags.memoryCheck {
			stdout, _ = exec.Command("valgrind", "--track-origins=yes", binary).CombinedOutput()
		} else {
			stdout, _ = exec.Command(binary).CombinedOutput()
		}
		os.Chdir(pwd)
		finalout := string(stdout)
		fmt.Println(finalout)
		return finalout
	}
	fmt.Println("===")
	return ""
}
