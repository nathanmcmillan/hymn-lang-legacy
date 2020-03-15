package main

import (
	"fmt"
	"path/filepath"
	"strings"
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
		}
		if er := me.eat(":"); er != nil {
			return er
		}
	}

	// TODO: if len(path) == 1 then it is a local file

	pack := path[0]
	var ok bool
	var location string

	if location, ok = me.program.packages[pack]; !ok {
		return err(me, ECodeImportPath, fmt.Sprintf("Package `%s` not found. Try including it in $HYMN_PACKAGES or through the -v flag.", pack))
	}

	fmt.Println("debug import package ::", strings.Join(path, ", "))
	fmt.Println("debug import location ::", location)

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

	fmt.Println("debug import hymn file path ::", hymnFilePath)
	fmt.Println("debug program packages ::", me.program.packages)

	if !filepath.IsAbs(hymnFilePath) {
		var er error
		hymnFilePath, er = filepath.Abs(filepath.Join(module.program.directory, hymnFilePath))
		if er != nil {
			return err(me, ECodeImportPath, er.Error())
		}
	}

	alias := pack
	if me.token.is == "as" {
		if er := me.eat("as"); er != nil {
			return er
		}
		alias = me.token.value
		if er := me.eat("id"); er != nil {
			return er
		}
	}

	var importing *hmfile
	found, ok := module.program.hmfiles[hymnFilePath]
	if ok {
		if _, ok := module.importPaths[hymnFilePath]; ok {
			return err(me, ECodeDoubleModuleImport, fmt.Sprintf("Module `%s` was already imported.", hymnFilePath))
		}
		importing = found
	} else {
		outputDirectory := alias
		var fer error
		outputDirectory, fer = filepath.Abs(filepath.Join(module.program.outputDirectory, outputDirectory))
		if fer != nil {
			return err(me, ECodeImportPath, fer.Error())
		}

		var er *parseError
		fmt.Println("debug ::", outputDirectory, "::", hymnFilePath, "::", alias)
		importing, er = module.program.parse(outputDirectory, hymnFilePath, module.program.libs)
		if er != nil {
			return er
		}

		if debug {
			fmt.Println("=== parse: " + module.name + " ===")
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
		}
	}

	if er := me.eat("line"); er != nil {
		return er
	}

	return nil
}
