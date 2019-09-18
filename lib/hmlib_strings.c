#include "hmlib_strings.h"

hmlib_string_head *hmlib_string_head_init(size_t len)
{
  size_t mem = sizeof(hmlib_string_head) + len + 1;
  hmlib_string_head *sh = (hmlib_string_head *)malloc(mem);
  memset(sh, 0, mem);
  sh->len = len;
  sh->cap = len;
  return sh;
}

hmlib_string hmlib_string_init_with_length(const char *init, size_t len)
{
  hmlib_string_head *sh = hmlib_string_head_init(len);
  char *s = (char *)sh + sizeof(hmlib_string_head);
  memcpy(s, init, len);
  s[len] = '\0';
  return (hmlib_string)s;
}

hmlib_string hmlib_string_concat(const hmlib_string a, const hmlib_string b)
{
  const size_t len_a = hmlib_string_len(a);
  const size_t len_b = hmlib_string_len(b);
  const size_t len = len_a + len_b;
  hmlib_string_head *sh = hmlib_string_head_init(len);
  char *s = (char *)sh + sizeof(hmlib_string_head);
  memcpy(s, a, len_a);
  memcpy(s + len_a, b, len_b + 1);
  s[len] = '\0';
  return (hmlib_string)s;
}

hmlib_string hmlib_string_init(const char *init)
{
  size_t len = strlen(init);
  return hmlib_string_init_with_length(init, len);
}

size_t hmlib_string_len(const hmlib_string s)
{
  hmlib_string_head *sh = (hmlib_string_head *)((char *)s - sizeof(hmlib_string_head));
  return sh->len;
}

void hmlib_string_free(const hmlib_string s)
{
  free((char *)s - sizeof(hmlib_string_head));
}

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

char *hmlib_int_to_string(const int number)
{
  int len = snprintf(NULL, 0, "%d", number);
  char *str = malloc(len + 1);
  snprintf(str, len + 1, "%d", number);
  return str;
}

char *hmlib_float_to_string(const float number)
{
  int len = snprintf(NULL, 0, "%f", number);
  char *str = malloc(len + 1);
  snprintf(str, len + 1, "%f", number);
  return str;
}

int hmlib_string_to_int(const char *str)
{
  return (int)strtol(str, NULL, 10);
}

float hmlib_string_to_float(const char *str)
{
  return strtof(str, NULL);
}
