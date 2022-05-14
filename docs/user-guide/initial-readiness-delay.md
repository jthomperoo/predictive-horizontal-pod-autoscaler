# Initial Readiness Delay

The equivalent of the Kubernetes HPA `--horizontal-pod-autoscaler-initial-readiness-delay` flag can be set by providing
the parameter `initialReadinessDelay` in the predicitve config YAML. Unlike the HPA this does not need to be set as a
flag for the kube-controller-manager on the master node, and can be autoscaler specific.

This option is set in seconds.
This option has a default value of `30` (30 seconds).
See the [configuration reference for more details](../../reference/configuration#initialreadinessdelay).
