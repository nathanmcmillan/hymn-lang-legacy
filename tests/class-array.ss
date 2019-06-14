class foo
  val string

function main
  arr = foo[3]
  arr[1] = new foo
  arr[1].val = "foo!"
  echo arr[1].val
