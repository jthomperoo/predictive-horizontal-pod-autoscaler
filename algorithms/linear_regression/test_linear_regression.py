# Copyright 2022 The Predictive Horizontal Pod Autoscaler Authors.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
"""
Tests the linear regression algorithm by calling it from the shell, giving different stdin and checking the return
code and stderr and stdout.
"""
import subprocess


def test_linear_regression(subtests):
    """
    Test the linear regression algorithm
    """
    test_cases = [{
        "description": "Empty stdin",
        "expected_status_code": 1,
        "expected_stderr": "No standard input provided to Linear Regression algorithm, exiting\n",
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
        "JSON stdin missing 'lookAhead'",
        "expected_status_code":
        1,
        "expected_stderr":
        "Invalid JSON provided: missing 'look_ahead', exiting\n",
        "expected_stdout":
        "",
        "stdin":
        """{
                "replicaHistory": [
                    {
                        "time": "2020-02-01T00:55:33Z",
                        "replicas": 2
                    }
                ]
            }"""
    }, {
        "description":
        "Invalid timestamp provided",
        "expected_status_code":
        1,
        "expected_stderr":
        "Invalid datetime format: time data 'invalid' does not match format " + "'%Y-%m-%dT%H:%M:%SZ'\n",
        "expected_stdout":
        "",
        "stdin":
        """{
                "lookAhead": 10,
                "replicaHistory": [
                    {
                        "time": "invalid",
                        "replicas": 2
                    }
                ]
            }"""
    }, {
        "description":
        "Invalid current time provided",
        "expected_status_code":
        1,
        "expected_stderr":
        "Invalid datetime format: time data 'invalid' does not match format " + "'%Y-%m-%dT%H:%M:%SZ'\n",
        "expected_stdout":
        "",
        "stdin":
        """{
                "lookAhead": 15000,
                "currentTime": "invalid",
                "replicaHistory": []
            }"""
    }, {
        "description":
        "Successful prediction, now",
        "expected_status_code":
        0,
        "expected_stderr":
        "",
        "expected_stdout":
        "5",
        "stdin":
        """{
                "lookAhead": 0,
                "currentTime": "2020-02-01T00:56:12Z",
                "replicaHistory": [
                    {
                        "replicas": 1,
                        "time": "2020-02-01T00:55:33Z"
                    },
                    {
                        "replicas": 2,
                        "time": "2020-02-01T00:55:43Z"
                    },
                    {
                        "replicas": 3,
                        "time": "2020-02-01T00:55:53Z"
                    },
                    {
                        "replicas": 4,
                        "time": "2020-02-01T00:56:03Z"
                    }
                ]
            }"""
    }, {
        "description":
        "Successful prediction, 10 seconds in the future",
        "expected_status_code":
        0,
        "expected_stderr":
        "",
        "expected_stdout":
        "6",
        "stdin":
        """{
                "lookAhead": 10000,
                "currentTime": "2020-02-01T00:56:12Z",
                "replicaHistory": [
                    {
                        "replicas": 1,
                        "time": "2020-02-01T00:55:33Z"
                    },
                    {
                        "replicas": 2,
                        "time": "2020-02-01T00:55:43Z"
                    },
                    {
                        "replicas": 3,
                        "time": "2020-02-01T00:55:53Z"
                    },
                    {
                        "replicas": 4,
                        "time": "2020-02-01T00:56:03Z"
                    }
                ]
            }"""
    }, {
        "description":
        "Successful prediction, 15 seconds in the future",
        "expected_status_code":
        0,
        "expected_stderr":
        "",
        "expected_stdout":
        "7",
        "stdin":
        """{
                "lookAhead": 15000,
                "currentTime": "2020-02-01T00:56:12Z",
                "replicaHistory": [
                    {
                        "replicas": 1,
                        "time": "2020-02-01T00:55:33Z"
                    },
                    {
                        "replicas": 2,
                        "time": "2020-02-01T00:55:43Z"
                    },
                    {
                        "replicas": 3,
                        "time": "2020-02-01T00:55:53Z"
                    },
                    {
                        "replicas": 4,
                        "time": "2020-02-01T00:56:03Z"
                    }
                ]
            }"""
    }]

    for i, test_case in enumerate(test_cases):
        with subtests.test(msg=test_case["description"], i=i):
            result = subprocess.run(["python", "./algorithms/linear_regression/linear_regression.py"],
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
