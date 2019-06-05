package main

import (
	"fmt"
	"os"
	"os/exec"
)

const (
	ftokens = "out/tokens"
	ftree   = "out/tree"
	fcode   = "out/main.c"
	fapp    = "out/main"
)

func fmc(depth int) string {
	space := ""
	for i := 0; i < depth; i++ {
		space += "  "
	}
	return space
}

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
	if exists(ftokens) {
		os.Remove(ftokens)
	}
	dump := ""
	for _, token := range tokens {
		dump += token.string() + "\n"
	}
	create(ftokens, dump)
	fmt.Println("=== parse ===")
	if exists(ftree) {
		os.Remove(ftree)
	}
	program := parse(tokens)
	dump = program.dump()
	if exists(ftree) {
		os.Remove(ftree)
	}
	create(ftree, dump)
	fmt.Println("=== code ===")
	code := compile(program)
	if exists(fcode) {
		os.Remove(fcode)
	}
	fmt.Println(code)
	create(fcode, code)
	fmt.Println("=== gcc ===")
	if exists(fapp) {
		os.Remove(fapp)
	}
	out, err := exec.Command("gcc", fcode, "-o", fapp).CombinedOutput()
	std := string(out)
	if std != "" {
		fmt.Println(std)
	}
	if err != nil {
		panic(err)
	}
	if exists(fapp) {
		fmt.Println("=== run ===")
		out, _ = exec.Command(fapp).CombinedOutput()
		fmt.Println(string(out))
	}
	fmt.Println("===")
}
