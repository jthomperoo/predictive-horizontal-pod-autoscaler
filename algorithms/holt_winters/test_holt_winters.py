"""
Tests the linear regression algorithm by calling it from the shell, giving different stdin and checking the return
code and stderr and stdout.
"""
import subprocess


def test_holt_winters(subtests):
    """
    Test the holt winters algorithm
    """
    test_cases = [{
        "description": "Empty stdin",
        "expected_status_code": 1,
        "expected_stderr": "No standard input provided to Holt-Winters algorithm, exiting\n",
        "expected_stdout": "",
        "stdin": ""
    }, {
        "description": "Invalid JSON stdin",
        "expected_status_code": 1,
        "expected_stderr": "Invalid JSON provided: Expecting value: line 1 column 1 (char 0), exiting\n",
        "expected_stdout": "",
        "stdin": "invalid"
    }, {
        "description":
        "JSON stdin missing 'trend'",
        "expected_status_code":
        1,
        "expected_stderr":
        "Invalid JSON provided: missing 'trend', exiting\n",
        "expected_stdout":
        "",
        "stdin":
        """{
                "seasonal": "add",
                "alpha": 0.9,
                "beta": 0.9,
                "gamma": 0.3,
                "seasonalPeriods": 3,
                "series": [1,3,1,1,3,1,1,3,1]
            }"""
    }, {
        "description":
        "Failure, less than required observations observations, 9 observations",
        "expected_status_code":
        1,
        "expected_stderr":
        "Invalid data provided, must be at least 10 + 2 * (seasonal_periods // 2) observations, exiting\n",
        "expected_stdout":
        "",
        "stdin":
        """{
                "trend": "add",
                "seasonal": "add",
                "alpha": 0.9,
                "beta": 0.9,
                "gamma": 0.3,
                "seasonalPeriods": 3,
                "series": [1,3,1,1,3,1,1,3,1]
            }"""
    }, {
        "description":
        "Failure, less than required observations observations, 2 observations",
        "expected_status_code":
        1,
        "expected_stderr":
        "Invalid data provided, must be at least 2 * seasonal_periods observations, exiting\n",
        "expected_stdout":
        "",
        "stdin":
        """{
                "trend": "add",
                "seasonal": "add",
                "alpha": 0.9,
                "beta": 0.9,
                "gamma": 0.3,
                "seasonalPeriods": 3,
                "series": [1,3]
            }"""
    }, {
        "description":
        "Successful prediction, additive, 13 observations",
        "expected_status_code":
        0,
        "expected_stderr":
        "",
        "expected_stdout":
        "3",
        "stdin":
        """{
                "trend": "add",
                "seasonal": "add",
                "alpha": 0.9,
                "beta": 0.9,
                "gamma": 0.3,
                "seasonalPeriods": 3,
                "series": [1,3,1,1,3,1,1,3,1,1,3,1,1]
            }"""
    }, {
        "description":
        "Successful prediction, multiplicative trend, 15 observations",
        "expected_status_code":
        0,
        "expected_stderr":
        "",
        "expected_stdout":
        "1",
        "stdin":
        """{
                "trend": "mul",
                "seasonal": "add",
                "alpha": 0.3,
                "beta": 0.3,
                "gamma": 0.9,
                "seasonalPeriods": 3,
                "series": [1,3,1,1,3,1,1,3,1,1,3,1,1,3,1]
            }"""
    }, {
        "description":
        "Successful prediction, additive trend + multiplicative seasonal, legacy-heuristic init " +
        "method, 19 observations",
        "expected_status_code":
        0,
        "expected_stderr":
        "",
        "expected_stdout":
        "6",
        "stdin":
        """{
                "trend": "add",
                "seasonal": "mul",
                "alpha": 0.005,
                "beta": 0.9,
                "gamma": 0.4,
                "seasonalPeriods": 3,
                "initialization_method": "legacy-heuristic",
                "series": [1,1,1,1,2,1,1,3,1,1,4,1,1,5,1,1,6,1,1]
            }"""
    }]

    for i, test_case in enumerate(test_cases):
        with subtests.test(msg=test_case["description"], i=i):
            result = subprocess.run(["python", "./algorithms/holt_winters/holt_winters.py"],
                                    input=test_case["stdin"].encode("utf-8"),
                                    capture_output=True,
                                    check=False)

            stderr = result.stderr
            if stderr is not None:
                stderr = stderr.decode("utf-8")

            stdout = result.stdout
            if stdout is not None:
                stdout = stdout.decode("utf-8")

            assert test_case["expected_status_code"] == result.returncode
            assert test_case["expected_stderr"] == stderr
            assert test_case["expected_stdout"] == stdout
