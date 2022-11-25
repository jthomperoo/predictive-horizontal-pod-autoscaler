[![Build](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/workflows/main/badge.svg)](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/actions)
[![go.dev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/jthomperoo/predictive-horizontal-pod-autoscaler)
[![Go Report Card](https://goreportcard.com/badge/github.com/jthomperoo/predictive-horizontal-pod-autoscaler)](https://goreportcard.com/report/github.com/jthomperoo/predictive-horizontal-pod-autoscaler)
[![Documentation Status](https://readthedocs.org/projects/predictive-horizontal-pod-autoscaler/badge/?version=latest)](https://predictive-horizontal-pod-autoscaler.readthedocs.io/en/latest)
[![License](https://img.shields.io/:license-apache-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

# Predictive Horizontal Pod Autoscaler

Predictive Horizontal Pod Autoscalers (PHPAs) are Horizontal Pod Autoscalers (HPAs) with extra predictive capabilities
baked in, allowing you to apply statistical models to the results of HPA calculations to make proactive scaling
decisions.

This extensively uses the the [jthomperoo/k8shorizmetrics](https://github.com/jthomperoo/k8shorizmetrics) library
to gather metrics and to evaluate them as the Kubernetes Horizontal Pod Autoscaler does.

# Why would I use it?

PHPAs lets you choose models and fine tune them in order to predict how many replicas a resource should have,
preempting events such as regular, repeated high load. This allows for proactive rather than simply reactive scaling
that can make intelligent ahead of time decisions.

# What systems would need it?

Systems that have predictable changes in load, for example; if over a 24 hour period the load on a resource is
generally higher between 3pm and 5pm - with enough data and use of correct models and tuning the autoscaler could
predict this and preempt the load, increasing responsiveness of the system to changes in load. This could be useful for
handling different userbases across different timezones, or understanding that if a load is rapidly increasing we can
prempt the load by predicting replica counts.

## Features

* Functionally identical to Horizontal Pod Autoscaler for calculating replica counts without prediction.
* Choice of statistical models to apply over Horizontal Pod Autoscaler replica counting logic.
    * Holt-Winters Smoothing
    * Linear Regression
* Allows customisation of Kubernetes autoscaling options without master node access. Can therefore work on managed
solutions such as EKS or GCP.
    * CPU Initialization Period.
    * Downscale Stabilization.
    * Sync Period.

## How does it work?

This project works by calculating the number of replicas a resource should have, then storing these values and using
statistical models against them to produce predictions for the future. These predictions are compared and can be used
instead of the raw replica count calculated by the Horizontal Pod Autoscaler logic.

## What does a Predictive Horizontal Pod Autoscaler look like?

PHPAs are designed to be as similar in configuration to Horizontal Pod Autoscalers as possible, with extra
configuration options.

PHPAs have their own custom resource:

```yaml
apiVersion: jamiethompson.me/v1alpha1
kind: PredictiveHorizontalPodAutoscaler
metadata:
  name: simple-linear
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: php-apache
  minReplicas: 1
  maxReplicas: 10
  downscaleStabilization: 0
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          averageUtilization: 50
          type: Utilization
  models:
    - type: Linear
      name: simple-linear
      linear:
        lookAhead: 10000
        historySize: 6
```

This PHPA acts like a Horizontal Pod Autoscaler and autoscales to try and keep the target resource's CPU utilization at
50%, but with the extra predictive layer of a linear regression model applied to the results.

## Installation

The operator for managing Predictive Horizontal Pod Autoscalers can be installed using Helm:

```bash
VERSION=v0.11.2
HELM_CHART=predictive-horizontal-pod-autoscaler-operator
helm install ${HELM_CHART} https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/releases/download/${VERSION}/predictive-horizontal-pod-autoscaler-${VERSION}.tgz
```

## Quick start

Check out the [getting started
guide](https://predictive-horizontal-pod-autoscaler.readthedocs.io/en/latest/user-guide/getting-started/) and the
[examples](./examples/) for ways to use Predictive Horizontal Pod Autoscalers.

## More information

See the [wiki for more information, such as guides and
references](https://predictive-horizontal-pod-autoscaler.readthedocs.io/en/latest/).

See the [`examples/` directory](./examples) for working code samples.

## Developing this project

Developing this project requires these dependencies:

* [Go](https://golang.org/doc/install) >= `1.18`
* [Python](https://www.python.org/downloads/) == `3.8.x`
* [Helm](https://helm.sh/) == `3.9.x`

Any Python dependencies must be installed by running:

```bash
pip install -r requirements-dev.txt
```

It is recommended to test locally using a local Kubernetes managment system, such as
[k3d](https://github.com/rancher/k3d) (allows running a small Kubernetes cluster locally using Docker).

You can deploy a PHPA example (see the [`examples/` directory](./examples) for choices) to test your changes.

### Commands

* `make run` - runs the PHPA locally against the cluster configured in your kubeconfig file.
* `make docker` - builds the PHPA image.
* `make lint` - lints the code.
* `make format` - beautifies the code, must be run to pass the CI.
* `make test` - runs the unit tests.
* `make doc` - hosts the documentation locally, at `127.0.0.1:8000`.
* `make coverage` - opens up any generated coverage reports in the browser.
