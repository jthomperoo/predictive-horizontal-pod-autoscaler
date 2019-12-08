[![Build](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/workflows/main/badge.svg)](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/actions)
[![codecov](https://codecov.io/gh/jthomperoo/predictive-horizontal-pod-autoscaler/branch/master/graph/badge.svg)](https://codecov.io/gh/jthomperoo/predictive-horizontal-pod-autoscaler)
[![GoDoc](https://godoc.org/github.com/jthomperoo/predictive-horizontal-pod-autoscaler?status.svg)](https://godoc.org/github.com/jthomperoo/predictive-horizontal-pod-autoscaler)
[![Go Report Card](https://goreportcard.com/badge/github.com/jthomperoo/predictive-horizontal-pod-autoscaler)](https://goreportcard.com/report/github.com/jthomperoo/predictive-horizontal-pod-autoscaler)
[![License](http://img.shields.io/:license-apache-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)
# Predictive Horizontal Pod Autoscaler
# Very early pre-release
This is a Custom Pod Autoscaler; aiming to have identical functionality to the Horizontal Pod Autoscaler, however with added predictive elements.  

This uses the [Horizontal Pod Autoscaler Custom Pod Autoscaler](https://www.github.com/jthomperoo/horizontal-pod-autoscaler) extensively to provide most functionality for the Horizontal Pod Autoscaler parts.  

This runs as a [Custom Pod Autoscaler](https://www.github.com/jthomperoo/custom-pod-autoscaler), which allows creation of custom autoscalers to run in a Kubernetes cluster; this project was made as an example of what is possible.

## How does it work?

This project works by calculating the number of replicas a resource should have, then storing these values and using statistical models against them to produce predictions for the future. These predictions are compared and can be used instead of the raw replica count calculated by the HPA logic.

## Usage

If you want to deploy this onto your cluster, you first need to install the [Custom Pod Autoscaler Operator](https://github.com/jthomperoo/custom-pod-autoscaler-operator), follow the [installation guide for instructions for installing the operator](https://github.com/jthomperoo/custom-pod-autoscaler-operator/blob/master/INSTALL.md).  

Once the Custom Pod Autoscaler Operator is installed, you can now use this project, see the `/example` folder for some samples.  

### Metrics
You can specify which metrics to scale on using the same YAML you would use for the Horizontal Pod Autoscaler; by putting it into the `metrics` option:
```yaml
- name: metrics
    value: |
    - type: Resource
      resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 50
```
See the [Horizontal Pod Autoscaler Custom Pod Autoscaler](https://www.github.com/jthomperoo/horizontal-pod-autoscaler) for more information.

### Models
#### Linear Regression
At the minute there is only a single type of predictive model available; Linear Regression. There are plans to add in more useful and complex prediction models, such as ARIMA or Holt-Winters. You can specify the model to use in the `predictiveConfig` option in the deployment YAML:
```yaml
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
```
For a more detailed example, see either example in the `/example` folder.

## Developing this project
### Environment
Developing this project requires these dependencies:

* Go >= 1.13
* Golint
* Docker

### Commands

* `make` - builds the Predictive HPA binary.
* `make docker` - builds the Predictive HPA image.
* `make lint` - lints the code.
* `make unittest` - runs the unit tests
* `make vendor` - generates a vendor folder.