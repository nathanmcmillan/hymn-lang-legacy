function square int:x int:z -> int
  x2 = x + 1
  z2 = z + 1
  w2 = 5 / 2
  return x2 * z2 * w2

function main
  x = 1
  z = 2
  w = square 3 5
  
  echo x
  echo z
  echo w