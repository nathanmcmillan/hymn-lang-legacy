#include "hmlib_system.h"

hmlib_string hmlib_popen(const char *command) {
    char buffer[128];
    FILE *fp;
    if ((fp = popen(command, "r")) == NULL) {
        printf("popen failed");
        exit(1);
    }
    hmlib_string in = NULL;
    while (fgets(buffer, 128, fp) != NULL) {
        if (in == NULL) {
            in = hmlib_string_init(buffer);
        } else {
            hmlib_string cat = hmlib_string_append(in, buffer);
            if (cat != in) {
                hmlib_string_free(in);
                in = cat;
            }
        }
    }
    if (pclose(fp)) {
        printf("popen close failed");
        exit(1);
    }
    return in;
}

hmlib_string hmlib_system(hmlib_string command) {
    const int PARENT_WRITE_PIPE = 0;
    const int PARENT_READ_PIPE = 1;

    const int READ_FD = 0;
    const int WRITE_FD = 1;

    int pipes[2][2];

    if (pipe(pipes[PARENT_READ_PIPE]) != 0) {
        printf("pipe failed");
        exit(1);
    }

    if (pipe(pipes[PARENT_WRITE_PIPE]) != 0) {
        printf("pipe failed");
        exit(1);
    }

    if (fork()) {
        printf("fork\n");
        fflush(stdout);

        char buffer[256];
        int count;

        close(pipes[PARENT_WRITE_PIPE][READ_FD]);
        close(pipes[PARENT_READ_PIPE][WRITE_FD]);

        command = hmlib_string_append(command, "\n");
        size_t len = hmlib_string_len_size(command);

        if (write(pipes[PARENT_WRITE_PIPE][WRITE_FD], command, len)) {
        }

        count = read(pipes[PARENT_READ_PIPE][READ_FD], buffer, sizeof(buffer) - 1);
        if (count >= 0) {
            buffer[count] = 0;
            printf("%s", buffer);
        } else {
            printf("IO Error\n");
        }
    } else {
        printf("not fork\n");
        fflush(stdout);

        char *argv[] = {"/usr/bin/bc", "-q", NULL};

        dup2(pipes[PARENT_WRITE_PIPE][READ_FD], STDIN_FILENO);
        dup2(pipes[PARENT_READ_PIPE][WRITE_FD], STDOUT_FILENO);

        close(pipes[PARENT_WRITE_PIPE][READ_FD]);
        close(pipes[PARENT_READ_PIPE][WRITE_FD]);
        close(pipes[PARENT_READ_PIPE][READ_FD]);
        close(pipes[PARENT_WRITE_PIPE][WRITE_FD]);

        execv(argv[0], argv);
    }

    return hmlib_string_init("foo");
}

hmlib_system_std hmlib_system_help(const char *command) {
    char buffer[256];
    FILE *fp;
    if ((fp = popen(command, "r")) == NULL) {
        printf("popen failed");
        exit(1);
    }
    hmlib_string in = NULL;
    hmlib_string err = NULL;
    while (fgets(buffer, 256, fp) != NULL) {
        if (in == NULL) {
            in = hmlib_string_init(buffer);
        } else {
            hmlib_string cat = hmlib_string_append(in, buffer);
            if (cat != in) {
                hmlib_string_free(in);
                in = cat;
            }
        }
    }
    int code = 0;
    if (pclose(fp)) {
        printf("popen close failed");
        exit(1);
    }
    hmlib_system_std tuple = {in, err, code};
    return tuple;
}
