# Models

## Shared Configuration

All models share these three options:

- **type** - The type of the model, for example 'Linear'.
- **name** - The name of the model, must be unique and not shared by multiple models.
- **perInterval** - The frequency that the model is used to recalculate and store values - tied to the interval as a base unit, with a value of `1` resulting in the model being recalculated every interval, a value of `2` meaning recalculated every other interval, `3` waits for two intervals after every calculation and so on.

All models use `interval` as a base unit, so if the interval is defined as `10000` (10 seconds), the models will base their timings and calculations as multiples of 10 seconds.

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
- **storedValues** - sets up the model to store the past `6` evaluations and to use these for predictions. If there are `> 6` evaluations, the oldest will be removed.

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

- **alpha**, **beta**, **gamma** - these are the smoothing coefficients for level, trend and seasonality respectively, requires tweaking and analysis to be able to optimise. See [here](https://github.com/jthomperoo/holtwinters) or [here](https://grisha.org/blog/2016/01/29/triple-exponential-smoothing-forecasting/) for more details.
- **seasonLength** - the length of a season in base unit intervals, for example if your interval was `10000` (10 seconds), and your repeated season was 60 seconds long, this value would be `6`.
- **storedSeasons** - the number of seasons to store, for example `4`, if there are `>4` seasons stored, the oldest season will be removed.

This is the model in action, taken from the `simple-holt-winters` example:  
![Predicted values overestimating but still fitting actual values](../img/holt_winters_prediction_vs_actual.svg)  
The red value is the predicted values, the blue value is the actual values. From this you can see that the prediction is overestimating, but still pre-emptively scaling - storing more seasons and adjusting alpha, beta and gamma values would reduce the overestimation and produce more accurate results.  

For a more detailed example, [see the example in `/example/simple-holt-winters`](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/tree/master/example/simple-holt-winters).
