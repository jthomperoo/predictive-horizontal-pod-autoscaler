# Downscale Stabilization

The equivalent of the Kubernetes HPA `--horizontal-pod-autoscaler-sync-period` flag can be set by providing the
parameter `interval` config option in the autoscaler YAML. Unlike the HPA this does not need to be set as a flag for
the kube-controller-manager on the master node, and can be autoscaler specific.

Interval can be set at deploy time; and are configuration options that are part of the [Custom Pod Autoscaler
Framework](https://custom-pod-autoscaler.readthedocs.io/en/latest).

This option is set in milliseconds.
This option has a default value of `15000` (15 seconds).
This option is part of the Custom Pod Autoscaler Framework, for more [information please view the Custom Pod Autoscaler
Wiki configuration reference](https://custom-pod-autoscaler.readthedocs.io/en/latest/reference/configuration/#interval).
