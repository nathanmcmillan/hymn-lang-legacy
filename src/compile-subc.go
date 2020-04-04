package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (me *cfile) subC(destination, rootname, hmlibs, filter string, name string) {

	if debug {
		fmt.Println("subcompile>", name)
	}

	module := me.hmfile

	guard := module.headerFileGuard(me.hmfile.pack, name)

	cfile := module.cFileInit(guard)

	cfile.pathLocal = name
	cfile.pathGlobal, _ = filepath.Rel(module.program.outsourcedir, filepath.Join(destination, cfile.pathLocal))

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
			panic(cfile.fail(nil) + "Missing definition")
		}
	}

	for _, f := range module.functionOrder {
		fun := module.functions[f]
		if fun.forClass == nil || fun.forClass.pathLocal != name {
			continue
		}
		cfile.compileFunction(f, fun, true)
	}

	if len(cfile.codeFn) > 0 {
		fileOut := filepath.Join(destination, name+".c")

		module.program.sources[name] = fileOut

		var code strings.Builder
		code.WriteString("#include \"" + name + ".h\"\n")
		write(fileOut, code.String())

		for _, cfn := range cfile.codeFn {
			fileappend(fileOut, cfn.String())
		}

		cfile.headSuffix.WriteString("\n")
	}

	cfile.headSuffix.WriteString("#endif\n")
	write(filepath.Join(destination, name+".h"), cfile.head())
}
