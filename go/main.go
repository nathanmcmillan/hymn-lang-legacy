package main

//
//           The Hymn Compiler
// Copyright 2019 Nathan Michael McMillan
//

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	debug        = true
	debugTokens  = false
	debugTree    = false
	debugCommand = false
)

type flags struct {
	cc              string
	path            string
	libc            string
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
	testing         bool
	doNotRun        bool
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
	flag.StringVar(&flags.libc, "d", "", "directory path of hymn libc files")
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
	flag.BoolVar(&flags.doNotRun, "r", false, "compile but do not run")
	flag.BoolVar(&flags.testing, "t", false, "assert unit tests")
	flag.Parse()

	if flags.help || flags.path == "" {
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

func parsePackages(dict map[string]string, value string) error {
	if value == "" {
		return nil
	}
	var raw interface{}
	er := json.Unmarshal([]byte(value), &raw)
	if er != nil {
		return er
	}
	table, ok := raw.(map[string]interface{})
	if !ok {
		return errors.New(value)
	}
	for k, v := range table {
		p, ok := v.(string)
		if !ok {
			return errors.New(value)
		}
		if v == "" {
			continue
		}
		path, _ := filepath.Abs(p)
		dict[k] = path
	}
	return nil
}

func execCompile(flags *flags) (string, *parseError, error) {
	var fer error
	var outsourcedir string

	outsourcedir, fer = filepath.Abs(flags.writeTo)
	if fer != nil {
		panic(fer.Error())
	}
	outsourcedir = filepath.Join(outsourcedir, "src")

	stat, er := os.Stat(flags.path)
	if er != nil {
		panic(er)
	}
	if stat.IsDir() {
		flags.path = filepath.Join(flags.path, "main.hm")
	}

	var directory string

	directory, fer = filepath.Abs(filepath.Dir(flags.path))
	if fer != nil {
		panic(fer.Error())
	}

	libc := os.Getenv("HYMN_LIBC")
	if flags.libc != "" {
		libc = flags.libc
	}
	if libc == "" {
		libc = "libc"
	}
	if !filepath.IsAbs(libc) {
		libc, _ = filepath.Abs(libc)
	}

	program := programInit()
	program.outsourcedir = outsourcedir
	program.libc = libc
	program.directory = directory
	program.testing = flags.testing

	er = parsePackages(program.packages, os.Getenv("HYMN_PACKAGES"))
	if er != nil {
		panic(er)
	}

	er = parsePackages(program.packages, flags.packages)
	if er != nil {
		panic(er)
	}

	dir := filepath.Base(program.directory)
	name := fileName(flags.path)

	program.packages[dir] = program.directory

	program.loadlibs(libc)

	hmlib := &hmlib{}
	hmlib.libs()
	program.hmlib = hmlib

	pack := []string{dir, name}

	_, perr := program.read(pack, flags.path)

	if perr != nil {
		return "", perr, nil
	}

	program.compile(flags.cc)

	fileOut := flags.writeTo + "/" + name
	if exists(fileOut) {
		os.Remove(fileOut)
	}
	if flags.doNotCompile {
		return "", nil, nil
	}
	program.gcc(flags, fileOut)
	if flags.doNotRun {
		return "", nil, nil
	}
	s, e := execBin(flags, name)
	return s, nil, e
}
