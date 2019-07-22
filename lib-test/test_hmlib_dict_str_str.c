#include "../lib/hmlib_dict_str_str.h"

char *hmlib_int_to_string(const int number)
{
  int len = snprintf(NULL, 0, "%d", number);
  char *str = malloc(len + 1);
  snprintf(str, len + 1, "%d", number);
  return str;
}

int main()
{
  HmLibDictStrStr *dict = hmlib_dict_str_str();

  hmlib_dict_str_str_set(dict, "foo", "bar");
  printf("get(%s) = %s\n", "foo", hmlib_dict_str_str_get(dict, "foo"));

  hmlib_dict_str_str_set(dict, "zoo", "tar");
  printf("get(%s) = %s\n", "zoo", hmlib_dict_str_str_get(dict, "zoo"));

  for (int i = 0; i < 32; i++)
  {
    hmlib_dict_str_str_set(dict, hmlib_int_to_string(i), hmlib_int_to_string(i + 1));
    printf("get(%s) = %s\n", hmlib_int_to_string(i), hmlib_dict_str_str_get(dict, hmlib_int_to_string(i)));
  }

  return 0;
}
