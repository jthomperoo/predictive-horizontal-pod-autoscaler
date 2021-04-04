[![Build](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/workflows/main/badge.svg)](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/actions)
[![codecov](https://codecov.io/gh/jthomperoo/predictive-horizontal-pod-autoscaler/branch/master/graph/badge.svg)](https://codecov.io/gh/jthomperoo/predictive-horizontal-pod-autoscaler)
[![go.dev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/jthomperoo/predictive-horizontal-pod-autoscaler)
[![Go Report Card](https://goreportcard.com/badge/github.com/jthomperoo/predictive-horizontal-pod-autoscaler)](https://goreportcard.com/report/github.com/jthomperoo/predictive-horizontal-pod-autoscaler)
[![Documentation Status](https://readthedocs.org/projects/predictive-horizontal-pod-autoscaler/badge/?version=latest)](https://predictive-horizontal-pod-autoscaler.readthedocs.io/en/latest)
[![License](https://img.shields.io/:license-apache-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)
# Predictive Horizontal Pod Autoscaler

Visit the GitHub repository at <https://www.github.com/jthomperoo/predictive-horizontal-pod-autoscaler> to see examples,
raise issues, and to contribute to the project.

# What is it?

This is a [Custom Pod Autoscaler](https://www.github.com/jthomperoo/custom-pod-autoscaler); aiming to have identical
functionality to the Horizontal Pod Autoscaler, however with added predictive elements using statistical models.

# Why would I use it?

This autoscaler lets you choose models and fine tune them in order to predict how many replicas a resource should have,
preempting events such as regular, repeated high load.

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
    * Initial Readiness Delay.
* Runs in Kubernetes as a standard Pod.
