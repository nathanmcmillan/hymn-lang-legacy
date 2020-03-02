package main

type parseError struct {
	code        int
	description string
	module      *hmfile
	line        int
	begin       int
	end         int
}

func err(parser *parser, description string) *parseError {
	e := &parseError{}
	e.description = parser.fail() + description
	return e
}

type compileError struct {
	code        int
	description string
	module      *hmfile
	line        int
	begin       int
	end         int
}

func errC(description string) *compileError {
	e := &compileError{}
	e.description = description
	return e
}
