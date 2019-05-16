#include <stdio.h>

int square(int x, int z) {
  x = x + 1;
  z = z + 1;
  int w = 5 / 2;
  return x * z * w;
}

int main() {
  int x = 1;
  int z = 2;
  int w = square(3, 5);
  printf("%d\n", x);
  printf("%d\n", z);
  printf("%d\n", w);
  return 0;
}
