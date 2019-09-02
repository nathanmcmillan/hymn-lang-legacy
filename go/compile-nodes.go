package main

type cnode struct {
	is    string
	value string
	has   []*cnode
	typed string
	vdata *varData
	code  string
}

func (me *cnode) copyType(other *cnode) {
	me.typed = other.typed
	me.vdata = other.vdata
}

func (me *cnode) copyTypeFromVar(other *variable) {
	me.vdata = other.vdat
}

func (me *cnode) getType() string {
	if me.vdata != nil {
		return me.vdata.full
	}
	return me.typed
}

func (me *cnode) asVar(module *hmfile) *varData {
	if me.vdata != nil {
		return me.vdata
	}
	me.vdata = module.typeToVarData(me.typed)
	return me.vdata
}
