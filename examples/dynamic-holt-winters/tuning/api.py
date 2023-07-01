# Copyright 2023 The Predictive Horizontal Pod Autoscaler Authors.
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

from flask import Flask, request
import json

app = Flask(__name__)

ALPHA_VALUE = 0.9
BETA_VALUE = 0.9
GAMMA_VALUE = 0.9

@app.route("/holt_winters")
def metric():
    app.logger.info('Received context: %s', request.data)

    return json.dumps({
        "alpha": ALPHA_VALUE,
        "beta": BETA_VALUE,
        "gamma": GAMMA_VALUE,
    })

if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0")
