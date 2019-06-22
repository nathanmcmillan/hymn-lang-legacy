class foo
  x int

foo dostuff int:z
  this.x += z

function main
  f = foo
  f.x = 3
  foo.dostuff 2
  echo foo.x
