package main

import "testing"

func TestCompile(t *testing.T) {
	t.Errorf("failed to compile %s", "bad argument")
}
