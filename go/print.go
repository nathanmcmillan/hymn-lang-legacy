package main

import (
	"strconv"
	"strings"
)

func (me *variable) string() string {
	return "{typed:" + me.typed + ", name:" + me.name + ", mutable:" + strconv.FormatBool(me.mutable) + "}"
}

func (me *node) string(lv int) string {
	s := ""
	s += fmc(lv) + "{is:" + me.is
	if me.value != "" {
		s += ", value:" + me.value
	}
	if me.typed != "" {
		s += ", typed:" + me.typed
	}
	if len(me.attributes) > 0 {
		s += ", attributes["
		for ix, has := range me.attributes {
			if ix > 0 {
				s += ", "
			}
			s += has
		}
		s += "]"
	}
	if len(me.has) > 0 {
		s += ", has[\n"
		lv++
		for ix, has := range me.has {
			if ix > 0 {
				s += "\n"
			}
			s += has.string(lv)
		}
		lv--
		s += "\n"
		s += fmc(lv) + "]"
	}
	s += "}"
	return s
}

func (me *cnode) string(lv int) string {
	s := ""
	s += fmc(lv) + "{is:" + me.is
	if me.value != "" {
		s += ", value:" + me.value
	}
	s += ", typed:" + me.typed
	s += ", code:" + me.code
	if len(me.has) > 0 {
		s += ", has[\n"
		lv++
		for ix, has := range me.has {
			if ix > 0 {
				s += ",\n"
			}
			s += has.string(lv)
		}
		lv--
		s += "\n"
		s += fmc(lv) + "]"
	}
	s += "}"
	return s
}

func (me *class) dump(lv int) string {
	s := fmc(lv) + me.name + "[\n"
	lv++
	for _, cls := range me.variableOrder {
		classVar := me.variables[cls]
		s += fmc(lv) + "{name:" + classVar.name + ", typed:" + classVar.typed + "}\n"
	}
	lv--
	s += fmc(lv) + "]\n"
	return s
}

func (me *enum) dump(lv int) string {
	s := fmc(lv) + me.name + "[\n"
	lv++
	for _, unionType := range me.typesOrder {
		if len(unionType.types) > 0 {
			types := strings.Join(unionType.types, ", ")
			s += fmc(lv) + "{name:" + unionType.name + ", union:<" + types + ">}\n"
		} else {
			s += fmc(lv) + "{name:" + unionType.name + "}\n"
		}
	}
	lv--
	s += fmc(lv) + "]\n"
	return s
}

func (me *function) dump(lv int) string {
	s := fmc(lv) + me.name + "{\n"
	lv++
	if len(me.args) > 0 {
		s += fmc(lv) + "args[\n"
		lv++
		for _, arg := range me.args {
			s += fmc(lv) + arg.string() + "\n"
		}
		lv--
		s += fmc(lv) + "]\n"
	}
	if len(me.expressions) > 0 {
		s += fmc(lv) + "expressions[\n"
		lv++
		for _, expr := range me.expressions {
			s += expr.string(lv) + "\n"
		}
		lv--
		s += fmc(lv) + "]\n"
	}
	lv--
	s += fmc(lv) + "}\n"
	return s
}

func (me *hmfile) dump() string {
	s := ""
	lv := 0
	if len(me.defineOrder) > 0 {
		s += fmc(lv) + "define[\n"
		lv++
		for _, name := range me.defineOrder {
			def := strings.Split(name, "_")
			name := def[0]
			typed := def[1]
			if typed == "type" {
				cl := me.classes[name]
				s += cl.dump(lv)
			} else if typed == "enum" {
				en := me.enums[name]
				s += en.dump(lv)
			}
		}
		lv--
		s += fmc(lv) + "]\n"
	}
	if len(me.statics) > 0 {
		s += fmc(lv) + "static[\n"
		lv++
		for _, st := range me.statics {
			s += st.string(lv) + "\n"
		}
		lv--
		s += fmc(lv) + "]\n"
	}
	s += fmc(lv) + "functions[\n"
	lv++
	for _, name := range me.functionOrder {
		fn := me.functions[name]
		s += fn.dump(lv)
	}
	lv--
	s += fmc(lv) + "]\n"
	return s
}
