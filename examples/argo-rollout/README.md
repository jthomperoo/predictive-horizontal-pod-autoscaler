# Argo Rollout

> Note: this feature is only available in the Predictive Horizontal Pod Autoscaler v0.9.0 and above.

> Note: this example requires using a Custom Pod Autoscaler Operator `v1.2.0` and above.

## Overview

This example is showing how to target an [Argo Rollout](https://argoproj.github.io/argo-rollouts/), other than that it
is identical to the [simple-linear example](../simple-linear).

This example was based on the [Horizontal Pod Autoscaler
Walkthrough](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/).

## Usage

### Enable Argo Rollouts

Using this requires Argo Rollouts to be enabled on your Kubernetes cluster, [follow this guide to set up Argo Rollouts
on your cluster](https://argoproj.github.io/argo-rollouts/installation/)

### Enable CPAs

Using this CPA requires CPAs to be enabled on your Kubernetes cluster, [follow this guide to set up CPAs on your
cluster](https://github.com/jthomperoo/custom-pod-autoscaler-operator#installation).

### Deploy an Argo Rollout to Manage

First a rollout needs to be deployed that the CPA can manage, you can deploy with the following command:

```bash
kubectl apply -f rollout.yaml
```

You can check if the rollout is deployed by running this command:

```bash
kubectl argo rollouts get rollout php-apache
```

You should see that the rollout is set up to initially have only `1` replica.

### Deploy the Predictive Horizontal Pod Autoscaler

Deploy the PHPA with the following command:

```bash
kubectl apply -f phpa.yaml
```

### Increase the CPU Load

Run the following command to increase the CPU load by making HTTP requests, this may take a while:

```bash
kubectl run -i --tty load-generator --rm --image=busybox --restart=Never -- /bin/sh -c "while sleep 0.01; do wget -q -O- http://php-apache; done"
```

You can stop increasing the load by using Ctrl+C.

You should see the number of replicas of the `php-apache` rollout increase.

## Configuration

This example targets the rollout using a `scaleTargetRef` configured using the following:

```yaml
scaleTargetRef:
  apiVersion: argoproj.io/v1alpha1
  kind: Rollout
  name: php-apache
```

This example also requests that the CPA is provisioned with the role required by providing the following option:

```yaml
roleRequiresArgoRollouts: true
```
