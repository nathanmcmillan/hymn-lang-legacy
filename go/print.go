package main

import (
	"strconv"
)

func (me *datatype) string(lv int) string {
	lv++
	s := "{\n"
	s += fmc(lv) + "\"is\": \"" + me.nameIs() + "\""
	if me.module != nil {
		name := me.module.cross(me.origin)
		if name == "" {
			name = me.module.reference(me.module.name)
		}
		s += ",\n" + fmc(lv) + "\"module\": \"" + name + "\""
	}
	if me.canonical != "" {
		s += ",\n" + fmc(lv) + "\"canonical\": \"" + me.canonical + "\""
	}
	if me.size != "" {
		s += ",\n" + fmc(lv) + "\"size\": \"" + me.size + "\""
	}
	if me.member != nil {
		s += ",\n" + fmc(lv) + "\"member\": " + me.member.string(lv)
	}
	if me.generics != nil {
		s += ",\n" + fmc(lv) + "\"generics\": [\n"
		lv++
		end := len(me.generics) - 1
		for i, v := range me.generics {
			s += fmc(lv) + v.string(lv)
			if i < end {
				s += ",\n"
			}
		}
		lv--
		s += "\n" + fmc(lv) + "]"

	}
	if me.parameters != nil {
		s += ",\n" + fmc(lv) + "\"parameters\": [\n"
		lv++
		end := len(me.parameters) - 1
		for i, v := range me.parameters {
			s += fmc(lv) + v.string(lv)
			if i < end {
				s += ",\n"
			}
		}
		lv--
		s += "\n" + fmc(lv) + "]"
	}
	if me.returns != nil {
		s += ",\n" + fmc(lv) + "\"returns\": " + me.returns.string(lv)
	}
	if me.class != nil {
		s += ",\n" + fmc(lv) + "\"class\": \"" + me.class.name + "\""
	}
	if me.enum != nil {
		s += ",\n" + fmc(lv) + "\"enum\": \"" + me.enum.name + "\""
		if me.union != nil {
			s += ",\n" + fmc(lv) + "\"union\": \"" + me.union.name + "\""
		}
	}
	lv--
	s += "\n" + fmc(lv) + "}"
	return s
}

func (me *variable) string(lv int) string {
	s := "{\n"
	lv++
	s += fmc(lv) + "\"name\": \"" + me.name + "\",\n"
	s += fmc(lv) + "\"data\": " + me.data().string(lv) + ",\n"
	s += fmc(lv) + "\"mutable\": " + strconv.FormatBool(me.mutable)
	lv--
	s += "\n" + fmc(lv) + "}"
	return s
}

func (me *node) string(current *hmfile, lv int) string {
	s := ""
	s += fmc(lv) + "{\n"
	lv++
	s += fmc(lv) + "\"is\": \"" + me.is + "\""
	if me.value != "" {
		s += ",\n" + fmc(lv) + "\"value\": \"" + me.value + "\""
	}
	if me.idata != nil {
		s += ",\n" + fmc(lv) + "\"id\": \"" + me.idata.string(current) + "\""
	}
	if me.fn != nil {
		s += ",\n" + fmc(lv) + "\"call\": \"" + me.fn.canonical(current) + "\""
	}
	if me.data() != nil {
		s += ",\n" + fmc(lv) + "\"data\": " + me.data().string(lv)
	}
	if len(me.attributes) > 0 {
		s += ",\n" + fmc(lv) + "\"attributes\": {\n"
		lv++
		ix := 0
		end := len(me.attributes) - 1
		for key, value := range me.attributes {
			s += fmc(lv) + "\"" + key + "\": \"" + value + "\""
			if ix < end {
				s += ",\n"
			}
			ix++
		}
		lv--
		s += "\n" + fmc(lv) + "}"
	}
	if len(me.has) > 0 {
		s += ",\n" + fmc(lv) + "\"has\": [\n"
		lv++
		end := len(me.has) - 1
		for i, has := range me.has {
			s += has.string(current, lv)
			if i < end {
				s += ",\n"
			}
		}
		lv--
		s += "\n"
		s += fmc(lv) + "]"
	}
	lv--
	s += "\n" + fmc(lv) + "}"
	return s
}

