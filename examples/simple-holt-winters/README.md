# Simple Holt Winters

This example shows how Holt-Winters can be used to with the Predictive Horizontal Pod Autoscaler (PHPA) to predict
scaling demand based on seasonal data.

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
![Predicted values overestimating but still fitting actual values](../../docs/img/holt_winters_prediction_vs_actual.svg)
From this you can see that the prediction is overestimating, but still pre-emptively scaling - storing more seasons and
adjusting alpha, beta and gamma values would reduce the overestimation and produce more accurate results.

## Requirements

To set up this example and follow the steps listed here you need:

- [kubectl](https://kubernetes.io/docs/tasks/tools/).
- A Kubernetes cluster that kubectl is configured to use - [k3d](https://github.com/rancher/k3d) is good for local
testing.
- [helm](https://helm.sh/docs/intro/install/) to install the PHPA operator.
- [jq](https://stedolan.github.io/jq/) to format some JSON output.

## Usage

If you want to deploy this onto your cluster, you first need to install the Predictive Horizontal Pod Autoscaler
Operator, follow the [installation guide for instructions for installing the
operator](https://predictive-horizontal-pod-autoscaler.readthedocs.io/en/latest/user-guide/installation).

This example was based on the [Horizontal Pod Autoscaler
Walkthrough](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/).

1. Run this command to spin up the app/deployment to manage, called `php-apache`:

```bash
kubectl apply -f deployment.yaml
```

2. Run this command to start the autoscaler, pointing at the previously created deployment:

```bash
kubectl apply -f phpa.yaml
```

3. Run this command to build the load tester image and import it into your Kubernetes cluster:

```bash
docker build -t load-tester load && k3d image import load-tester
```

4. Run this command to deploy the load tester, note the time as it will run for 30 seconds every minute:

```bash
kubectl apply -f load/load.yaml
```

5. Run this command to see the autoscaler working and the log output it produces:

```bash
kubectl logs -l name=predictive-horizontal-pod-autoscaler -f
```

6. Run this command to see the replica history for the autoscaler stored in a configmap and tracked by the autoscaler:

```bash
kubectl get configmap predictive-horizontal-pod-autoscaler-simple-holt-winters-data -o=json | jq -r '.data.data | fromjson | .modelHistories["simple-holt-winters"].replicaHistory[] | .time,.replicas'
```

Every minute the load tester will increase the load on the application we are autoscaling for 30 seconds. The PHPA will
initially without any data just act like a Horizontal Pod Autoscaler and will reactively scale up to meet this demand
as best as it can after the demand has already started. After the load tester has run a couple of times the PHPA will
have built up enough data that it can start to make predictions ahead of time using the Holt Winters model, and it
will start calculating these predictions and proactively scaling up ahead of time to meet demand that it expects based
on the data collected in the past.

## Explanation

This example is split into four parts:

- The deployment to autoscale
- Predictive Horizontal Pod Autoscaler (PHPA)
- Tuning Service
- Load Tester

### Deployment

The deployment to autoscale is a simple service that responds to HTTP requests, it uses the `k8s.gcr.io/hpa-example`
image to return `OK!` to any HTTP GET requests. This deployment will have the number of pods assigned to it scaled up
and down.

### Predictive Horizontal Pod Autoscaler

The PHPA contains some configuration for how the scaling should be applied, the configuration defines how the
autoscaler will act:

```yaml
apiVersion: jamiethompson.me/v1alpha1
kind: PredictiveHorizontalPodAutoscaler
metadata:
  name: simple-holt-winters
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: php-apache
  minReplicas: 1
  maxReplicas: 10
  syncPeriod: 20000
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 30
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 50
  models:
  - type: HoltWinters
    name: simple-holt-winters
    startInterval: 60s
    resetDuration: 5m
    holtWinters:
      alpha: 0.9
      beta: 0.9
      gamma: 0.9
      seasonalPeriods: 6
      storedSeasons: 4
      trend: additive
      seasonal: additive
```

- `scaleTargetRef` is the resource the autoscaler is targeting for scaling.
- `minReplicas` and `maxReplicas` are the minimum and maximum number of replicas the autoscaler can scale the resource
between.
- `syncPeriod` is how frequently this autoscaler will run in milliseconds, so this autoscaler will run every 20000
milliseconds (20 seconds).
- `behavior.scaleDown.stabilizationWindowSeconds` handles how quickly an autoscaler can scale down, ensuring that it
will pick the highest evaluation that has occurred within the last time period described, by default it will pick the
highest evaluation over the past 5 minutes. In this case it will pick the highest evaluation over the past 30 seconds.
- `metrics` defines the metrics that the PHPA should use to scale with, in this example it will try to keep average
CPU utilization at 50% per pod.
- `models` - predictive models to apply.
  - `type` - 'HoltWinters', using a Holt-Winters predictive model.
  - `name` - Unique name of the model.
  - `startInterval` - The model will only apply at the top of the next full minute
  - `resetDuration` - The model's replica history will be cleared out if it's been longer than 5 minutes without any
  data recorded (e.g. if the cluster is turned off).
  - `holtWinters` - Holt-Winters specific configuration.
      * `alpha`, `beta`, `gamma` - these are the smoothing coefficients for level, trend and seasonality
      respectively.
    * `seasonalPeriods` - the length of a season in base unit sync periods, for this example sync period is `20000`
    (20 seconds), and season length is `6`, resulting in a season length of 20 * 6 = 120 seconds = 2 minutes.
    * `storedSeasons` - the number of seasons to store, for this example `4`, if there are more than 4 seasons
    stored, the oldest ones are removed.
    * `trend` - Either `add`/`additive` or `mul`/`multiplicative`, defines the method for the trend element.
    * `seasonal` - Either `add`/`additive` or `mul`/`multiplicative`, defines the method for the seasonal element.

### Tuning Service

The tuning service is a simple Flask service that returns the `alpha`, `beta` and `gamma` values in JSON form (in the
format required by the Holt Winters runtime tuning). It also prints out the values provided to it by the Holt Winters
request, these could be used to help calculate the tuning values.

### Load Tester

This is a simple pod that runs a bash script to send HTTP requests as fast as possible to the `php-apache` deplyoment
being autoscaled to simulate increased load.
