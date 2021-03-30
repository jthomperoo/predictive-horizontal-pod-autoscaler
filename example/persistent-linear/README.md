# Persistent Linear Example

## Overview

This example is showing a predictive horizontal pod autoscaler using a linear regression model, with persistent volume storage, so if the scaler is deleted the data will persist.  

## Usage
If you want to deploy this onto your cluster, you first need to install the [Custom Pod Autoscaler Operator](https://github.com/jthomperoo/custom-pod-autoscaler-operator), follow the [installation guide for instructions for installing the operator](https://github.com/jthomperoo/custom-pod-autoscaler-operator/blob/master/INSTALL.md).  

This example was based on the [Horizontal Pod Autoscaler Walkthrough](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/).  

You must first set up a Persistent volume on your cluster, this example will assume you are using Minikube.
1. Open a shell to the node `minikube ssh`.
2. Create a data dir `sudo mkdir /mnt/data`.
3. Exit the shell `exit`.
Now your persistent volume is set up, you can set up the autoscaler.

1. Use `kubectl apply -f deployment.yaml` to spin up the app/deployment to manage, called `php-apache`.
2. Use `kubectl apply -f phpa.yaml` to start the autoscaler, pointing at the previously created deployment.
3. Use `kubectl logs simple-linear-example --follow` to see the autoscaler working and the log output it produces.
4. Increase the load with: `kubectl run --generator=run-pod/v1 -it --rm load-generator --image=busybox /bin/sh`
    * Once it has loaded, run this command to create load `while true; do wget -q -O- http://php-apache.default.svc.cluster.local; done`
5. Watch as the number of replicas increases.
6. Use `kubectl exec -it simple-linear-example sqlite3 /store/predictive-horizontal-pod-autoscaler.db 'SELECT * FROM evaluation;'` to see the evaluations stored locally and tracked by the autoscaler.

## Explained

The example has some configuration
```yaml
config: 
  - name: minReplicas
    value: "1"
  - name: maxReplicas
    value: "5"
  - name: predictiveConfig
    value: |
      models:
      - type: Linear
        name: LinearPrediction
        perInterval: 1
        linear:
          lookAhead: 10
          storedValues: 6
      decisionType: "maximum"
      metrics:
      - type: Resource
        resource:
          name: cpu
          target:
            averageUtilization: 50
            type: Utilization
  - name: interval
    value: "10000"
  - name: downscaleStabilization
    value: "30"
```
The `minReplicas`, `maxReplicas` and `interval` are Custom Pod Autoscaler options, setting minimum and maximum replicas, and the time interval inbetween each autoscale being run, i.e. the autoscaler checks every 10 seconds.  
The `downscaleStabilization` is also a Custom Pod Autoscaler option, in this case changing the `downscaleStabilization` from the default 300 seconds (5 minutes), to 30 seconds. The `downscaleStabilization` option handles how quickly an autoscaler can scale down, ensuring that it will pick the highest evaluation that has occurred within the last time period described, in this case it will pick the highest evaluation over the past 30 seconds.  
The `predictiveConfig` option is the Predictive Horizontal Pod Autoscaler options, detailing a linear regression model that runs on every interval, looking 10 seconds ahead, keeping track of the past 6 replica values in order to predict the next result, and the `decisionType` is maximum, which if there were multiple models provided would mean that the PHPA would use the one with the highest replica count; there are two other options, `mean` and `minimum`. The `metrics` option is a Horizontal Pod Autoscaler option, targeting CPU utilisation.  

