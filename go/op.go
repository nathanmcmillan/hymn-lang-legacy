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

func read(path string) []byte {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	contents, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	return contents
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func fileName(path string) string {
	return path[strings.LastIndex(path, "/")+1 : strings.LastIndex(path, ".")]
}

func fileDir(path string) string {
	return path[0:strings.LastIndex(path, "/")]
}
