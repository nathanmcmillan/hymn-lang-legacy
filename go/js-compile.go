package main

import (
	"fmt"
	"path/filepath"
)

func (me *hmfile) generateJavaScript() string {

	folder := me.out
	filename := fileName(me.path)

	if debug {
		fmt.Println("=== javascript: " + filename + " ===")
	}

	file := filepath.Join(folder, filename+".js")

	write(file, "class foo {}")

	return file
}
