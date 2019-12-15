package main

import (
	"fmt"
	"os"
	"strings"
)

func initC(module *hmfile, folder, root, name, hmlibs string) *cfile {
	cfile := module.cFileInit()
	guard := module.defNameSpace(root, name)

	cfile.headIncludeSection.WriteString("#ifndef " + guard + "\n")
	cfile.headIncludeSection.WriteString("#define " + guard + "\n")

	cfile.headIncludeSection.WriteString("\n#include <stdio.h>")
	cfile.headIncludeSection.WriteString("\n#include <stdlib.h>")
	cfile.headIncludeSection.WriteString("\n#include <stdint.h>")
	cfile.headIncludeSection.WriteString("\n#include <inttypes.h>")
	cfile.headIncludeSection.WriteString("\n#include <stdbool.h>")

	requirelibs := false
	hmlibls := scan(hmlibs)
	for _, f := range hmlibls {
		name := f.Name()
		if strings.HasSuffix(name, ".c") {
			cfile.hmfile.program.sources[name] = hmlibs + "/" + name
		} else if strings.HasSuffix(name, ".h") {
			if !requirelibs {
				requirelibs = true
				cfile.headIncludeSection.WriteString("\n")
			}
			cfile.headIncludeSection.WriteString("\n#include \"" + hmlibs + "/" + name + "\"")
		}
	}

	return cfile
}

func (me *cfile) subC(root, folder, rootname, subfolder, name, hmlibs, filter string, filters map[string]string) string {
	if debug {
		fmt.Println("=== " + subfolder + "/" + name + " ===")
	}

	folder = folder + "/" + subfolder

	module := me.hmfile
	cfile := initC(module, folder, rootname, name, hmlibs)

	module.program.sources[name] = folder + "/" + name + ".c"

	fx := 0
	for f, fname := range filters {
		if f == filter {
			continue
		}
		subfolder := f[0:strings.Index(f, "<")]
		cfile.headIncludeSection.WriteString("\n#include \"" + root + "/" + subfolder + "/" + fname + ".h\"")
		fx++
	}
	if fx > 0 {
		cfile.headIncludeSection.WriteString("\n")
	}

	var code strings.Builder
	code.WriteString("#include \"" + name + ".h\"\n")

	for _, c := range module.defineOrder {
		underscore := strings.LastIndex(c, "_")
		name := c[0:underscore]
		typed := c[underscore+1:]
		if !strings.HasPrefix(name, filter) {
			if typed == "type" {
				cfile.typedefClass(module.classes[name])
			}
		} else {
			if typed == "type" {
				cfile.defineClass(module.classes[name])
			} else if typed == "enum" {
				cfile.defineEnum(module.enums[name])
			}
		}
	}

	if module.needInit {
		for _, s := range module.statics {
			cfile.declareStatic(s)
		}
	}

	for _, f := range module.functionOrder {
		if !strings.HasPrefix(f, filter) {
			continue
		}
		cfile.compileFunction(f, module.functions[f], true)
	}

	fileCode := folder + "/" + name + ".c"

	os.Mkdir(folder, os.ModePerm)

	write(fileCode, code.String())

	if len(cfile.codeFn) > 0 {
		for _, cfn := range cfile.codeFn {
			fileappend(fileCode, cfn.String())
		}
		cfile.headSuffix.WriteString("\n")
	}

	cfile.headSuffix.WriteString("\n#endif\n")
	write(folder+"/"+name+".h", cfile.head())

	return fileCode
}
