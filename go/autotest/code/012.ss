class foo
  val string

function main
  x = 3
  y = 2
  arr = foo[x + 1]
  arr[1 + y] = foo
  arr[y + 1].val = "bye!"
  echo arr[y + 1 / 1].val
