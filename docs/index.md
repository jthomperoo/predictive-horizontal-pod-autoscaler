[![Build](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/workflows/main/badge.svg)](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/actions)
[![codecov](https://codecov.io/gh/jthomperoo/predictive-horizontal-pod-autoscaler/branch/master/graph/badge.svg)](https://codecov.io/gh/jthomperoo/predictive-horizontal-pod-autoscaler)
[![GoDoc](https://godoc.org/github.com/jthomperoo/predictive-horizontal-pod-autoscaler?status.svg)](https://godoc.org/github.com/jthomperoo/predictive-horizontal-pod-autoscaler)
[![Go Report Card](https://goreportcard.com/badge/github.com/jthomperoo/predictive-horizontal-pod-autoscaler)](https://goreportcard.com/report/github.com/jthomperoo/predictive-horizontal-pod-autoscaler)
[![Documentation Status](https://readthedocs.org/projects/predictive-horizontal-pod-autoscaler/badge/?version=latest)](https://predictive-horizontal-pod-autoscaler.readthedocs.io/en/latest)
[![License](http://img.shields.io/:license-apache-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)
# Predictive Horizontal Pod Autoscaler

# What is it?

This is a [Custom Pod Autoscaler](https://www.github.com/jthomperoo/custom-pod-autoscaler); 
aiming to have identical functionality to the Horizontal Pod Autoscaler, however with added 
predictive elements using statistical models.  

This uses the 
[Horizontal Pod Autoscaler reimplemented as a Custom Pod Autoscaler](https://www.github.com/jthomperoo/horizontal-pod-autoscaler) 
extensively to provide most functionality for the Horizontal Pod Autoscaler parts.  

# Why would I use it?

This autoscaler lets you choose models and fine tune them in order to predict how many replicas a 
resource should have, preempting events such as regular, repeated high load. 

# What systems would need it?

Systems that have predictable changes in load, for example; if over a 24 hour period the load on a 
resource is generally higher between 3pm and 5pm - with enough data and use of correct models and 
tuning the autoscaler could predict this and preempt the load, increasing responsiveness of the 
system to changes in load. This could be useful for handling different userbases across different 
timezones, or understanding that if a load is rapidly increasing we can prempt the load by 
predicting replica counts.

# How does it work?

This project works by calculating the number of replicas a resource should have, then storing 
these values and using statistical models against them to produce predictions for the future. 
These predictions are compared and can be used instead of the raw replica count calculated by 
the Horizontal Pod Autoscaler logic.
