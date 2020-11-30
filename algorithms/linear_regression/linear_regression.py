# Copyright 2020 The Predictive Horizontal Pod Autoscaler Authors.
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

# Takes in list of stored evaluations and the look ahead value:
# {
#   "lookAhead": 3,
#   "evaluations": [
#       {
#           "id": 0,
#           "created": "2020-02-01T00:55:33Z",
#           "val": {
#               "targetReplicas": 3
#           }
#       },
#       {
#           "id": 1,
#           "created": "2020-02-01T00:56:33Z",
#           "val": {
#               "targetReplicas": 2
#           }
#       }
#   ]
# }

@dataclass_json(letter_case=LetterCase.CAMEL)
@dataclass
class EvaluationValue:
    """
    JSON data representation of an evaluation value, contains the scaling target replicas
    """
    target_replicas: int

@dataclass_json(letter_case=LetterCase.CAMEL)
@dataclass
class Evaluation:
    """
    JSON data representation of a timestamped evaluation
    """
    id: int
    created: str
    val: EvaluationValue

@dataclass_json(letter_case=LetterCase.CAMEL)
@dataclass
class AlgorithmInput:
    """
    JSON data representation of the data this algorithm requires to be provided to it.
    """
    look_ahead: int
    evaluations: List[Evaluation]
    current_time: Optional[str] = None

stdin = sys.stdin.read()

if stdin is None or stdin == "":
    print("No standard input provided to Linear Regression algorithm, exiting", file=sys.stderr)
    sys.exit(1)

try:
    algorithm_input = AlgorithmInput.from_json(stdin)
except JSONDecodeError as ex:
    print("Invalid JSON provided: {0}, exiting".format(str(ex)), file=sys.stderr)
    sys.exit(1)
except KeyError as ex:
    print("Invalid JSON provided: missing {0}, exiting".format(str(ex)), file=sys.stderr)
    sys.exit(1)

current_time = datetime.utcnow()

if algorithm_input.current_time is not None:
    try:
        current_time = datetime.strptime(algorithm_input.current_time, "%Y-%m-%dT%H:%M:%SZ")
    except ValueError as ex:
        print("Invalid datetime format: {0}".format(str(ex)), file=sys.stderr)
        sys.exit(1)

search_time = datetime.timestamp(current_time + timedelta(milliseconds=int(algorithm_input.look_ahead)))

x = []
y = []

# Build up data for linear model, in order to not deal with huge values and get rounding errors, use the difference
# between the time being searched for and the metric recorded time in seconds
for i, evaluation in enumerate(algorithm_input.evaluations):
    try:
        created = datetime.strptime(evaluation.created, "%Y-%m-%dT%H:%M:%SZ")
    except ValueError as ex:
        print("Invalid datetime format: {0}".format(str(ex)), file=sys.stderr)
        sys.exit(1)

    x.append(search_time - datetime.timestamp(created))
    y.append(evaluation.val.target_replicas)

# Add constant for OLS, constant is 1.0
x = sm.add_constant(x)

model = sm.OLS(y, x).fit()

# Predict the value at the search time (0), include the constant (1).
# The search time is 0 as the values used in training are search time - evaluation time, so the search time will be 0
print(math.ceil(model.predict([[1, 0]])[0]), end="")
