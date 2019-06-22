function less int:a int:b -> bool
  return a < b

function ls int:a int:b -> bool
  if a < b
    return true
  return false

function gr int:a int:b -> bool
  if a > b
    return true
  return false

function le int:a int:b -> bool
  if a <= b
    return true
  return false

function ge int:a int:b -> bool
  if a >= b
    return true
  return false

function eq int:a int:b -> bool
  if a = b
    return true
  return false

function main
  # this is a comment
  x = 2
  y = 2
  echo less x y
  echo ls x y
  echo gr x y
  echo le x y
  echo ge x y
  echo eq x y
  e = x = y
  g = x > y
  f = x < y
  w = x >= y
  k = x <= y
  p = true
  echo p
  echo e
  echo g
  echo f
  echo w
  echo k
  if p
    echo "good"
  else
    echo "bad"
  if true
    echo "good"
  else
    echo "bad"
  if false
    echo "bad"
  else
    echo "good"
  if true = true
    echo "good"
  else
    echo "bad"
  if false = false
    echo "good"
  else
    echo "bad"
