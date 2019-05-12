#include <stdio.h>

int square(int x, int y) {
  int z = x + y;
  return z * z;
}

int main() {
  int w = square(3);
  printf("%d\n", w);
  return 0;
}
