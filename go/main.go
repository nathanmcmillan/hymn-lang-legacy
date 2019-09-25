package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	debug = true
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

func main() {
	args := os.Args
	size := len(args)
	if size < 3 {
		fmt.Println("lib? path?")
		return
	}
	libDir := args[1]
	path := args[2]
	if size >= 4 && args[3] == "--fmt" {
		hymnFmt(path)
		return
	}
	isLib := false
	if size >= 4 {
		if args[3] == "--lib" {
			isLib = true
		}
	}
	linker("out", path, libDir, isLib)
}

func linker(out, path, libDir string, isLib bool) string {
	prog := programInit()
	prog.out = out
	prog.libDir = libDir
	prog.directory = fileDir(path)

	prog.compile(out, path, libDir)

	name := fileName(path)
	fileOut := out + "/" + name
	if exists(fileOut) {
		os.Remove(fileOut)
	}
	gcc(prog.sources, fileOut, isLib)
	return app(fileOut)
}

func (me *program) compile(out, path, libDir string) {
	name := fileName(path)
	hymn := me.hymnFileInit(name)
	me.hmfiles[name] = hymn
	hymn.parse(out, path)
	source := hymn.generateC(out, name, libDir)
	me.sources[name] = source
}

func gcc(sources map[string]string, fileOut string, isLib bool) {
	fmt.Println("=== gcc ===")
	paramGcc := make([]string, 0)
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
	fmt.Println("gcc", strings.Join(paramGcc, " "))
	cmd := exec.Command("gcc", paramGcc...)
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
