package main

type compileError struct {
	code        int
	description string
	module      *hmfile
	line        int
	begin       int
	end         int
}

func errCompiling(description string) *compileError {
	e := &compileError{}
	e.description = description
	return e
}
