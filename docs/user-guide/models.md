# Models

## Shared Configuration

All models share these three options:

- **type** - The type of the model, for example 'Linear'.
- **name** - The name of the model, must be unique and not shared by multiple models.
- **perInterval** - The frequency that the model is used to recalculate and store values - tied to the interval as a
base unit, with a value of `1` resulting in the model being recalculated every interval, a value of `2` meaning
recalculated every other interval, `3` waits for two intervals after every calculation and so on.

All models use `interval` as a base unit, so if the interval is defined as `10000` (10 seconds), the models will base
their timings and calculations as multiples of 10 seconds.

## Linear Regression
Example:
```yaml
- name: predictiveConfig
  value: |
    models:
    - type: Linear
      name: LinearPrediction
      perInterval: 1
      linear:
        lookAhead: 10000
        storedValues: 6
      decisionType: "maximum"
```
The **linear** component of the configuration handles configuration of the Linear regression options:

- **lookAhead** - sets up the model to try to predict `10 seconds` ahead of time (time in milliseconds).
- **storedValues** - sets up the model to store the past `6` evaluations and to use these for predictions. If there
are `> 6` evaluations, the oldest will be removed.

For a more detailed example, [see the example in `/example/simple-linear`](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/tree/master/example/simple-linear).

## Holt-Winters Time Series prediction

Example:
```yaml
- name: predictiveConfig
  value: |
    models:
    - type: HoltWinters
      name: HoltWintersPrediction
      perInterval: 1
      holtWinters:
        alpha: 0.9
        beta: 0.9
        gamma: 0.9
        seasonLength: 6
        storedSeasons: 4
        method: "additive"
    decisionType: "maximum"
```
The **holtWinters** component of the configuration handles configuration of the Linear regression options:

