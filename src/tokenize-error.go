package main

type tokenizeError struct {
	reason string
}

func (me *tokenizer) exception(reason string) *tokenizeError {
	t := &tokenizeError{}
	t.reason = reason
	return t
}
