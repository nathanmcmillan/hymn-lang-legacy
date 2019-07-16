package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestCompile(t *testing.T) {
	debug = false
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
		os.MkdirAll(out, os.ModePerm)
		stdout := linker(out, path, false)
		expected := string(read(folder + "/assert/" + nameNum + ".out"))
		if stdout != expected {
			t.Errorf("assert failed for " + info.Name())
		}
	}
}
