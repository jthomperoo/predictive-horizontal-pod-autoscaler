[![Build](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/workflows/main/badge.svg)](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/actions)
[![codecov](https://codecov.io/gh/jthomperoo/predictive-horizontal-pod-autoscaler/branch/master/graph/badge.svg)](https://codecov.io/gh/jthomperoo/predictive-horizontal-pod-autoscaler)
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
This is a [Custom Pod Autoscaler](https://www.github.com/jthomperoo/custom-pod-autoscaler); building on the Horizontal
Pod Autoscaler functionality to add predictive capabilities by using various statistical methods.

This uses the
[Horizontal Pod Autoscaler Custom Pod Autoscaler](https://www.github.com/jthomperoo/horizontal-pod-autoscaler)
extensively to provide most functionality for the Horizontal Pod Autoscaler parts.

## How does it work?

This project works by calculating the number of replicas a resource should have, then storing these values and using
statistical models against them to produce predictions for the future. These predictions are compared and can be used
instead of the raw replica count calculated by the Horizontal Pod Autoscaler logic.

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

## More information

See the [wiki for more information, such as guides and
references](https://predictive-horizontal-pod-autoscaler.readthedocs.io/en/latest/).

## Developing this project
### Environment
Developing this project requires these dependencies:

* [Go](https://golang.org/doc/install) >= `1.13`
* [Golint](https://github.com/golang/lint)
* [Docker](https://docs.docker.com/install/)

To view docs locally, requires:

* [mkdocs](https://www.mkdocs.org/)

### Commands

* `go mod vendor` - generates a vendor folder.
* `make` - builds the Predictive HPA binary.
* `make docker` - builds the Predictive HPA image.
* `make lint` - lints the code.
* `make beautify` - beautifies the code, must be run to pass the CI
* `make unittest` - runs the unit tests
* `make doc` - hosts the documentation locally, at `127.0.0.1:8000`.
