class foo
  a int

function main
  x = foo[3]
  x[1] = new foo
  x[1].a = 9
  echo x[1].a
