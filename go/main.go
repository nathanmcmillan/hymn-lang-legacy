package main

import (
	"fmt"
	"os"
	"os/exec"
)

var (
	debug = true
)

func fmc(depth int) string {
	space := ""
	for i := 0; i < depth; i++ {
		space += "  "
	}
	return space
}

func main() {
	args := os.Args
	size := len(args)
	if size < 2 {
		fmt.Println("path?")
		return
	}
	path := args[1]
	isLib := false
	if size >= 3 {
		if args[2] == "--lib" {
			isLib = true
		}
	}
	linker("out", path, isLib)
}

func linker(out, path string, isLib bool) string {
	prog := programInit()
	prog.out = out
	prog.directory = fileDir(path)

	prog.compile(out, path)

	name := fileName(path)
	fileOut := out + "/" + name
	if exists(fileOut) {
		os.Remove(fileOut)
	}
	gcc(prog.sources, fileOut, isLib)
	return app(fileOut)
}

func (me *program) compile(out, path string) {
	name := fileName(path)

	hymn := me.hymnFileInit(name)

	me.hmfiles[name] = hymn

	hymn.parse(out, path)
	if debug {
		fileTree := out + "/" + name + ".tree"
		dump := hymn.dump()
		if exists(fileTree) {
			os.Remove(fileTree)
		}
		create(fileTree, dump)
		fmt.Println("=== generate C ===")
	}

	source := hymn.generateC(out, name)
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
	stdout, err := exec.Command("gcc", paramGcc...).CombinedOutput()
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
