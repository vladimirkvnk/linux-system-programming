#[cfg(test)]
mod e2e {
    use std::process::{Command, Stdio};

    const BINARY_PATH: &str = "./../target/debug/select";

    #[test]
    fn timeout() {
        let output = Command::new(BINARY_PATH)
            .stdin(Stdio::inherit())
            .output()
            .expect("Failed to execute command");

        assert_eq!(
            String::from_utf8_lossy(&output.stdout),
            "3 seconds elapsed.\n"
        );
    }
}
