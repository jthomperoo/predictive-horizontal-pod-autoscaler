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
This holt winters script performs a holt winters calculation using the provided values and configuration using the
statsmodel library.
"""

import sys
import math
from json import JSONDecodeError
from dataclasses import dataclass
from typing import List, Optional
import warnings
import statsmodels.tsa.api as sm
from dataclasses_json import dataclass_json, LetterCase
from statsmodels.tools.sm_exceptions import ConvergenceWarning

warnings.simplefilter('ignore', ConvergenceWarning)

# Takes in the replica series, alpha, beta, gamma, season length, and the trend and seasonal (add vs mul)
# {
#   "trend": "add",
#   "seasonal": "mul"
#   "alpha": 0.1,
#   "beta": 0.1,
#   "gamma": 0.1,
#   "seasonalPeriods": 5,
#   "series": [
#       0,
#       1,
#       2,
#       3,
#       4
#   ],
#   "dampedTrend": false,
# }


@dataclass_json(letter_case=LetterCase.CAMEL)
@dataclass
class AlgorithmInput:
    """
    JSON data representation of the data this algorithm requires to be provided to it.
    """
    trend: str
    seasonal: str
    seasonal_periods: int
    alpha: float
    beta: float
    gamma: float
    series: List[int]
    damped_trend: bool = False
    initialization_method: str = "estimated"
    initial_level: Optional[float] = None
    initial_trend: Optional[float] = None
    initial_seasonal: Optional[float] = None


stdin = sys.stdin.read()

if stdin is None or stdin == "":
    print("No standard input provided to Holt-Winters algorithm, exiting", file=sys.stderr)
    sys.exit(1)

try:
    algorithm_input = AlgorithmInput.from_json(stdin)
except JSONDecodeError as ex:
    print(f"Invalid JSON provided: {str(ex)}, exiting", file=sys.stderr)
    sys.exit(1)
except KeyError as ex:
    print(f"Invalid JSON provided: missing {str(ex)}, exiting", file=sys.stderr)
    sys.exit(1)

if len(algorithm_input.series) < 2 * algorithm_input.seasonal_periods:
    print("Invalid data provided, must be at least 2 * seasonal_periods observations, exiting", file=sys.stderr)
    sys.exit(1)

if len(algorithm_input.series) < 10 + 2 * (algorithm_input.seasonal_periods // 2):
    print("Invalid data provided, must be at least 10 + 2 * (seasonal_periods // 2) observations, exiting",
          file=sys.stderr)
    sys.exit(1)

model = sm.ExponentialSmoothing(algorithm_input.series,
                                trend=algorithm_input.trend,
                                seasonal=algorithm_input.seasonal,
                                seasonal_periods=algorithm_input.seasonal_periods,
                                initialization_method=algorithm_input.initialization_method,
                                damped_trend=algorithm_input.damped_trend,
                                initial_level=algorithm_input.initial_level,
                                initial_trend=algorithm_input.initial_trend,
                                initial_seasonal=algorithm_input.initial_seasonal)

fitted_model = model.fit(smoothing_level=algorithm_input.alpha,
                         smoothing_trend=algorithm_input.beta,
                         smoothing_seasonal=algorithm_input.gamma,
                         optimized=False)

# Predict the value one ahead
print(math.ceil(fitted_model.forecast(steps=1)[0]), end="")
