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
	flags.hmlib = path.Clean(path.Join(pwd, "..", "lib"))
	fmt.Println("$LIB", flags.hmlib)
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
		flags.path = tests + "/" + info.Name()
		flags.writeTo = folder + "/out/" + nameNum
		stdout, err := execCompile(flags)
		if err != nil {
			t.Errorf("compile error for " + info.Name() + ". " + err.Error())
		}
		expected := string(read(folder + "/assert/" + nameNum + ".out"))
		if stdout != expected {
			outln := strings.Split(stdout, "\n")
			expectln := strings.Split(expected, "\n")
			min := len(outln)
			temp := len(expectln)
			if temp < min {
				min = temp
			}
			var i int
			badln := false
			for i = 0; i < min; i++ {
				if outln[i] != expectln[i] {
					badln = true
					break
				}
			}
			e := "assert failed for " + info.Name()
			if badln {
				e += " on line " + strconv.Itoa(i+1) + ". expected <[" + expectln[i] + "]> but was <[" + outln[i] + "]>"
			} else if min == temp {
				e += ". expected output was shorter than actual output."
			} else {
				e += ". actual output was shorter than expected output."
			}
			t.Errorf(e)
		}
	}
}