- **alpha**, **beta**, **gamma** - these are the smoothing coefficients for level, trend and seasonality respectively,
requires tweaking and analysis to be able to optimise. See [here](https://github.com/jthomperoo/holtwinters) or
[here](https://grisha.org/blog/2016/01/29/triple-exponential-smoothing-forecasting/) for more details.
- **seasonLength** - the length of a season in base unit intervals, for example if your interval was `10000`
(10 seconds), and your repeated season was 60 seconds long, this value would be `6`.
- **storedSeasons** - the number of seasons to store, for example `4`, if there are `>4` seasons stored, the oldest
season will be removed.

This is the model in action, taken from the `simple-holt-winters` example:
![Predicted values overestimating but still fitting actual values](../img/holt_winters_prediction_vs_actual.svg)
The red value is the predicted values, the blue value is the actual values. From this you can see that the prediction
is overestimating, but still pre-emptively scaling - storing more seasons and adjusting alpha, beta and gamma values
would reduce the overestimation and produce more accurate results.

For a more detailed example, [see the example in
`/example/simple-holt-winters`](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/tree/master/example/simple-holt-winters).

### Holt-Winters Runtime Tuning

The PHPA supports dynamically fetching the tuning values for the Holt-Winters algorithm (`alpha`, `beta`, and `gamma`).
This is done using a feature of the Custom Pod Autoscaler called
[methods](https://custom-pod-autoscaler.readthedocs.io/en/latest/user-guide/methods/) - these are configurable hooks
that allow the PHPA to reach out to an external source to determine what the tuning values should be.

For example, a method using a HTTP request to fetch the values of runtime is configured as:

```yaml
- name: predictiveConfig
  value: |
    models:
    - type: HoltWinters
      name: HoltWintersPrediction
      perInterval: 1
      holtWinters:
        runtimeTuningFetch:
          type: "http"
          timeout: 2500
          http:
            method: "GET"
            url: "http://tuning/holt_winters"
            successCodes:
              - 200
            parameterMode: query
        seasonLength: 6
        storedSeasons: 4
        method: "additive"
    decisionType: "maximum"
```

The method is defined with the name `runtimeTuningFetch`.

The supported methods for the PHPA are:

- Shell scripts
- HTTP requests

The process is as follows:

1. PHPA begins Holt-Winters calculation.
2. The values are initially set to any hardcoded values supplied in the configuration.
3. A runtime tuning configuration has been supplied, using this configuration a method is executed (for example a HTTP
request is sent).
  - This method will provide as input data in JSON that includes the current model, and an array of timestamped
  previous scaling decisions (referred to as `evaluations`, [see below](#request-format)).
  - The response should conform to the expected JSON structure ([see below](#response-format)).
4. If the method execution is successful, and the response is valid, the tuning values are extracted and any provided
values overwrite the hardcoded values.
5. If all required tuning values are provided the tuning values are used to calculate.

The tuning values can be both hardcoded and fetched at runtime, for example the `alpha` value could be calculated at
runtime, and the `beta` and `gamma` values could be hardcoded in configuration:

```yaml
- name: predictiveConfig
  value: |
    models:
    - type: HoltWinters
      name: HoltWintersPrediction
      perInterval: 1
      holtWinters:
        runtimeTuningFetch:
          type: "http"
          timeout: 2500
          http:
            method: "GET"
            url: "http://tuning/holt_winters"
            successCodes:
              - 200
            parameterMode: query
        beta: 0.9
        gamma: 0.9
        seasonLength: 6
        storedSeasons: 4
        method: "additive"
    decisionType: "maximum"
```

Or the values could be provided, and if they are not returned by the external source the hardcoded values will be used
as a backup:

```yaml
- name: predictiveConfig
  value: |
    models:
    - type: HoltWinters
      name: HoltWintersPrediction
      perInterval: 1
      holtWinters:
        runtimeTuningFetch:
          type: "http"
          timeout: 2500
          http:
            method: "GET"
            url: "http://tuning/holt_winters"
            successCodes:
              - 200
            parameterMode: query
        alpha: 0.9
        beta: 0.9
        gamma: 0.9
        seasonLength: 6
        storedSeasons: 4
        method: "additive"
    decisionType: "maximum"
```

If any value is missing, the PHPA will report it as an error (e.g.
`No alpha tuning value provided for Holt-Winters prediction`) and not calculate and scale.

#### Request Format

The data that the external source will recieve will be formatted as:

```json
{
  "model": {
    "type": "HoltWinters",
    "name": "HoltWintersPrediction",
    "perInterval": 1,
    "linear": null,
    "holtWinters": {
      "alpha": null,
      "beta": null,
      "gamma": null,
      "runtimeTuningFetch": {
        "type": "http",
        "timeout": 2500,
        "shell": null,
        "http": {
          "method": "GET",
          "url": "http://tuning/holt_winters",
          "successCodes": [
            200
          ],
          "parameterMode": "query"
        }
      },
      "seasonLength": 6,
      "storedSeasons": 4,
      "method": "additive"
    }
  },
  "evaluations": [
    {
      "id": 1,
      "created": "2020-10-19T19:12:20Z",
      "val": {
        "targetReplicas": 0
      }
    },
    {
      "id": 2,
      "created": "2020-10-19T19:12:40Z",
      "val": {
        "targetReplicas": 0
      }
    },
    {
      "id": 3,
      "created": "2020-10-19T19:13:00Z",
      "val": {
        "targetReplicas": 0
      }
    },
  ]
}
```

This provides information around the model being used, how it is configured, and previous scaling decisions
(`evaluations`). This data could be used to help calculate the tuning values, or it could be ignored.

#### Response Format

The response that the external tuning source should return must be in JSON, and in the following format:

```json
{
  "alpha": <alpha_value>,
  "beta": <beta_value>,
  "gamma": <gamma_value>
}
```

This is a simple JSON structure, for example:

```json
{
  "alpha": 0.9,
  "beta": 0.6,
  "gamma": 0.8
}
```

Each of these values is optional, for example perhaps only the `alpha` and `beta` should be runtime, and `gamma` should
rely on the hardcoded configuration value, this response would be valid:

```json
{
  "alpha": 0.9,
  "beta": 0.6
}
```

For a more detailed example, [see the example in
`/example/dynamic-holt-winters`](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/tree/master/example/dynamic-holt-winters).
