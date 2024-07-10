#[cfg(test)]
#[cfg(unix)]
mod e2e {
    use std::process::{Command, Stdio};

    const BINARY_PATH: &str = "./../target/debug/poll";

    #[test]
    fn stdout_writeable() {
        let output = Command::new(BINARY_PATH)
            .stdin(Stdio::piped())
            .output()
            .expect("Failed to execute command");

        assert_eq!(
            String::from_utf8_lossy(&output.stdout),
            "stdout is writeable\n"
        );
    }

    #[test]
    fn stdout_writeable_stdin_readable() {
        let output = Command::new(BINARY_PATH)
            .output()
            .expect("Failed to execute command");

        assert_eq!(
            String::from_utf8_lossy(&output.stdout),
            "stdin is readable\nstdout is writeable\n",
        );
    }
}
