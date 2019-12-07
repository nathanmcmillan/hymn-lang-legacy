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
	debugTree   = true

	helpFlag     bool
	formatFlag   bool
	pathFlag     string
	hmlibFlag    string
	libraryFlag  bool
	analysisFlag bool
)

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

	flag.BoolVar(&helpFlag, "h", false, "show usage")
	flag.BoolVar(&formatFlag, "f", false, "format the given code")
	flag.BoolVar(&analysisFlag, "a", false, "run static analysis on the given code")
	flag.StringVar(&pathFlag, "p", "", "path to main hymn file")
	flag.StringVar(&hmlibFlag, "d", "", "directory path of hmlib files")
	flag.BoolVar(&libraryFlag, "lib", false, "generate code for use as a library")
	flag.Parse()

	if helpFlag || pathFlag == "" || hmlibFlag == "" {
		helpExit()
	}

	if formatFlag {
		execFormat(pathFlag)
	} else {
		execCompile("out", pathFlag, hmlibFlag, analysisFlag, libraryFlag)
	}
}

func execCompile(out, path, libDir string, isAnalysis, isLib bool) string {
	program := programInit()
	program.out = out
	program.libDir = libDir
	program.directory = fileDir(path)

	hmlib := &hmlib{}
	hmlib.libs()
	program.hmlib = hmlib

	program.compile(out, path, libDir)

	name := fileName(path)
	fileOut := out + "/" + name
	if exists(fileOut) {
		os.Remove(fileOut)
	}
	gcc(program.sources, fileOut, isAnalysis, isLib)
	return app(fileOut)
}

func (me *program) compile(out, path, libDir string) {
	name := fileName(path)
	hymn := me.hymnFileInit(name)
	me.hmfiles[name] = hymn
	me.hmorder = append(me.hmorder, hymn)
	hymn.parse(out, path)
	source := hymn.generateC(out, name, libDir)
	me.sources[name] = source
}

func gcc(sources map[string]string, fileOut string, isAnalysis, isLib bool) {
	fmt.Println("=== gcc ===")
	command := "gcc"
	paramGcc := make([]string, 0)
	if isAnalysis {
		paramGcc = append(paramGcc, "-v")
		paramGcc = append(paramGcc, "-o")
		paramGcc = append(paramGcc, "out")
		paramGcc = append(paramGcc, command)
		command = "scan-build"
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
	if isLib {
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
		panic(err)
	}
}

func app(path string) string {
	if exists(path) {
		fmt.Println("=== run ===")
		stdout, _ := exec.Command(path).CombinedOutput()
		finalout := string(stdout)
		fmt.Println(finalout)
		return finalout
	}
	fmt.Println("===")
	return ""
}
