package main

import (
	"fmt"
	"os"
)

func main() {
	path := os.Args[1]
	compile(path)
}

func compile(path string) {
	data := read(path)
	stream := newStream(data)
	fmt.Println("===")
	fmt.Println(string(data))
	fmt.Println("===")
	tokens := tokenize(stream)
	fmt.Println("===")
	instructions := parse(tokens)
	fmt.Println(instructions)
	fmt.Println("===")
}
