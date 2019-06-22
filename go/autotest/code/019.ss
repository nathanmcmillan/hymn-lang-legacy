function foo int:x -> int
  return x + 2

function bar int:x int:y -> int
  return x + y

function main
  x = 3
  y = 4
  # 9 == 10
  if x * x = bar x y + 3
    echo "a"
  # 9 == 9
  if x * x = foo y + 3 
    echo "b"
  # 0 == 9
  if x - x = (foo y) + 3
    echo "c"
  # 1 == 8
  if x / x = foo (y + 3)
    echo "d"
  # 6 == 6
  if x + x = y + 2
    echo "e"
  else
    echo "f"
  # 1 == 8
  if foo 8 * 5 = 50
    echo "yes!"
  if foo (8 * 5) = 42
    echo "yes!!"
  echo "bye!"
