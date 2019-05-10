package main

type scope struct {
	mutables   map[string]string
	immutables map[string]string
	defs       map[string]string
}

var global = scope{}
