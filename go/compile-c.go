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
			cfile.headIncludeSection.WriteString("\n#include \"" + name + "\"")
		}
	}

	return cfile
}

func (me *cfile) subC(root, folder, rootname, hmlibs, filter string, subc *subc, filterOrder []string, filters map[string]subc) string {
	name := subc.fname
	subfolder := subc.subfolder

	if debug {
		fmt.Println("=== " + subfolder + "/" + name + " ===")
	}

	folder = folder + "/" + subfolder

	module := me.hmfile
	cfile := initC(module, folder, rootname, name, hmlibs)

	module.program.sources[name] = folder + "/" + name + ".c"

	havefilters := false
	for _, f := range filterOrder {
		if f == filter {
			continue
		}
		if !havefilters {
			havefilters = true
			cfile.headIncludeSection.WriteString("\n")
		}
		subc := filters[f]
		cfile.headIncludeSection.WriteString("\n#include \"" + subc.subfolder + "/" + subc.fname + ".h\"")
	}

	var code strings.Builder
	code.WriteString("#include \"" + name + ".h\"\n")

	for _, c := range module.defineOrder {
		underscore := strings.LastIndex(c, "_")
		name := c[0:underscore]
		typed := c[underscore+1:]
		matching := name == filter
		if typed == "type" {
			cl := module.classes[name]
			if matching {
				cfile.defineClass(cl)
			} else {
				cfile.typedefClass(cl)
			}
		} else if typed == "enum" {
			en := module.enums[name]
			if matching {
				cfile.defineEnum(en)
			} else {
				if en.baseEnum() == en {
					cfile.typedefEnum(en)
				}
			}
		} else {
			panic("missing type")
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
