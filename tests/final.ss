import
  math

variables
  mutable nextId
  name = "foo"

class foo
  f float32[]

function glob_foo int:x -> int
  return x + 2

function appender list<string>:ls
   ls[1] = "foo"

function setter map<string int>:mp
  mp["foo"] = 99

function main
  x = new foo
  x.a = 2
  z = x.a
  echo z
  echo x.a
  free x

