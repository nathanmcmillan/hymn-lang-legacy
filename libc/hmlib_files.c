#include "hmlib_files.h"

size_t hmlib_file_size(const char *path) {
    FILE *fp = fopen(path, "r");
    if (fp == NULL) {
        printf("Could not open file: %s", path);
        exit(1);
    }
    size_t num = 0;
    char ch;
    while ((ch = fgetc(fp)) != EOF) {
        num++;
    }
    fclose(fp);
    return num;
}

hmlib_string hmlib_cat(const char *path) {
    size_t size = hmlib_file_size(path);
    FILE *fp = fopen(path, "r");
    if (fp == NULL) {
        printf("Could not open file: %s", path);
        exit(1);
    }
    char *content = hmlib_malloc((size + 1) * sizeof(char));
    for (size_t i = 0; i < size; i++) {
        content[i] = fgetc(fp);
    }
    fclose(fp);
    hmlib_string s = hmlib_string_init_with_length(content, size);
    free(content);
    return s;
}

void hmlib_write(const char *path, const char *content) {
    FILE *fp = fopen(path, "a");
    if (fp == NULL) {
        printf("Could not open file: %s", path);
        exit(1);
    }
    fputs(content, fp);
    fclose(fp);
}
