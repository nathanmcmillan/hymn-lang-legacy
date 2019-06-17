package main

import (
	"testing"
	"fmt"
)

func TestCompile(t *testing.T) {
	path := "autotest"
	dir := scan(path)
	for _, info := range dir {
		fmt.Println(info.Name())
		data := read(path + "/" + info.Name())
		out := compiler("out", data)
		fmt.Println(out)
		t.Errorf("failed to compile %s", "bad argument")
	}
}
