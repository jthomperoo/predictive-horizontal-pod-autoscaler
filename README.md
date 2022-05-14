[![Build](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/workflows/main/badge.svg)](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/actions)
[![go.dev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/jthomperoo/predictive-horizontal-pod-autoscaler)
[![Go Report Card](https://goreportcard.com/badge/github.com/jthomperoo/predictive-horizontal-pod-autoscaler)](https://goreportcard.com/report/github.com/jthomperoo/predictive-horizontal-pod-autoscaler)
[![Documentation Status](https://readthedocs.org/projects/predictive-horizontal-pod-autoscaler/badge/?version=latest)](https://predictive-horizontal-pod-autoscaler.readthedocs.io/en/latest)
[![License](https://img.shields.io/:license-apache-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

<p>This project is supported by:</p>
<p>
  <a href="https://www.digitalocean.com/">
    <img src="https://opensource.nyc3.cdn.digitaloceanspaces.com/attribution/assets/SVG/DO_Logo_horizontal_blue.svg" width="201px">
  </a>
</p>

# Predictive Horizontal Pod Autoscaler

This is a [Custom Pod Autoscaler](https://www.github.com/jthomperoo/custom-pod-autoscaler); aiming to have identical
functionality to the Horizontal Pod Autoscaler, however with added predictive elements using statistical models.

This extensively uses the the [jthomperoo/k8shorizmetrics](https://github.com/jthomperoo/k8shorizmetrics) library
to gather metrics and to evaluate them as the Kubernetes Horizontal Pod Autoscaler does.

# Why would I use it?

This autoscaler lets you choose models and fine tune them in order to predict how many replicas a resource should have,
preempting events such as regular, repeated high load.

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
    * Initial Readiness Delay.
* Runs in Kubernetes as a standard Pod.

## How does it work?

This project works by calculating the number of replicas a resource should have, then storing these values and using
statistical models against them to produce predictions for the future. These predictions are compared and can be used
instead of the raw replica count calculated by the Horizontal Pod Autoscaler logic.

## More information

See the [wiki for more information, such as guides and
references](https://predictive-horizontal-pod-autoscaler.readthedocs.io/en/latest/).

See the [`examples/` directory](./examples) for working code samples.

## Developing this project

### Environment

Developing this project requires these dependencies:

* [Go](https://golang.org/doc/install) >= `1.17`
* [staticcheck](https://staticcheck.io/docs/getting-started/) == `v0.3.0 (2022.1)`
* [Docker](https://docs.docker.com/install/)
* [Python](https://www.python.org/downloads/) == `3.8.5`

Any Python dependencies must be installed by running:

```bash
pip install -r requirements-dev.txt
```

It is recommended to test locally using a local Kubernetes managment system, such as
[k3d](https://github.com/rancher/k3d) (allows running a small Kubernetes cluster locally using Docker).

Once you have
a cluster available, you should install the [Custom Pod Autoscaler Operator
(CPAO)](https://github.com/jthomperoo/custom-pod-autoscaler-operator/blob/master/INSTALL.md)
onto the cluster to let you install the PHPA.

With the CPAO installed you can install your development builds of the PHPA onto the cluster by building the image
locally, and then pushing the image to the K8s cluster's registry (to do that with k3d you can use the
`k3d image import` command).

Finally you can deploy a PHPA example (see the [`examples/` directory](./examples) for choices) to test your changes.

> Note that the examples generally use `ImagePullPolicy: Always`, you may need to change this to
> `ImagePullPolicy: IfNotPresent` to use your local build.

### Commands

* `make` - builds the Predictive HPA binary.
* `make docker` - builds the Predictive HPA image.
* `make lint` - lints the code.
* `make format` - beautifies the code, must be run to pass the CI.
* `make test` - runs the unit tests.
* `make doc` - hosts the documentation locally, at `127.0.0.1:8000`.
* `make coverage` - opens up any generated coverage reports in the browser.
