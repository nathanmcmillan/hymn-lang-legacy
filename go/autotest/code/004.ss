class foo
  a int
  b string

class bar
  c foo
  d string

function main
  x = foo
  z = bar
  z.c = x
  x.b = "hello test foo & bar"
  echo z.c.b