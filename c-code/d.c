#include <stdio.h>

int square(int x) {
  int z = x + 1;
  return z * z;
}

int main() {
  int z = square(3);
  printf("%d\n", z);
  return 0;
}
