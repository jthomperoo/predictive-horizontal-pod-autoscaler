# Copyright 2022 The Predictive Horizontal Pod Autoscaler Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# pylint: disable=no-member, invalid-name
"""
This linear regression script performs a linear regression using the provided values and configuration using the
statsmodel library.
"""

import sys
import math
from json import JSONDecodeError
from datetime import datetime, timedelta
from dataclasses import dataclass
from typing import List, Optional
import statsmodels.api as sm
from dataclasses_json import dataclass_json, LetterCase

# Takes in a replica history and the look ahead value:
# {
#   "lookAhead": 3,
#   "replicaHistory": [
#       {
#           "time": "2020-02-01T00:55:33Z",
#           "replicas": 3
#       },
#       {
#           "time": "2020-02-01T00:56:33Z",
#           "replicas": 6
#       }
#   ]
# }


@dataclass_json(letter_case=LetterCase.CAMEL)
@dataclass
class TimestampedReplica:
    """
    JSON data representation of a timestamped evaluation
    """
    time: str
    replicas: int


@dataclass_json(letter_case=LetterCase.CAMEL)
@dataclass
class AlgorithmInput:
    """
    JSON data representation of the data this algorithm requires to be provided to it.
    """
    look_ahead: int
    replica_history: List[TimestampedReplica]
    current_time: Optional[str] = None


stdin = sys.stdin.read()

if stdin is None or stdin == "":
    print("No standard input provided to Linear Regression algorithm, exiting", file=sys.stderr)
    sys.exit(1)

try:
    algorithm_input = AlgorithmInput.from_json(stdin)
except JSONDecodeError as ex:
    print(f"Invalid JSON provided: {str(ex)}, exiting", file=sys.stderr)
    sys.exit(1)
except KeyError as ex:
    print(f"Invalid JSON provided: missing {str(ex)}, exiting", file=sys.stderr)
    sys.exit(1)

current_time = datetime.utcnow()

if algorithm_input.current_time is not None:
    try:
        current_time = datetime.strptime(algorithm_input.current_time, "%Y-%m-%dT%H:%M:%SZ")
    except ValueError as ex:
        print(f"Invalid datetime format: {str(ex)}", file=sys.stderr)
        sys.exit(1)

search_time = datetime.timestamp(current_time + timedelta(milliseconds=int(algorithm_input.look_ahead)))

x = []
y = []

# Build up data for linear model, in order to not deal with huge values and get rounding errors, use the difference
# between the time being searched for and the metric recorded time in seconds
for i, timestamped_replica in enumerate(algorithm_input.replica_history):
    try:
        created = datetime.strptime(timestamped_replica.time, "%Y-%m-%dT%H:%M:%SZ")
    except ValueError as ex:
        print(f"Invalid datetime format: {str(ex)}", file=sys.stderr)
        sys.exit(1)

    x.append(search_time - datetime.timestamp(created))
    y.append(timestamped_replica.replicas)

# Add constant for OLS, constant is 1.0
x = sm.add_constant(x)

model = sm.OLS(y, x).fit()

# Predict the value at the search time (0), include the constant (1).
# The search time is 0 as the values used in training are search time - evaluation time, so the search time will be 0
print(math.ceil(model.predict([[1, 0]])[0]), end="")
