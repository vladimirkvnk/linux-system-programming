#include <stdio.h>
#include <unistd.h>
#include <poll.h>

#define TIMEOUT_SEC 5 /* Delay in poll */

int main (void)
{
    struct pollfd fds[2];
    int ret;

    /* Track input on stdin */
    fds[0].fd = STDIN_FILENO;
    fds[0].events = POLLIN;

    /* Track the write ability on stdout (which will always be true in practice). */
    fds[1].fd = STDOUT_FILENO;
    fds[1].events = POLLOUT;

    /* All set up, now block them all */
    ret = poll(fds, 2, TIMEOUT_SEC * 1000);
    if (ret == -1) {
        perror("poll");
        return 1;
    }

    if (!ret) {
        printf("%d seconds elapsed.\n", TIMEOUT_SEC);
        return 0;
    }

    if (fds[0].revents & POLLIN) {
        printf("stdin is readable\n");
    }

    if (fds[1].revents & POLLOUT) {
        printf("stdout is writeable\n");
    }

    return 0;
}
