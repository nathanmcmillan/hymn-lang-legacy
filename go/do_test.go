package main

import (
	"fmt"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
)

func TestCompile(t *testing.T) {
	printStacktrace = false
	args := os.Args
	min := 0
	max := int(math.MaxUint32 >> 1)
	positive := true
	negative := true
	if len(args) > 3 {
		temp, err := strconv.Atoi(args[3])
		if err == nil {
			min = temp
		}
		if len(args) > 4 {
			temp, err := strconv.Atoi(args[4])
			if err == nil {
				max = temp
			}
			if len(args) > 5 {
				if args[5] == "--negative" {
					positive = false
				} else if args[5] == "--no-negative" {
					negative = false
				}
			}
		}
	}
	debug = false
	flags := &flags{}
	flags.cc = "gcc"
	pwd, _ := os.Getwd()
	fmt.Println("$PWD", pwd)
	flags.hmlib = path.Clean(path.Join(pwd, "..", "lib"))
	fmt.Println("$LIB", flags.hmlib)
	folder := "autotest"
	if positive {
		tests := folder + "/code"
		source := scan(tests)
		for _, info := range source {
			name := strings.TrimSuffix(info.Name(), ".hm")
			nameNum := strings.Split(name, "-")[0]
			number, err := strconv.Atoi(nameNum)
			if err != nil {
				continue
			}
			if number < min || number > max {
				continue
			}
			fmt.Println("====================================================================== test", info.Name())
			flags.path = tests + "/" + info.Name()
			flags.writeTo = tests + "/out/" + nameNum
			stdout, errp, errf := execCompile(flags)
			if errp != nil {
				t.Errorf("Parsing error for " + info.Name() + "\n" + errp.print())
			}
			if errf != nil {
				t.Errorf("Compile error for " + info.Name() + ". " + errf.Error())
			}
			reading, er := read(tests + "/assert/" + nameNum + ".out")
			if er != nil {
				panic(er)
			}
			expected := string(reading)
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
	if negative {
		tests := folder + "/negative"
		source := scan(tests)
		for _, info := range source {
			name := strings.TrimSuffix(info.Name(), ".hm")
			nameNum := strings.Split(name, "-")[0]
			number, err := strconv.Atoi(nameNum)
			if err != nil {
				continue
			}
			if number < min || number > max {
				continue
			}
			fmt.Println("====================================================================== negative test", info.Name())
			flags.path = tests + "/" + info.Name()
			flags.writeTo = tests + "/out/" + nameNum
			stdout, errp, errf := execCompile(flags)
			if errp != nil {
				errprint := errp.simple()
				reading, er := read(tests + "/assert/" + nameNum + ".out")
				if er != nil {
					panic(er)
				}
				expected := string(reading)
				if errprint != expected {
					outln := strings.Split(errprint, "\n")
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
			} else if errf != nil {
				t.Errorf("Compile error for " + info.Name() + ". " + errf.Error())
			} else {
				t.Errorf("Negative test expects a parsing error to be thrown but it completed normally:\n" + stdout)
			}
		}
	}
}
