package main

import (
	"fmt"
	"path/filepath"
)

func (me *parser) importing() *parseError {
	if er := me.eat("import"); er != nil {
		return er
	}

	path := make([]string, 0)
	for {
		value := me.token.value
		if er := me.eat("id"); er != nil {
			return er
		}
		path = append(path, value)
		if me.token.is == "line" || me.token.is == "(" || me.token.is == "as" {
			break
		} else if er := me.eat(":"); er != nil {
			return er
		}
	}

	if len(path) == 1 {
		name := path[0]
		pack := me.hmfile.pack
		path = []string{}
		path = append(path, pack[0:len(pack)-1]...)
		path = append(path, name)
	}

	var ok bool
	var location string

	if location, ok = me.program.packages[path[0]]; !ok {
		return err(me, ECodeImportPath, fmt.Sprintf("Package `%s` not found. Try including it in $HYMN_PACKAGES or through the -v flag.", path[0]))
	}

	statics := make([]string, 0)
	if me.token.is == "(" {
		if er := me.eat("("); er != nil {
			return er
		}
		if me.token.is == "line" {
			if er := me.eat("line"); er != nil {
				return er
			}
		}
		for me.token.is != ")" {
			value := me.token.value
			if er := me.eat("id"); er != nil {
				return er
			}
			statics = append(statics, value)
			if me.token.is == "line" {
				if er := me.eat("line"); er != nil {
					return er
				}
			} else if me.token.is == "," {
				if er := me.eat(","); er != nil {
					return er
				}
			}
		}
		if er := me.eat(")"); er != nil {
			return er
		}
	}

	module := me.hmfile

	hymnFilePath := filepath.Join(path[1:]...) + ".hm"
	hymnFilePath = filepath.Join(location, hymnFilePath)

	if !filepath.IsAbs(hymnFilePath) {
		var er error
		hymnFilePath, er = filepath.Abs(filepath.Join(module.program.directory, hymnFilePath))
		if er != nil {
			return err(me, ECodeImportPath, er.Error())
		}
	}

	alias := path[len(path)-1]
	if me.token.is == "as" {
		if er := me.eat("as"); er != nil {
			return er
		}
		alias = me.token.value
		if er := me.eat("id"); er != nil {
			return er
		}
	}

	module.imports[alias] = nil
	module.importPaths[hymnFilePath] = nil

	var importing *hmfile
	found, ok := module.program.hmfiles[hymnFilePath]
	if ok {
		if m, _ := module.importPaths[hymnFilePath]; m != nil {
			return err(me, ECodeDoubleModuleImport, fmt.Sprintf("Module `%s` was already imported.", hymnFilePath))
		}
		importing = found

		if me.isCyclical(module, importing) {
			return err(me, ECodeImportPath, fmt.Sprintf("Cyclical dependency between `%s` and `%s`", module.path, hymnFilePath))
		}

	} else {
		var er *parseError
		importing, er = module.program.read(path, hymnFilePath)
		if er != nil {
			return er
		}

		if debug {
			fmt.Println("continue>", module.name)
		}
	}

	module.imports[alias] = importing
	module.importPaths[hymnFilePath] = importing
	module.importOrder = append(module.importOrder, alias)
	importing.crossref[module] = alias

	for _, s := range statics {
		if cl, ok := importing.classes[s]; ok {
			if _, ok := module.types[cl.name]; ok {
				return err(me, ECodeDoubleClassImport, "Cannot import class \""+cl.name+"\". It is already defined.")
			}
			module.classes[cl.name] = cl
			module.namespace[cl.name] = "class"
			module.types[cl.name] = "class"

			module.classes[cl.uid()] = cl
			module.namespace[cl.uid()] = "class"
			module.types[cl.uid()] = "class"

		} else if in, ok := importing.interfaces[s]; ok {
			if _, ok := module.types[in.name]; ok {
				return err(me, ECodeDoubleInterfaceImport, "Cannot import interface \""+in.name+"\". It is already defined.")
			}
			module.interfaces[in.name] = in
			module.namespace[in.name] = "interface"
			module.types[in.name] = "interface"

			module.interfaces[in.uid()] = in
			module.namespace[in.uid()] = "interface"
			module.types[in.uid()] = "interface"

		} else if en, ok := importing.enums[s]; ok {
			if _, ok := module.types[en.name]; ok {
				return err(me, ECodeDoubleEnumImport, "Cannot import enum \""+en.name+"\". It is already defined.")
			}
			module.enums[en.name] = en
			module.namespace[en.name] = "enum"
			module.types[en.name] = "enum"

			module.enums[en.uid()] = en
			module.namespace[en.uid()] = "enum"
			module.types[en.uid()] = "enum"

		} else if fn, ok := importing.functions[s]; ok {
			name := fn.getname()
			if _, ok := module.types[name]; ok {
				return err(me, ECodeDoubleFunctionImport, "Cannot import function \""+name+"\". It is already defined.")
			}
			module.functions[name] = fn
			module.namespace[name] = "function"
			module.types[name] = "function"

		} else if st, ok := importing.staticScope[s]; ok {
			if _, ok := module.types[st.v.name]; ok {
				return err(me, ECodeDoubleStaticVariableImport, "Cannot import variable \""+st.v.name+"\". It is already defined.")
			}
			module.staticScope[st.v.name] = st
			module.scope.variables[st.v.name] = st.v
		} else {
			return err(me, ECodeImportPath, fmt.Sprintf("I could not find the static type `%s` importing from `%s`", s, alias))
		}
	}

	if er := me.eat("line"); er != nil {
		return er
	}

	return nil
}

func (me *parser) isCyclical(module *hmfile, importing *hmfile) bool {
	for path := range importing.importPaths {
		if path == module.path {
			return true
		}
	}
	return false
}
