package main

// OrderedMap strings
type OrderedMap struct {
	order []string
	dict  map[string]string
}

func newOrderMap() *OrderedMap {
	m := &OrderedMap{}
	m.order = make([]string, 0)
	m.dict = make(map[string]string)
	return m
}

func (me *OrderedMap) add(k, v string) {
	if _, ok := me.dict[k]; !ok {
		me.order = append(me.order, k)
		me.dict[k] = v
	}
}

func (me *OrderedMap) has(k string) bool {
	_, ok := me.dict[k]
	return ok
}

func (me *OrderedMap) get(k string) string {
	v := me.dict[k]
	return v
}

// OrderedSet strings
type OrderedSet struct {
	order []string
	dict  map[string]bool
}

func newOrderSet() *OrderedSet {
	m := &OrderedSet{}
	m.order = make([]string, 0)
	m.dict = make(map[string]bool)
	return m
}

func (me *OrderedSet) add(k string) {
	if _, ok := me.dict[k]; !ok {
		me.order = append(me.order, k)
		me.dict[k] = true
	}
}

func (me *OrderedSet) has(k string) bool {
	_, ok := me.dict[k]
	return ok
}

func (me *OrderedSet) delete(k string) {
	if _, ok := me.dict[k]; ok {
		for i, f := range me.order {
			if f == k {
				me.order = append(me.order[:i], me.order[i+1:]...)
				break
			}
		}
		delete(me.dict, k)
	}
}
