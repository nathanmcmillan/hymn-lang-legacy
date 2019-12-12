package main

import (
	"fmt"
	"os"
	"strings"
)

func initC(module *hmfile, folder, name, hmlibs string) *cfile {
	cfile := module.cFileInit()
	guard := module.defNameSpace(name)

	cfile.headPrefix.WriteString("#ifndef " + guard + "\n")
	cfile.headPrefix.WriteString("#define " + guard + "\n\n")

	cfile.headIncludeSection.WriteString("#include <stdio.h>\n")
	cfile.headIncludeSection.WriteString("#include <stdlib.h>\n")
	cfile.headIncludeSection.WriteString("#include <stdint.h>\n")
	cfile.headIncludeSection.WriteString("#include <inttypes.h>\n")
	cfile.headIncludeSection.WriteString("#include <stdbool.h>\n")

	hmlibls := scan(hmlibs)
	for _, f := range hmlibls {
		name := f.Name()
		if strings.HasSuffix(name, ".c") {
			cfile.hmfile.program.sources[name] = hmlibs + "/" + name
		} else if strings.HasSuffix(name, ".h") {
			cfile.headIncludeSection.WriteString("#include \"" + hmlibs + "/" + name + "\"\n")
		}
	}

	return cfile
}

func (me *cfile) functionC(root, folder, name, hmlibs string, funcs []string) string {
	if debug {
		fmt.Println("=== generate C " + folder + "/" + name + " ===")
	}

	module := me.hmfile
	cfile := initC(module, folder, name, hmlibs)

	cfile.headIncludeSection.WriteString("#include \"" + root + "\"\n")
	cfile.headIncludeSection.WriteString("\n")

	var code strings.Builder
	code.WriteString("#include \"" + name + ".h\"\n\n")

	for _, f := range funcs {
		cfile.compileFunction(f, module.functions[f], true)
	}

	fmt.Println("=== end C ===")

	fileCode := folder + "/" + name + ".c"

	os.Mkdir(folder, os.ModePerm)

	write(fileCode, code.String())
	for _, cfn := range cfile.codeFn {
		fileappend(fileCode, cfn.String())
	}

	cfile.headSuffix.WriteString("\n#endif\n")
	write(folder+"/"+name+".h", cfile.head())

	return fileCode
}
