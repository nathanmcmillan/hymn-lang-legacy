#include "hmlib_strings.h"

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

  printf("int to string: %s\n", hmlib_int_to_string(4029));
  printf("int to string: %s\n", hmlib_int_to_string(-4029));
  printf("int to string: %s\n", hmlib_int_to_string(0));

  printf("float to string: %s\n", hmlib_float_to_string(3029.34));
  printf("float to string: %s\n", hmlib_float_to_string(-3029.34));
  printf("float to string: %s\n", hmlib_float_to_string(0.0));

  return 0;
}
