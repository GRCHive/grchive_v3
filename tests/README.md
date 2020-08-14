# Tests

To run all tests run: `bazel test --test_output=all //tests/...`.

If you wish to disable integration tests and only run unit tests run: `bazel test --test_output=all --define gotags=unit //tests/...`
