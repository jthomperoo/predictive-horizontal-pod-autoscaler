# Dynamic Holt Winters

This example shows how Holt-Winters can be used to with the Predictive Horizontal Pod Autoscaler to predict scaling
demand based on seasonal data. This example is described as *dynamic* as it fetches it's tuning values from an external
source at runtime, allowing these values to be dynamically calculated at runtime rather than being hardcoded.

This example specifically uses a HTTP request to a tuning service to fetch the `alpha`, `beta` and `gamma` Holt Winters
tuning values at runtime.

## Usage

If you want to deploy this onto your cluster, you first need to install the [Custom Pod Autoscaler
Operator](https://github.com/jthomperoo/custom-pod-autoscaler-operator), follow the [installation guide for
instructions for installing the
operator](https://github.com/jthomperoo/custom-pod-autoscaler-operator/blob/master/INSTALL.md).

This example was based on the [Horizontal Pod Autoscaler
Walkthrough](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/). This example
assumes you are using Minikube, or working out of the same Docker registry as your Kubernetes cluster.

1. Use `kubectl apply -f deployment.yaml` to spin up the app/deployment to manage, called `php-apache`.
2. Build the tuning service.
  - Point to the cluster's Docker registry (e.g. for this Minikube) `eval $(minikube docker-env)`.
  - Build the load tester image `docker build -t tuning tuning`
4. Deploy the tuning service `kubectl apply -f tuning/tuning.yaml`.
5. Use `kubectl apply -f phpa.yaml` to start the autoscaler, pointing at the previously created deployment.
6. Build the load tester.
  - Point to the cluster's Docker registry (e.g. for this Minikube) `eval $(minikube docker-env)`.
  - Build the load tester image `docker build -t load-tester load`
7. Deploy the load tester, note the time as it will run for 30 seconds every minute `kubectl apply -f load/load.yaml`.
8. Use `kubectl logs simple-linear-example --follow` to see the autoscaler working and the log output it produces.
9. Use `kubectl logs tuning --follow` to see the logs of the tuning service, it will report any time it is queried and
it will print the value provided to it.

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

The PHPA part of the example is just a PHPA that contains some configuration for how the scaling should be applied,
the configuration defines how the autoscaler will act:

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
      * **runtimeTuningFetch** - This is a [method] that is used to dynamically fetch the `alpha`, `beta` and `gamma`
      values at runtime, in this example it is using a `HTTP` request to `http://tuning/holt_winters`.
      * **seasonLength** - the length of a season in base unit intervals, for this example interval is `20000`
      (20 seconds), and season length is `6`, resulting in a season length of 20 * 6 = 120 seconds = 2 minutes.
      * **storedSeasons** - the number of seasons to store, for this example `4`, if there are more than 4 seasons
      stored, the oldest ones are removed.
      * **method** - the Holt-Winters method to use, either `additive` or `multiplicative`.
  * **decisionType** - strategy for resolving multiple models, either `maximum`, `minimum` or `mean`, in this case
  `maximum`, meaning take the highest predicted value.
  * **metrics** - Horizontal Pod Autoscaler option, targeting 50% CPU utilisation.

### Tuning Service

The tuning service is a simple Flask service that returns the `alpha`, `beta` and `gamma` values in JSON form (in the
format required by the Holt Winters runtime tuning). It also prints out the values provided to it by the Holt Winters
request, these could be used to help calculate the tuning values.

### Load Tester

This is a simple pod that runs a bash script to send HTTP requests as fast as possible to the `php-apache` deplyoment
being autoscaled to simulate increased load.
