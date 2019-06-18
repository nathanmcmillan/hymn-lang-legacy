function square int:x int:y -> int
  z = x + y * 5 - 7
  return z * z

function main
  w = square 2 3
  echo w
