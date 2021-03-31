# Simple Holt-Winters Example

## Overview

This example is showing a predictive horizontal pod autoscaler using holt-winters exponential smoothing time series
prediction, with no persistent storage, so if the scaler is deleted the data will not persist.

This uses the Holt-Winters time series prediction method, which allows for defining seasons to predict how to scale.
For example, defining a season as 24 hours,a deployment regularly has a higher CPU load between 3pm and 5pm, the model
will gather data and once enough seasons have been gathered, will make predictions based on its knowledge of CPU load
being higher between 3pm and 5pm, leading to pre-emptive scaling that will keep latency down and keep the system ready
and responsive.

This example is a smaller scale of the example described above, with an interval time of 20 seconds, and a season of
length 6 (6 * 20 = 120 seconds = 2 minutes). The example will store up to 4 previous seasons to make predictions with.
The example includes a load tester, which runs for 30 seconds every minute.

This is the result of running the example plotted, with red values being predicted values and blue values being actual
values:
![Predicted values overestimating but still fitting actual values](./graph.svg)
From this you can see that the prediction is overestimating, but still pre-emptively scaling - storing more seasons and
adjusting alpha, beta and gamma values would reduce the overestimation and produce more accurate results.

## Usage
If you want to deploy this onto your cluster, you first need to install the [Custom Pod Autoscaler
Operator](https://github.com/jthomperoo/custom-pod-autoscaler-operator), follow the [installation guide for
instructions for installing the
operator](https://github.com/jthomperoo/custom-pod-autoscaler-operator/blob/master/INSTALL.md).

This example was based on the [Horizontal Pod Autoscaler Walkthrough](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/). This example assumes you are using Minikube, or working out of the same Docker
registry as your Kubernetes cluster.

1. Use `kubectl apply -f deployment.yaml` to spin up the app/deployment to manage, called `php-apache`.
2. Use `kubectl apply -f phpa.yaml` to start the autoscaler, pointing at the previously created deployment.
3. Build the load tester.
  - Point to the cluster's Docker registry (e.g. for this Minikube) `eval $(minikube docker-env)`.
  - Build the load tester image `docker build -t load-tester load`
4. Deploy the load tester, note the time as it will run for 30 seconds every minute `kubectl apply -f load/load.yaml`.
5. Use `kubectl logs simple-linear-example --follow` to see the autoscaler working and the log output it produces.

## Explained

The example has some configuration:
```yaml
config:
  - name: minReplicas
    value: "1"
  - name: maxReplicas
    value: "10"
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
          seasonalPeriods: 6
          storedSeasons: 4
          method: "additive"
      decisionType: "maximum"
      metrics:
      - type: Resource
        resource:
          name: cpu
          target:
            type: Utilization
            averageUtilization: 50
  - name: interval
    value: "20000"
  - name: startTime
    value: "60000"
  - name: downscaleStabilization
    value: "30"
```
- **minReplicas**, **maxReplicas**, **startTime** and **interval** - Custom Pod Autoscaler options, setting minimum and
maximum replicas, the starting time - for this example will start at the nearest full minute, and the time interval
inbetween each autoscale being run, i.e. the autoscaler checks every 20 seconds.
- **downscaleStabilization** is also a Custom Pod Autoscaler option, in this case changing the `downscaleStabilization`
from the default 300 seconds (5 minutes), to 30 seconds. The `downscaleStabilization` option handles how quickly an
autoscaler can scale down, ensuring that it will pick the highest evaluation that has occurred within the last time
period described, in this case it will pick the highest evaluation over the past 30 seconds.
- **predictiveConfig** - configuration of the predictive elements.
  * **models** - predictive models to apply.
    - **type** - 'HoltWinters', using a Holt-Winters predictive model.
    - **name** - Unique name of the model.
    - **holtWinters** - Holt-Winters specific configuration.
      * **alpha**, **beta**, **gamma** - these are the smoothing coefficients for level, trend and seasonality
      respectively.
      * **seasonalPeriods** - the length of a season in base unit intervals, for this example interval is `20000`
      (20 seconds), and season length is `6`, resulting in a season length of 20 * 6 = 120 seconds = 2 minutes.
      * **storedSeasons** - the number of seasons to store, for this example `4`, if there are more than 4 seasons
      stored, the oldest ones are removed.
      * **method** - the Holt-Winters method to use, either `additive` or `multiplicative`.
  * **decisionType** - strategy for resolving multiple models, either `maximum`, `minimum` or `mean`, in this case
  `maximum`, meaning take the highest predicted value.
  * **metrics** - Horizontal Pod Autoscaler option, targeting 50% CPU utilisation.
