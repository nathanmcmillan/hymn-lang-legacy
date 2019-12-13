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

	ccFlag              string
	pathFlag            string
	hmlibFlag           string
	writeToFlag         string
	helpFlag            bool
	formatFlag          bool
	libraryFlag         bool
	analysisFlag        bool
	memoryCheckFlag     bool
	sanitizeAddressFlag bool
	infoFlag            bool
	optimizeFlag        bool
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

	flag.StringVar(&ccFlag, "c", "gcc", "specify what compiler to use")
	flag.StringVar(&pathFlag, "p", "", "path to main hymn file")
	flag.StringVar(&hmlibFlag, "d", "", "directory path of hmlib files")
	flag.StringVar(&writeToFlag, "w", "out", "write generated files to this directory")
	flag.BoolVar(&helpFlag, "h", false, "show usage")
	flag.BoolVar(&formatFlag, "f", false, "format the given code")
	flag.BoolVar(&analysisFlag, "a", false, "run static analysis on the generated binary")
	flag.BoolVar(&sanitizeAddressFlag, "s", false, "includes memory analysis in the binary (sends -fsanitize=address to the compiler)")
	flag.BoolVar(&memoryCheckFlag, "m", false, "run dynamic memory analysis on the generated binary")
	flag.BoolVar(&libraryFlag, "l", false, "generate code for use as a library")
	flag.BoolVar(&infoFlag, "i", false, "includes additional information in the binary (sends -g flag to the compiler)")
	flag.BoolVar(&optimizeFlag, "o", false, "optimizes the binary (sends -O2 flag to the compiler)")
	flag.Parse()

	if helpFlag || pathFlag == "" || hmlibFlag == "" {
		helpExit()
	}

	if formatFlag {
		execFormat(pathFlag)
	} else {
		execCompile(writeToFlag, pathFlag, hmlibFlag)
	}
}

func execCompile(out, path, libDir string) string {

	os.MkdirAll(out, os.ModePerm)

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
	gcc(program.sources, fileOut)
	return app(out, name)
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

func gcc(sources map[string]string, fileOut string) {
	command := ccFlag
	fmt.Println("=== " + command + " ===")
	paramGcc := make([]string, 0)
	if analysisFlag {
		paramGcc = append(paramGcc, "-v")
		paramGcc = append(paramGcc, "-o")
		paramGcc = append(paramGcc, writeToFlag)
		paramGcc = append(paramGcc, command)
		command = "scan-build"
	}
	if infoFlag {
		paramGcc = append(paramGcc, "-g")
	}
	if sanitizeAddressFlag {
		paramGcc = append(paramGcc, "-fsanitize=address")
	}
	if optimizeFlag {
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
	if libraryFlag {
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

func app(folder, name string) string {
	path := folder + "/" + name
	if exists(path) {
		fmt.Println("=== run ===")
		var stdout []byte
		pwd, _ := os.Getwd()
		os.Chdir(folder)
		bwd, _ := os.Getwd()
		binary := bwd + "/" + name
		if memoryCheckFlag {
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
