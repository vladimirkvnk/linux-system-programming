#!/bin/sh

# Function to run a test and compare output
run_test() {
    test_name=$1
    command=$2
    expected_output=$3

    # Capture the output of the executable
    output=$(eval "$command")

    # Compare the output with the expected value
    message="Test $command: $test_name"
    if [ "$output" = "$expected_output" ]; then
        echo "$message: Test passed"
    else
        echo "$message: Test failed"
        echo "Expected: '$expected_output'"
        echo "Got: '$output'"
    fi
}

# Test poll
run_test "STDOUT should be writable" "./poll" "stdout is writeable"
run_test "STDOUT should be writable, STDIN readable" "./poll < $0" "stdin is readable
stdout is writeable"
