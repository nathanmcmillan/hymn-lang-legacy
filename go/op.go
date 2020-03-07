package main

import (
	"io/ioutil"
	"os"
	"strings"
)

func create(path string) *os.File {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	return f
}

func write(path, content string) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString(content)
	if err != nil {
		panic(err)
	}
}

func fileappend(path, content string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString(content)
	if err != nil {
		panic(err)
	}
}

func scan(path string) []os.FileInfo {
	dir, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}
	return dir
}

func read(path string) ([]byte, error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func fileName(path string) string {
	slash := strings.LastIndex(path, "/")
	dot := strings.LastIndex(path, ".")
	if slash == -1 {
		if dot == -1 {
			return path
		}
		return path[0:dot]
	} else {
		if dot == -1 {
			return path[slash+1:]
		}
		return path[slash+1 : dot]
	}
}

func fileDir(path string) string {
	return path[0:strings.LastIndex(path, "/")]
}
