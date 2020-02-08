package main

import (
	"fmt"
	"strings"
)

func (me *cfile) subC(root, folder, rootname, hmlibs, filter string, subc *subc, filterOrder []string, filters map[string]subc) {
	name := subc.fname

	if debug {
		fmt.Println("=== " + subc.location() + " ===")
	}

	module := me.hmfile

	cfile := module.cFileInit()
	guard := module.headerFileGuard(rootname, name)

	cfile.headStdIncludeSection.WriteString("#ifndef " + guard + "\n")
	cfile.headStdIncludeSection.WriteString("#define " + guard)

	cfile.location = subc.location()

	for _, c := range module.defineOrder {
		underscore := strings.LastIndex(c, "_")
		name := c[0:underscore]
		typed := c[underscore+1:]
		matching := name == filter
		if typed == "type" {
			if matching {
				cl := module.classes[name]
				cfile.defineClass(cl)
			}
		} else if typed == "enum" {
			if matching {
				en := module.enums[name]
				cfile.defineEnum(en)
			}
		} else {
			panic("missing type")
		}
	}

	for _, f := range module.functionOrder {
		fun := module.functions[f]
		if fun.forClass == nil || fun.forClass.location != name {
			continue
		}
		cfile.compileFunction(f, fun, true)
	}

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
