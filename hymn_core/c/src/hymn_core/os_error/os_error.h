#ifndef HM_HYMN_CORE_OS_ERROR_H
#define HM_HYMN_CORE_OS_ERROR_H

#include "hmlib_string.h"

typedef struct HmOsError HmOsError;
struct HmOsError {
    int code;
    hmlib_string reason;
};

#endif
