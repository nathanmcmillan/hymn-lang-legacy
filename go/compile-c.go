package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (me *cfile) subC(root, folder, rootname, hmlibs, filter string, name string) {

	if debug {
		fmt.Println("=== compile: " + name + " ===")
	}

	module := me.hmfile

	guard := module.headerFileGuard(rootname, name)

	cfile := module.cFileInit(guard)

	cfile.pathLocal = name
	cfile.pathGlobal, _ = filepath.Rel(module.program.out, filepath.Join(folder, cfile.pathLocal))

	for _, def := range module.defineOrder {
		if def.class != nil {
			if def.class.name == filter {
				cfile.defineClass(def.class)
			}
		} else if def.enum != nil {
			if def.enum.name == filter {
				cfile.defineEnum(def.enum)
			}
		} else {
			panic("Missing definition")
		}
	}

	for _, f := range module.functionOrder {
		fun := module.functions[f]
		if fun.forClass == nil || fun.forClass.pathLocal != name {
			continue
		}
		cfile.compileFunction(f, fun, true)
	}

	if debug {
		fmt.Println("=== end: " + name + " ===")
	}

	if len(cfile.codeFn) > 0 {
		cFile := filepath.Join(folder, name+".c")

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
	write(filepath.Join(folder, name+".h"), cfile.head())
}
