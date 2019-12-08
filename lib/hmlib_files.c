#include "hmlib_files.h"

char *hmlib_read(const char *path) {
    FILE *fp = fopen(path, "r");
    if (fp == NULL) {
        printf("file not opened");
        exit(0);
    } else {
        char ch;
        while ((ch = fgetc(fp)) != EOF) {
            printf("%c", ch);
        }
    }
    fclose(fp);
    return NULL;
}

char *hmlib_buffer_read(const char *path) {
    FILE *fp = fopen(path, "r");
    if (fp == NULL) {
        printf("file not opened");
        exit(0);
    } else {
        const int size = 255;
        char buffer[size];
        if (fgets(buffer, size, fp)) {
        }
        printf("%s", buffer);
    }
    fclose(fp);
    return NULL;
}

char *hmlib_write(const char *path, const char *content) {
    FILE *fp = fopen(path, "w");
    if (fp == NULL) {
        printf("file not opened");
        exit(0);
    }
    fputs(content, fp);
    fclose(fp);
    return NULL;
}

size_t hmlib_file_size(const char *path) {
    FILE *fp = fopen(path, "r");
    if (fp == NULL) {
        printf("could not open file");
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
        printf("could not open file");
        exit(1);
    }
    char *content = malloc((size + 1) * sizeof(char));
    for (size_t i = 0; i < size; i++) {
        content[i] = fgetc(fp);
    }
    fclose(fp);
    hmlib_string s = hmlib_string_init_with_length(content, size);
    free(content);
    return s;
}
