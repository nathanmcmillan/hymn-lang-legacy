object foo
  a

function main
  x = new foo
  x.a = 2
  z = x.a
  echo z
  echo x.a
  delete x
