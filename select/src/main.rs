use std::{ffi::CString, process};

use libc::{c_void, fd_set, perror, read, timeval, FD_ISSET, FD_SET, FD_ZERO, STDIN_FILENO};

fn main() {
    process::exit(select());
}

pub fn select() -> i32 {
    const TIMEOUT_SEC: i64 = 3;
    const BUF_LEN: usize = 1024;

    unsafe {
        let mut readfs: fd_set = std::mem::zeroed();

        // Wait on stdin for input
        FD_ZERO(&mut readfs);
        FD_SET(STDIN_FILENO, &mut readfs);

        let mut tv = timeval {
            tv_sec: TIMEOUT_SEC,
            tv_usec: 0,
        };

        let result = libc::select(
            STDIN_FILENO + 1,
            &mut readfs,
            std::ptr::null_mut(),
            std::ptr::null_mut(),
            &mut tv,
        );

        if result == -1 {
            let err = CString::new("Error in select").expect("Failed to create C string");
            perror(err.as_ptr());
            return -1;
        }

        if result == 0 {
            println!("{} seconds elapsed.", TIMEOUT_SEC);
            return 0;
        }

        /*
         * Is our file descriptor ready to read?
         * (It must be, as it was the only fd that
         * we provided and the call returned
         * nonzero, but we will humor ourselves.)
         */
        if !FD_ISSET(STDIN_FILENO, &readfs) {
            eprintln!("This should never happen!");
            return -1;
        }

        let mut buf: [u8; BUF_LEN] = [0; BUF_LEN];
        let len = read(STDIN_FILENO, buf.as_mut_ptr() as *mut c_void, BUF_LEN);

        match len {
            -1 => {
                let err = CString::new("Read").expect("Failed to create C string");
                perror(err.as_ptr());
                -1
            }
            _ => {
                if len != 0 {
                    println!("Read: {}", String::from_utf8_lossy(&buf[0..len as usize]))
                } else {
                    print!("Nothing read.");
                }
                0
            }
        }
    }
}
