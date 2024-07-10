#[cfg(test)]
#[cfg(unix)]
mod e2e {
    use std::{
        io::Write,
        process::{Command, Stdio},
    };

    const BINARY_PATH: &str = "./../target/debug/select";

    #[test]
    fn timeout() {
        let output = Command::new(BINARY_PATH)
            .stdin(Stdio::null())
            .output()
            .expect("Failed to execute command");

        assert_eq!(String::from_utf8_lossy(&output.stdout), "Nothing read.\n");
    }

    #[test]
    fn read_string() {
        let mut child = Command::new(BINARY_PATH)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .spawn()
            .expect("Failed to spawn process");

        // Access the child's stdin and write to it
        if let Some(ref mut stdin) = child.stdin {
            stdin
                .write_all(b"simple_string")
                .expect("Failed to write to stdin");
        }

        // Retrieve and print the output of the child process
        let output = child.wait_with_output().expect("Failed to read stdout");

        assert_eq!(
            String::from_utf8_lossy(&output.stdout),
            "Read: simple_string\n"
        );
    }
}
