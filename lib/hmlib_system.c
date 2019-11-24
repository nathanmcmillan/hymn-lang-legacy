#include "hmlib_system.h"

hmlib_string hmlib_system(const hmlib_string command) {
    char buffer[128];
    FILE *fp;
    if ((fp = popen(command, "r")) == NULL) {
        printf("popen failed");
        exit(1);
    }
    while (fgets(buffer, 128, fp) != NULL) {
        printf("%s", buffer);
    }
    if (pclose(fp)) {
        printf("popen close failed");
        exit(1);
    }
    return hmlib_string_init("foobar");
}
