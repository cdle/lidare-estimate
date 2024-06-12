#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <wiringPi.h>
#include <wiringSerial.h>
#include <unistd.h>
#include <uv.h>

#define BAUD_RATE 115200

uv_loop_t* loop;
uv_poll_t poll_handle2;
uv_poll_t poll_handle3;
int uart2_fd;
int uart3_fd;
extern void Receive(char*, char*);;

void on_poll2(uv_poll_t* handle, int status, int events)
{
    if (status < 0) {
        fprintf(stderr, "Error in uv_poll: %s\n", uv_strerror(status));
        return;
    }

    if (events & UV_READABLE) {
        char buffer[256];
        int bytesRead = read(uart2_fd, buffer, sizeof(buffer));
        if (bytesRead > 0) {
            buffer[bytesRead] = '\0';
            Receive("keting", buffer);
        }
    }
}

void on_poll3(uv_poll_t* handle, int status, int events)
{
    if (status < 0) {
        fprintf(stderr, "Error in uv_poll: %s\n", uv_strerror(status));
        return;
    }

    if (events & UV_READABLE) {
        char buffer[256];
        int bytesRead = read(uart3_fd, buffer, sizeof(buffer));
        if (bytesRead > 0) {
            buffer[bytesRead] = '\0';
            Receive("chufang", buffer);
        }
    }
}

int Start()
{
    if (wiringPiSetup() == -1) {
        fprintf(stderr, "Failed to initialize WiringPi\n");
        return 1;
    }

    uart2_fd = serialOpen("/dev/ttyAMA2", BAUD_RATE);
    if (uart2_fd == -1) {
        fprintf(stderr, "Failed to open serial port\n");
        return 1;
    }

    uart3_fd = serialOpen("/dev/ttyAMA3", BAUD_RATE);
    if (uart2_fd == -1) {
        fprintf(stderr, "Failed to open serial port\n");
        return 1;
    }

    loop = uv_default_loop();

    uv_poll_init(loop, &poll_handle2, uart2_fd);
    uv_poll_init(loop, &poll_handle3, uart3_fd);
    uv_poll_start(&poll_handle2, UV_READABLE, on_poll2);
    uv_poll_start(&poll_handle3, UV_READABLE, on_poll3);

    uv_run(loop, UV_RUN_DEFAULT);

    uv_poll_stop(&poll_handle2);
    uv_poll_stop(&poll_handle3);
    uv_close((uv_handle_t*)&poll_handle2, NULL);
    uv_close((uv_handle_t*)&poll_handle3, NULL);
    serialClose(uart2_fd);
    serialClose(uart3_fd);

    return 0;
}
