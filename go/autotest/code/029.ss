function printme int[]:a -> int[]
  for i = 0, i < 3, i += 1
    echo a[i]
  a[2] = 777
  return a

function main
  a = int[3]
  for i = 0, i < 3, i += 1
    a[i] = (i + 1) * 9
  printme a
  echo a[2]