func (me *cnode) string(lv int) string {
	s := fmc(lv) + "{\"is\":" + me.is
	if me.value != "" {
		s += ", value:" + me.value
	}
	if me.typed != "" {
		s += ", typed:" + me.typed
	}
	if me.data() != nil {
		s += ", var:" + me.data().string(lv)
	}
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

func (me *codeblock) string(lv int) string {
	s := "["
	for i, n := range me.flatten() {
		if i != 0 {
			s += ", "
		}
		s += n.string(lv)
	}
	s += "]"
	return s
}

func (me *class) string(lv int) string {
	s := fmc(lv) + "\"" + me.name + "\": [\n"
	lv++
	end := len(me.variables) - 1
	for i, classVar := range me.variables {
		s += fmc(lv) + "{\n"
		lv++
		s += fmc(lv) + "\"name\": \"" + classVar.name + "\",\n" + fmc(lv)
		s += "\"typed\": " + classVar.data().string(lv)
		lv--
		s += "\n" + fmc(lv) + "}"
		if i < end {
			s += ","
		}
		s += "\n"
	}
	lv--
	s += fmc(lv) + "]"
	return s
}

func (me *enum) string(lv int) string {
	s := fmc(lv) + "\"" + me.name + "\": [\n"
	lv++
	end := len(me.typesOrder) - 1
	for i, unionType := range me.typesOrder {
		size := unionType.types.size()
		if size > 0 {
			s += fmc(lv) + "{\n"
			lv++
			s += fmc(lv) + "\"name\": \"" + unionType.name + "\",\n"
			s += fmc(lv) + "\"unions\": [\n"
			lv++
			unionEnd := size - 1
			for ix, typeKey := range unionType.types.order {
				typ := unionType.types.table[typeKey]
				s += fmc(lv) + typ.string(lv)
				if ix < unionEnd {
					s += ",\n"
				}
			}
			lv--
			s += "\n" + fmc(lv) + "]\n"
			lv--
			s += fmc(lv) + "}"
		} else {
			s += fmc(lv) + "{\n"
			lv++
			s += fmc(lv) + "\"name\": \"" + unionType.name + "\"\n"
			lv--
			s += fmc(lv) + "}"
		}
		if i < end {
			s += ",\n"
		}
	}
	lv--
	s += "\n" + fmc(lv) + "]"
	return s
}

func (me *function) string(current *hmfile, lv int) string {
	s := fmc(lv) + "\""
	s += me.getname()
	s += "\": {\n"
	lv++
	comma := false
	if me.interfaces != nil {
		if comma {
			s += ",\n"
		} else {
			comma = true
		}
		s += fmc(lv) + "\"interfaces\": {\n"
		lv++
		end := len(me.interfaces) - 1
		i := 0
		for k, v := range me.interfaces {
			s += fmc(lv) + "\"" + k + "\": [\n"
			lv++
			xend := len(v) - 1
			for xi, x := range v {
				s += fmc(lv) + "\"" + x.name + "\""
				if xi < xend {
					s += ",\n"
				}
			}
			lv--
			s += "\n" + fmc(lv) + "]"
			if i < end {
				s += ",\n"
			}
			i++
		}
		lv--
		s += "\n" + fmc(lv) + "}"
	}
	if len(me.args) > 0 {
		if comma {
			s += ",\n"
		} else {
			comma = true
		}
		s += fmc(lv) + "\"args\": [\n"
		lv++
		end := len(me.args) - 1
		for i, arg := range me.args {
			s += fmc(lv) + arg.string(lv)
			if i < end {
				s += ",\n"
			}
		}
		lv--
		s += "\n" + fmc(lv) + "]"
	}
	if me.returns != nil {
		if comma {
			s += ",\n"
		} else {
			comma = true
		}
		s += fmc(lv) + "\"returns\": " + me.returns.string(lv)
	}
	if len(me.expressions) > 0 {
		if comma {
			s += ",\n"
		} else {
			comma = true
		}
		s += fmc(lv) + "\"expressions\": [\n"
		lv++
		end := len(me.expressions) - 1
		for i, expr := range me.expressions {
			s += expr.string(current, lv)
			if i < end {
				s += ",\n"
			}
		}
		lv--
		s += "\n" + fmc(lv) + "]\n"
	}
	lv--
	s += fmc(lv) + "}"
	return s
}

func (me *hmfile) string() string {
	s := "{\n"
	lv := 1
	s += fmc(lv) + "\"uid\": \"" + me.uid + "\",\n"
	s += fmc(lv) + "\"name\": \"" + me.name + "\""
	if len(me.defineOrder) > 0 {
		s += ",\n"
		s += fmc(lv) + "\"define\": {\n"
		lv++
		end := len(me.defineOrder) - 1
		for i, c := range me.defineOrder {
			if c.class != nil {
				s += c.class.string(lv)
			} else if c.enum != nil {
				s += c.enum.string(lv)
			}
			if i < end {
				s += ","
			}
			s += "\n"
		}
		lv--
		s += fmc(lv) + "}"
	}
	if len(me.statics) > 0 {
		s += ",\n"
		s += fmc(lv) + "\"static\": [\n"
		lv++
		end := len(me.statics) - 1
		for i, st := range me.statics {
			s += st.string(me, lv) + "\n"
			if i < end {
				s += ","
			}
		}
		lv--
		s += fmc(lv) + "]"
	}
	s += ",\n"
	s += fmc(lv) + "\"functions\": {\n"
	lv++
	end := len(me.functionOrder) - 1
	for i, name := range me.functionOrder {
		fn := me.functions[name]
		s += fn.string(me, lv)
		if i < end {
			s += ","
		}
		s += "\n"
	}
	lv--
	s += fmc(lv) + "}\n}\n"
	return s
}
