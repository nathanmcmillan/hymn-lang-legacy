function main
  mutable i = 0
  for
    echo "loop!"
    if i > 4
      echo "break!"
      break
    i += 1
  echo "bye!"
    