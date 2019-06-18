package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestCompile(t *testing.T) {
	folder := "autotest"
	tests := folder + "/code"
	source := scan(tests)
	for _, info := range source {
		fmt.Println(info.Name())
		data := read(tests + "/" + info.Name())
		name := strings.TrimSuffix(info.Name(), ".ss")
		files := folder + "/out/" + name
		os.MkdirAll(files, os.ModePerm)
		out := compile(false, files, data)
		expected := string(read(folder + "/assert/" + name + ".out"))
		if out != expected {
			t.Errorf("assert failed for " + info.Name())
		}
	}
}
