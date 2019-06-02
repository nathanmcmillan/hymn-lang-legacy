package main

import (
	"fmt"
	"os"
	"os/exec"
)

const (
	ftree = "out/tree"
	fcode = "out/main.c"
	fapp  = "out/main"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("path?")
		return
	}
	path := os.Args[1]
	build(path)
}

func build(path string) {
	data := read(path)
	stream := newStream(data)
	fmt.Println("=== content ===")
	fmt.Println(string(data))
	fmt.Println("=== tokens ===")
	tokens := tokenize(stream)
	fmt.Println("=== parse ===")
	if exists(ftree) {
		os.Remove(ftree)
	}
	tree := parse(tokens)
	dump := tree.dump()
	fmt.Println(dump)
	if exists(ftree) {
		os.Remove(ftree)
	}
	create(ftree, dump)
	fmt.Println("=== code ===")
	code := compile()
	if exists(fcode) {
		os.Remove(fcode)
	}
	fmt.Println(code)
	fmt.Println("=== run ===")
	if exists(fapp) {
		os.Remove(fapp)
	}
	if exists(fcode) {
		out, _ := exec.Command("gcc " + fcode + " -o " + fapp).CombinedOutput()
		fmt.Println(string(out))
		if exists(fapp) {
			out, _ = exec.Command(fapp).CombinedOutput()
			fmt.Println(string(out))
		}
	}
	fmt.Println("===")
}
