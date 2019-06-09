class foo
  a int
  b string

class bar
  c foo
  d string

function main
  x = new foo
  z = new bar
  z.c = x
  x.b = "hello test J"
  echo z.c.b
