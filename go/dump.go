package main

func (me *variable) string() string {
	return "{is:" + me.is + ", name:" + me.name + "}"
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
	if me.attribute != "" {
		s += ", attribute:" + me.attribute
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

func (me *program) dump() string {
	s := ""
	lv := 0
	if len(me.classOrder) > 0 {
		s += fmc(lv) + "classes[\n"
		for _, name := range me.classOrder {
			class := me.classes[name]
			lv++
			s += fmc(lv) + class.name + "[\n"
			lv++
			for _, classVar := range class.variables {
				s += fmc(lv) + "{name:" + classVar.name + ", is:" + classVar.is + "}\n"
			}
			lv--
			s += fmc(lv) + "]\n"
			lv--
		}
		s += fmc(lv) + "]\n"
	}
	s += fmc(lv) + "functions[\n"
	for _, name := range me.functionOrder {
		function := me.functions[name]
		lv++
		s += fmc(lv) + name + "{\n"
		lv++
		if len(function.args) > 0 {
			s += fmc(lv) + "args[\n"
			lv++
			for _, arg := range function.args {
				s += fmc(lv) + arg.string() + "\n"
			}
			lv--
			s += fmc(lv) + "]\n"
		}
		if len(function.expressions) > 0 {
			s += fmc(lv) + "expressions[\n"
			lv++
			for _, expr := range function.expressions {
				s += expr.string(lv) + "\n"
			}
			lv--
			s += fmc(lv) + "]\n"
		}
		lv--
		s += fmc(lv) + "}\n"
		lv--
	}
	s += fmc(lv) + "]\n"
	return s
}
