package main

type ordereddata struct {
	order []string
	table map[string]*datatype
}

func newordereddata() *ordereddata {
	d := &ordereddata{}
	d.order = make([]string, 0)
	d.table = make(map[string]*datatype)
	return d
}

func (me *ordereddata) push(key string, data *datatype) {
	me.order = append(me.order, key)
	me.table[key] = data
}

func (me *ordereddata) size() int {
	return len(me.order)
}

func (me *ordereddata) get(index int) *datatype {
	key := me.order[index]
	return me.table[key]
}
