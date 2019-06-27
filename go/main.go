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
	name := fileName(path)
	sources := make(map[string]string)

	program := compile(out, path)
	pathSource := generateC(out, name, program)

	sources[name] = pathSource

	fileOut := out + "/" + name
	if exists(fileOut) {
		os.Remove(fileOut)
	}
	gcc(sources, fileOut, isLib)
	return app(fileOut)
}

func compile(out, path string) *program {
	name := fileName(path)
	program := parse(out, path)
	if debug {
		fileTree := out + "/" + name + ".tree"
		dump := program.dump()
		if exists(fileTree) {
			os.Remove(fileTree)
		}
		create(fileTree, dump)
		fmt.Println("=== code ===")
	}
	return program
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
