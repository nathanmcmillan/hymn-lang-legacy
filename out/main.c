#include <stdio.h>

int square(int x, int z) {
  int x = x + 1;
  int z = z + 1;
  return x * z;
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
