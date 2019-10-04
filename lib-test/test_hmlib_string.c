#include "../lib/hmlib_string.h"

int main()
{
  hmlib_string s = hmlib_string_init("hello special string");
  printf("string = %s\n", s);
  printf("length = %ld\n", hmlib_string_len(s));

  hmlib_string sb = hmlib_string_init(". foo bar");
  printf("string = %s\n", sb);
  printf("length = %ld\n", hmlib_string_len(sb));

  hmlib_string sc = hmlib_string_concat(s, sb);
  printf("string = %s\n", sc);
  printf("length = %ld\n", hmlib_string_len(sc));

  hmlib_string_free(s);
  hmlib_string_free(sb);
  hmlib_string_free(sc);

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
