#include "hmlib_files.h"

char *hmlib_read(const char *path)
{
  FILE *fp = fopen(path, "r");
  if (fp == NULL)
  {
    printf("file not opened");
    exit(0);
  }
  else
  {
    char ch;
    while ((ch = fgetc(fp)) != EOF)
    {
      printf("%c", ch);
    }
  }
  fclose(fp);
}

char *hmlib_buffer_read(const char *path)
{
  FILE *fp = fopen(path, "r");
  if (fp == NULL)
  {
    printf("file not opened");
    exit(0);
  }
  else
  {
    const int size = 255;
    char buffer[size];
    fgets(buffer, size, fp);
    printf("%s", buffer);
  }
  fclose(fp);
}

char *hmlib_write(const char *path, const char *content)
{
  FILE *fp = fopen(path, "w");
  if (fp == NULL)
  {
    printf("file not opened");
    exit(0);
  }
  fputs(content, fp);
  fclose(fp);
}

int main()
{
  hmlib_write("test_write.txt", "hello file\n");
  hmlib_read("test_write.txt");
  hmlib_buffer_read("test_write.txt");
  return 0;
}
