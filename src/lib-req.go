package main

// Hymn Library
const (
	HmLibFiles  = "hmlib_files"
	HmLibSlice  = "hmlib_slice"
	HmLibString = "hmlib_string"
	HmLibSystem = "hmlib_system"
	HmLibMem    = "hmlib_mem"
)

// Dependencies
var (
	HmLibDependencies = map[string][]string{
		HmLibString: []string{HmLibMem},
		HmLibFiles:  []string{HmLibMem},
		HmLibSlice:  []string{HmLibMem},
	}
)

// C Standard Library
const (
	CStdUnistd   = "unistd"
	CStdIo       = "stdio"
	CStdLib      = "stdlib"
	CStdInt      = "stdint"
	CStdIntTypes = "inttypes"
	CStdBool     = "stdbool"
)
