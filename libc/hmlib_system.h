#ifndef HMLIB_SYSTEM_H
#define HMLIB_SYSTEM_H

#define _GNU_SOURCE

#include "hmlib_string.h"
#include <stdarg.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

typedef struct hmlib_system_std hmlib_system_std;

struct hmlib_system_std {
    hmlib_string stdin;
    hmlib_string stdout;
    int code;
};

hmlib_string hmlib_popen(const char *command);
hmlib_string hmlib_system(const hmlib_string command);

#endif
