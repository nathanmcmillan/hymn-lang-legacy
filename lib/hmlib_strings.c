#include "hmlib_strings.h"

char *hmlib_concat(const char *a, const char *b)
{
  const size_t len1 = strlen(a);
  const size_t len2 = strlen(b);
  char *cat = calloc(sizeof(char), len1 + len2 + 1);
  memcpy(cat, a, len1);
  memcpy(cat + len1, b, len2 + 1);
  return cat;
}

char *hmlib_concat_list(const char **ls, const int size)
{
  size_t len = 1;
  for (int i = 0; i < size; i++)
  {
    len += strlen(ls[i]);
  }
  char *cat = calloc(sizeof(char), len);
  size_t pos = 0;
  for (int i = 0; i < size; i++)
  {
    size_t len_i = strlen(ls[i]);
    memcpy(cat + pos, ls[i], len_i);
    pos += len_i;
  }
  return cat;
}

char *hmlib_concat_varg(const int size, ...)
{
  va_list ap;

  size_t len = 1;
  va_start(ap, size);
  for (int i = 0; i < size; i++)
  {
    len += strlen(va_arg(ap, char *));
  }
  va_end(ap);

  char *cat = calloc(sizeof(char), len);
  size_t pos = 0;

  va_start(ap, size);
  for (int i = 0; i < size; i++)
  {
    char *param = va_arg(ap, char *);
    size_t len_i = strlen(param);
    memcpy(cat + pos, param, len_i);
    pos += len_i;
  }
  va_end(ap);

  return cat;
}

int main()
{
  const char *one = "foo";
  const char *two = " bar";
  const char *three = hmlib_concat(one, two);
  printf("%s\n", three);
  const char *four = "hello";
  printf("%s\n", hmlib_concat(four, " world"));

  const char *many[3];
  many[0] = "my";
  many[1] = " cool";
  many[2] = " string";
  printf("%s\n", hmlib_concat_list(many, 3));

  printf("%s\n", hmlib_concat_varg(3, "super", " cool", " varg"));

  return 0;
}
