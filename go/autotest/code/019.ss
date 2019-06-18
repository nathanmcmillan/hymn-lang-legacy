function foo int:x -> int
  return x + 2

function bar int:x int:y -> int
  return x + y

function main
  x = 3
  y = 4
  if x * x = bar x y + 3
    echo "true"
  if x * x = foo y + 3
    echo "true"
  if x - x = (foo y) + 3
    echo "true"
  if x / x = foo (y + 3)
    echo "true"
  if x + x = y + 2
    echo "true"
  else
    echo "false"
  echo "bye!"
