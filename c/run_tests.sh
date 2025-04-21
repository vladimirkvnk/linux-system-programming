#!/bin/sh

result="failed"

# Function to run a test and compare output
run_test() {
    test_name=$1
    command=$2
    expected_output=$3

    # Capture the output of the executable
    output=$(eval "$command")

    result="failed"

    # Compare the output with the expected value
    message="Test $test_name"
    if [ "$output" = "$expected_output" ]; then
        echo "$message: passed"
        result="passed"
    else
        echo "$message: failed"
        echo ""
        echo "-->Expected: '$expected_output'"
        echo "-->Got: '$output'"
        echo ""
        result="failed"
    fi
}

# Test poll
run_test "STDOUT should be writable" "./poll" "stdout is writeable"
run_test "STDOUT should be writable, STDIN readable" "./poll < $0" "stdin is readable
stdout is writeable"

# Test select
run_test "STDIN is readable with no data" "./select < /dev/null" "nothing read"
run_test "STDIN is readable, read some data" "echo 'some data' | ./select" "read: some data"

echo "**********************"
echo "-->>Result: $result<<--"
echo "**********************"
