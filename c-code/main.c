#include <stdio.h>
#include <stdlib.h>

struct thing {
  int x;
  int y;
  int z;
};

int main() {
  struct thing *t;
  t = (struct thing *) malloc(sizeof(struct thing));
  (*t).x = 1;
  (*t).y = 2;
  (*t).z = 3;
  printf("%d, %d, %d\n", (*t).x, (*t).y, (*t).z);
  free(t);
  return 0;
}
