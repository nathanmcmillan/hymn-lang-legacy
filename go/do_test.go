package main

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
)

func TestCompile(t *testing.T) {
	debug = false
	flags := &flags{}
	flags.cc = "gcc"
	pwd, _ := os.Getwd()
	fmt.Println("$PWD", pwd)
	libDir := path.Clean(path.Join(pwd, "..", "lib"))
	fmt.Println("$LIB", libDir)
	folder := "autotest"
	tests := folder + "/code"
	source := scan(tests)
	for _, info := range source {
		name := strings.TrimSuffix(info.Name(), ".hm")
		nameNum := strings.Split(name, "-")[0]
		_, err := strconv.Atoi(nameNum)
		if err != nil {
			continue
		}
		fmt.Println("====================================================================== test", info.Name())
		path := tests + "/" + info.Name()
		out := folder + "/out/" + nameNum
		stdout := execCompile(flags, out, path, libDir)
		expected := string(read(folder + "/assert/" + nameNum + ".out"))
		if stdout != expected {
			t.Errorf("assert failed for " + info.Name())
		}
	}
}
