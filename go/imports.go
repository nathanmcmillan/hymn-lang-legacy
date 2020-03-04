package main

import (
	"fmt"
	"path/filepath"
)

func (me *parser) importing() *parseError {
	if er := me.eat("import"); er != nil {
		return er
	}

	value := me.token.value
	if er := me.eat(TokenStringLiteral); er != nil {
		return er
	}

	value = variableSubstitution(value, me.hmfile.program.shellvar)
	absolute, er := filepath.Abs(value)
	if er != nil {
		return err(me, ECodeImportPath, "Failed to parse import \""+value+"\". "+er.Error())
	}

	alias := filepath.Base(absolute)
	if me.token.is == "as" {
		if er := me.eat("as"); er != nil {
			return er
		}
		alias = me.token.value
		if er := me.eat("id"); er != nil {
			return er
		}
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

	path, er := filepath.Abs(filepath.Join(module.program.directory, value+".hm"))
	if er != nil {
		return err(me, ECodeImportPath, "Failed to parse import \""+value+"\". "+er.Error())
	}

	fmt.Println("debug 1 ::", value)
	fmt.Println("debug 2 ::", absolute) // this is wrong
	fmt.Println("debug 3 ::", path)     // this is also wrong

	var importing *hmfile
	found, ok := module.program.hmfiles[path]
	if ok {
		if _, ok := module.importPaths[path]; ok {
			return err(me, ECodeDoubleModuleImport, "Module \""+path+"\" was already imported.")
		}
		importing = found
	} else {
		out, fer := filepath.Abs(filepath.Join(module.program.out, value))
		if fer != nil {
			return err(me, ECodeImportPath, "Failed to parse import \""+value+"\". "+er.Error())
		}

		var er *parseError
		importing, er = module.program.parse(out, path, module.program.libs)
		if er != nil {
			return er
		}

		if debug {
			fmt.Println("=== parse: " + module.name + " ===")
		}
	}

	module.imports[alias] = importing
	module.importPaths[path] = importing
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
