#include "../lib/hmlib_dict.h"

char *hmlib_int_to_string(const int number)
{
  int len = snprintf(NULL, 0, "%d", number);
  char *str = malloc(len + 1);
  snprintf(str, len + 1, "%d", number);
  return str;
}

int main()
{
  HmLibDict *dict = hmlib_dict();

  printf("get: %s\n", hmlib_dict_get(dict, 3982));
  printf("get: %s\n", hmlib_dict_get(dict, 0));

  hmlib_dict_set(dict, 3982, "foo");
  hmlib_dict_set(dict, 0, "bar");
  hmlib_dict_set(dict, 3983, "zoo");

  printf("get: %s\n", hmlib_dict_get(dict, 3982));
  printf("get: %s\n", hmlib_dict_get(dict, 3983));
  printf("get: %s\n", hmlib_dict_get(dict, 0));

  hmlib_dict_set(dict, 0, "tar");
  printf("get: %s\n", hmlib_dict_get(dict, 0));

  for (int i = 0; i < 256; i++)
  {
    hmlib_dict_set(dict, i, hmlib_int_to_string(i));
    printf("get(%d) = %s\n", i, hmlib_dict_get(dict, i));
  }

  return 0;
}
