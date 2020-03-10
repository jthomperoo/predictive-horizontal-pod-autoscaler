# Tolerance

The equivalent of the Kubernetes HPA `--horizontal-pod-autoscaler-tolerance` flag can be set by providing the parameter `tolerance` in the autoscaler YAML. Unlike the HPA this does not need to be set as a flag for the kube-controller-manager on the master node, and can be autoscaler specific.  

This option has a default value of `0.1`.  
See the [configuration reference for more details](../../reference/configuration#tolerance).