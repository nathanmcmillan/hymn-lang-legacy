package main

import (
	"fmt"
	"os"
	"strings"
)

func (me *cfile) subC(root, folder, rootname, hmlibs, filter string, subc *subc, filterOrder []string, filters map[string]subc) {
	name := subc.fname
	// subfolder := subc.subfolder

	if debug {
		fmt.Println("=== " + subc.location() + " ===")
	}

	// folder = folder + "/" + subfolder

	module := me.hmfile

	cfile := module.cFileInit()
	guard := module.defNameSpace(rootname, name)

	cfile.headStdIncludeSection.WriteString("#ifndef " + guard + "\n")
	cfile.headStdIncludeSection.WriteString("#define " + guard)

	// cfile.headStdIncludeSection.WriteString("\n#include <stdio.h>")
	// cfile.headStdIncludeSection.WriteString("\n#include <stdlib.h>")
	// cfile.headStdIncludeSection.WriteString("\n#include <stdint.h>")
	// cfile.headStdIncludeSection.WriteString("\n#include <inttypes.h>")
	// cfile.headStdIncludeSection.WriteString("\n#include <stdbool.h>")

	// for _, f := range filterOrder {
	// 	if f == filter {
	// 		continue
	// 	}
	// 	subc := filters[f]
	// 	cfile.headReqIncludeSection.WriteString("\n#include \"" + subc.subfolder + "/" + subc.fname + ".h\"")
	// }

	cfile.location = subc.location()

	for _, c := range module.defineOrder {
		underscore := strings.LastIndex(c, "_")
		name := c[0:underscore]
		typed := c[underscore+1:]
		matching := name == filter
		if typed == "type" {
			cl := module.classes[name]
			if matching {
				cfile.defineClass(cl)
			}
			//  else {
			// 	cfile.typedefClass(cl)
			// }
		} else if typed == "enum" {
			en := module.enums[name]
			if matching {
				cfile.defineEnum(en)
			}
			//  else {
			// 	if en.baseEnum() == en {
			// 		cfile.typedefEnum(en)
			// 	}
			// }
		} else {
			panic("missing type")
		}
	}

	for _, f := range module.functionOrder {
		fun := module.functions[f]
		if fun.forClass == nil || fun.forClass.getLocation() != name {
			continue
		}
		cfile.compileFunction(f, fun, true)
	}

	os.Mkdir(folder, os.ModePerm)

	if len(cfile.codeFn) > 0 {
		cFile := folder + "/" + name + ".c"

		module.program.sources[name] = cFile

		var code strings.Builder
		code.WriteString("#include \"" + name + ".h\"\n")
		write(cFile, code.String())

		for _, cfn := range cfile.codeFn {
			fileappend(cFile, cfn.String())
		}

		cfile.headSuffix.WriteString("\n")
	}

	cfile.headSuffix.WriteString("\n#endif\n")
	write(folder+"/"+name+".h", cfile.head())
}
