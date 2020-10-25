# Copyright 2020 The Predictive Horizontal Pod Autoscaler Authors.

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
This linear regression script performs a linear regression using the provided values and configuration using the
statsmodel library.
"""

import sys
import json
import math
from datetime import datetime, timedelta
import statsmodels.api as sm

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

print("Parsing stdin for linear regression parameters", file=sys.stderr)

parameters = json.loads(sys.stdin.read())

evaluations = parameters["evaluations"]
search_time = datetime.timestamp(datetime.now() + timedelta(milliseconds=int(parameters["lookAhead"])))

print("Calculating relative data for regression", file=sys.stderr)
x = []
y = []

# Build up data for linear model, in order to not deal with huge values and get rounding errors, use the difference
# between the time being searched for and the metric recorded time in seconds
for i, evaluation in enumerate(evaluations):
    x.append(search_time - datetime.timestamp(datetime.strptime(evaluation["created"], "%Y-%m-%dT%H:%M:%SZ")))
    y.append(int(evaluation["val"]["targetReplicas"]))

print("Building linear regression model", file=sys.stderr)

# Add constant for OLS, constant is 1.0
x = sm.add_constant(x)

model = sm.OLS(y, x).fit()

print("Making prediction using linear regression", file=sys.stderr)

# Predict the value at the search time (0), include the constant (1).
# The search time is 0 as the values used in training are search time - evaluation time, so the search time will be 0
print(math.ceil(model.predict([[1, 0]])[0]), end="")
