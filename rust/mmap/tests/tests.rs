#[cfg(test)]
#[cfg(unix)]
mod e2e {
    use std::process::{Command, Stdio};

    const BINARY_PATH: &str = "./../target/debug/mmap";

    #[test]
    fn mmap_succeed() {
        let output = Command::new(BINARY_PATH)
            .arg("test.txt")
            .output()
            .expect("Failed to execute command");

        assert_eq!(String::from_utf8_lossy(&output.stdout), "1 2 3 4\n",);
    }
}
