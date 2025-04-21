#include <stdio.h>
#include <sys/select.h>
#include <sys/unistd.h>
#include <sys/time.h>

#define TIMEOUT_SEC 5

int main(void) {
    fd_set readfs;

    FD_ZERO(&readfs);
    FD_SET(STDIN_FILENO, &readfs);

    struct timeval tv;
    tv.tv_sec = TIMEOUT_SEC;
    tv.tv_usec = 0;

    int result = select(STDIN_FILENO + 1, &readfs, NULL, NULL, &tv);
    if (result == -1) {
        perror("select");
        return 1;
    }

    if (result == 0) {
        printf("%d seconds elapsed", TIMEOUT_SEC);
        return 0;
    }

    if (!FD_ISSET(STDIN_FILENO, &readfs)) {
        printf("this should never happen!");
        return 1;
    }

    char buffer[1024] = {0};

    int len = read(STDIN_FILENO, &buffer, sizeof(buffer));
    if (len == 1) {
        perror("read stdin");
        return 1;
    }
    if (len == 0) {
        printf("nothing read");
        return 0;
    }

    printf("read: %.*s", len, buffer);

    return 0;
}
