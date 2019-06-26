package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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
	data := read(path)
	name := path[strings.LastIndex(path, "/")+1 : strings.LastIndex(path, ".")]
	fmt.Println("file name:", name)
	compile(true, "out", name, data)
}

func compile(debug bool, out, name string, data []byte) string {
	stream := newStream(data)
	if debug {
		fmt.Println("=== content ===")
		fmt.Println(string(data))
		fmt.Println("=== tokens ===")
	}
	tokens := tokenize(stream)
	if debug {
		dump := ""
		for _, token := range tokens {
			dump += token.string() + "\n"
		}
		fileTokens := out + "/" + name + ".tokens"
		if exists(fileTokens) {
			os.Remove(fileTokens)
		}
		create(fileTokens, dump)

		fmt.Println("=== parse ===")
	}

	program := parse(tokens)
	if debug {
		fileTree := out + "/" + name + ".tree"
		dump := program.dump()
		if exists(fileTree) {
			os.Remove(fileTree)
		}
		create(fileTree, dump)
		fmt.Println("=== code ===")
	}

	fmt.Println("=== gcc ===")
	fileCode := makecode(out, name, program)
	fileOut := out + "/" + name
	if exists(fileOut) {
		os.Remove(fileOut)
	}
	stdout, err := exec.Command("gcc", fileCode, "-o", fileOut).CombinedOutput()
	std := string(stdout)
	if std != "" {
		fmt.Println(std)
	}
	if err != nil {
		panic(err)
	}
	if exists(fileOut) {
		fmt.Println("=== run ===")
		stdout, _ = exec.Command(fileOut).CombinedOutput()
		finalout := string(stdout)
		fmt.Println(finalout)
		return finalout
	}
	fmt.Println("===")
	return ""
}
