[![Build](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/workflows/main/badge.svg)](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/actions)
[![go.dev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/jthomperoo/predictive-horizontal-pod-autoscaler)
[![Go Report Card](https://goreportcard.com/badge/github.com/jthomperoo/predictive-horizontal-pod-autoscaler)](https://goreportcard.com/report/github.com/jthomperoo/predictive-horizontal-pod-autoscaler)
[![Documentation Status](https://readthedocs.org/projects/predictive-horizontal-pod-autoscaler/badge/?version=latest)](https://predictive-horizontal-pod-autoscaler.readthedocs.io/en/latest)
[![License](https://img.shields.io/:license-apache-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

# Predictive Horizontal Pod Autoscaler

Predictive Horizontal Pod Autoscalers (PHPAs) are Horizontal Pod Autoscalers (HPAs) with extra predictive capabilities,
allowing you to autoscale using statistical models for ahead of time predictions.

## Why would I use it?

PHPAs can better scaling results by making proactive decisions to scale up ahead of demand, meaning that a
resource does not have to wait for performance to degrade before autoscaling kicks in.

## What systems would need it?

Any systems that have regular/predictable demand peaks/troughs.

Some use cases:

* A service that sees demand peak between 3pm and 5pm every week day, this is a regular and predictable load which
could be pre-empted.
* A service which sees a surge in demand at 12pm every day for 10 minutes, this is such a short time interval that
by the time a regular HPA made the decision to scale up there could already be major performance/availablity issues.

PHPAs are not a silver bullet, and require tuning using real data for there to be any benefits of using it. A poorly
tuned PHPA could easily end up being worse than a normal HPA.

## How does it work?

This project works by doing the same calculations as the Horizontal Pod Autoscaler does to determine how many replicas
a resource should have, then applies statistical models against the calculated replica count and the replica history.

## Supported Kubernetes versions

The minimum Kubernetes version the autoscaler can run on is `v1.23` because it relies on the `autoscaling/v2` API which
was only available in `v1.23` and above.

The autoscaler is only tested against the latest Kubernetes version - if there are bugs that affect older Kubernetes
versions we will try to fix them, but there is no guarantee of support.

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
