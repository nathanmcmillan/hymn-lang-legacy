package main

type parser struct {
	tokens []*token
	pos    int
}

func parse(tokens []*token) {
	parser := parser{}
	parser.tokens = tokens
	// for {
	// 	panic("parser failed on " + parser.fail())
	// }
	parser.eot()
}

func (me *parser) next() *token {
	t := me.tokens[me.pos]
	me.pos++
	return t
}

func (me *parser) peek() *token {
	return me.tokens[me.pos]
}

func (me *parser) eot() bool {
	return me.pos == len(me.tokens)
}

func (me *parser) fail() string {
	return "parser failed"
}
