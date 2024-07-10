use std::{ffi::CString, process};

use libc::{perror, pollfd, POLLIN, POLLOUT, STDIN_FILENO, STDOUT_FILENO};

fn main() {
    #[cfg(unix)]
    {
        process::exit(poll());
    }

    #[cfg(not(unix))]
    {
        println!("This program is intended to run on Unix systems only.");
        process::exit(-1);
    }
}

pub fn poll() -> i32 {
    const TIMEOUT_SEC: i32 = 3;

    let mut fds: [pollfd; 2] = [
        pollfd {
            fd: STDIN_FILENO,
            events: POLLIN,
            revents: 0,
        },
        pollfd {
            fd: STDOUT_FILENO,
            events: POLLOUT,
            revents: 0,
        },
    ];

    unsafe {
        let result = libc::poll(fds.as_mut_ptr(), 2, TIMEOUT_SEC * 1000);

        if result == -1 {
            let err = CString::new("Error in select").expect("Failed to create C string");
            perror(err.as_ptr());
            return -1;
        }

        if result == 0 {
            println!("{} seconds elapsed.", TIMEOUT_SEC);
            return 0;
        }

        if fds[0].revents & POLLIN != 0 {
            println!("stdin is readable");
        }

        if fds[1].revents & POLLOUT != 0 {
            println!("stdout is writeable");
        }
    }

    0
}
