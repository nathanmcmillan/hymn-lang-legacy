package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestCompile(t *testing.T) {
	debug = false
	folder := "autotest"
	tests := folder + "/code"
	source := scan(tests)
	for _, info := range source {
		fmt.Println("====================================================================== test", info.Name())
		path := tests + "/" + info.Name()
		name := strings.TrimSuffix(info.Name(), ".hm")
		out := folder + "/out/" + name
		os.MkdirAll(out, os.ModePerm)
		stdout := linker(out, path, false)
		expected := string(read(folder + "/assert/" + name + ".out"))
		if stdout != expected {
			t.Errorf("assert failed for " + info.Name())
		}
	}
}
