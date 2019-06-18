package main

import (
	"fmt"
	"os"
	"os/exec"
)

const (
	ftokens = "tokens"
	ftree   = "tree"
	fcode   = "main.c"
	fapp    = "main"
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
	compile(true, "out", data)
}

func compile(debug bool, out string, data []byte) string {
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
		ptokens := out + "/" + ftokens
		if exists(ptokens) {
			os.Remove(ptokens)
		}
		create(ptokens, dump)

		fmt.Println("=== parse ===")
	}
	program := parse(tokens)
	if debug {
		dump := program.dump()
		if exists(ftree) {
			os.Remove(ftree)
		}
		ptree := out + "/" + ftree
		create(ptree, dump)
		fmt.Println("=== code ===")
	}
	makecode(out, program)
	pcode := out + "/" + fcode
	papp := out + "/" + fapp
	fmt.Println("=== gcc ===")
	if exists(papp) {
		os.Remove(papp)
	}
	stdout, err := exec.Command("gcc", pcode, "-o", papp).CombinedOutput()
	std := string(stdout)
	if std != "" {
		fmt.Println(std)
	}
	if err != nil {
		panic(err)
	}
	if exists(papp) {
		fmt.Println("=== run ===")
		stdout, _ = exec.Command(papp).CombinedOutput()
		finalout := string(stdout)
		fmt.Println(finalout)
		return finalout
	}
	fmt.Println("===")
	return ""
}
