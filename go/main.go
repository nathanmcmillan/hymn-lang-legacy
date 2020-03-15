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
	debug        = true
	debugTokens  = false
	debugTree    = true
	debugCommand = false
)

type flags struct {
	cc              string
	path            string
	hmlib           string
	writeTo         string
	packages        string
	help            bool
	format          bool
	library         bool
	analysis        bool
	memoryCheck     bool
	sanitizeAddress bool
	info            bool
	optimize        bool
	makefile        bool
	script          bool
	doNotCompile    bool
}

func fmc(depth int) string {
	space := ""
	for i := 0; i < depth; i++ {
		space += "    "
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
	flag.StringVar(&flags.path, "p", "", "path to main hymn file or directory")
	flag.StringVar(&flags.hmlib, "d", "", "directory path of hmlib files")
	flag.StringVar(&flags.writeTo, "w", "out", "write generated files to this directory")
	flag.StringVar(&flags.packages, "v", "", "Set of additional package directories")
	flag.BoolVar(&flags.help, "h", false, "show usage")
	flag.BoolVar(&flags.format, "f", false, "format the given code")
	flag.BoolVar(&flags.analysis, "a", false, "run static analysis on the generated binary")
	flag.BoolVar(&flags.sanitizeAddress, "s", false, "includes memory analysis in the binary (sends -fsanitize=address to the compiler)")
	flag.BoolVar(&flags.memoryCheck, "m", false, "run dynamic memory analysis on the generated binary")
	flag.BoolVar(&flags.library, "l", false, "generate code for use as a library")
	flag.BoolVar(&flags.info, "i", false, "includes additional information in the binary (sends -g flag to the compiler)")
	flag.BoolVar(&flags.optimize, "o", false, "optimizes the binary (sends -O2 flag to the compiler)")
	flag.BoolVar(&flags.makefile, "g", false, "generate a makefile")
	flag.BoolVar(&flags.script, "b", false, "generate a shell script for compiling")
	flag.BoolVar(&flags.doNotCompile, "x", false, "do not compile")
	flag.Parse()

	if flags.help || flags.path == "" || flags.hmlib == "" {
		helpExit()
	}

	if flags.format {
		execFormat(flags.path)
	} else {
		_, parseError, fsError := execCompile(flags)
		if parseError != nil {
			fmt.Println(parseError.print())
			os.Exit(1)
		}
		if fsError != nil {
			fmt.Println(fsError.Error())
			os.Exit(1)
		}
	}
}

func execCompile(flags *flags) (string, *parseError, error) {
	var fer error
	var outputDirectory string

	outputDirectory, fer = filepath.Abs(flags.writeTo)
	if fer != nil {
		panic(fer.Error())
	}

	if stat, er := os.Stat(flags.path); er != nil {
		if stat.IsDir() {
			flags.path = filepath.Join(flags.path, "main.hm")
		}
	}

	var directory string

	directory, fer = filepath.Abs(filepath.Dir(flags.path))
	if fer != nil {
		panic(fer.Error())
	}

	program := programInit()
	program.outputDirectory = filepath.Join(outputDirectory, "src")
	program.libs = flags.hmlib
	program.directory = directory

	parsePackages(program.packages, os.Getenv("HYMN_PACKAGES"))
	parsePackages(program.packages, flags.packages)

	program.packages[filepath.Base(program.directory)] = program.directory

	program.loadlibs(flags.hmlib)

	hmlib := &hmlib{}
	hmlib.libs()
	program.hmlib = hmlib

	if _, er := program.parse(flags.writeTo, flags.path, flags.hmlib); er != nil {
		return "", er, nil
	}

	program.compile(flags.cc)

	name := fileName(flags.path)
	fileOut := flags.writeTo + "/" + name
	if exists(fileOut) {
		os.Remove(fileOut)
	}
	if flags.doNotCompile {
		return "", nil, nil
	}
	gcc(flags, program.sources, fileOut)
	s, e := execBin(flags, name)
	return s, nil, e
}

func (me *program) parse(out, path, libs string) (*hmfile, *parseError) {
	uid := strconv.Itoa(me.moduleUID)
	me.moduleUID++

	path, _ = filepath.Abs(path)
	name := fileName(path)
	module := me.hymnFileInit(uid, name, out, path, libs)

	me.modules[uid] = module
	me.hmfiles[path] = module
	me.hmorder = append(me.hmorder, module)

	if er := module.parse(out, path); er != nil {
		return nil, er
	}

	return module, nil
}

func (me *program) compile(cc string) {
	list := me.hmorder
	for x := len(list) - 1; x >= 0; x-- {
		module := list[x]
		os.MkdirAll(module.outputDirectory, os.ModePerm)
		file := module.generateC()
		if file != "" {
			me.sources[module.path] = file
		}
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
	hmpathabs, _ := filepath.Abs(filepath.Join(flags.writeTo, "src"))
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

	if debugCommand {
		fmt.Println(command, strings.Join(paramGcc, " "))
	}

	if flags.script {
		name := fileName(flags.path)
		var script strings.Builder
		script.WriteString("#!/bin/sh\n\n")
		script.WriteString(command)
		for _, line := range paramGcc {
			script.WriteString(" \\\n")
			script.WriteString(line)
		}
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

func execBin(flags *flags, name string) (string, error) {
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

func parsePackages(dict map[string]string, value string) {
	list := strings.Split(value, ":")
	for _, item := range list {
		if item == "" {
			continue
		}
		name := filepath.Base(item)
		path, _ := filepath.Abs(item)
		dict[name] = path
	}
}
