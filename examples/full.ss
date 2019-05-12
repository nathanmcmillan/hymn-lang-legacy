
module main

import
--

global
  HelloWorld = "hello world"
  mutable BigIndex = 1
--

interface abc
  do()
--

object foo
  x int32
  a string
  b []byte
  c def(int64)
--

def foo()
  print("foo")
--
       
def foo(string)
  print(arg0)
--

def foo() bool
  return true
--

function foo.bar(int32, int32, int32) int64 maybe:error
  num = len(foo.b)
  loop i = 0, i < num, i += 1
    print("index =", i)
  --
  return 2, none
--
