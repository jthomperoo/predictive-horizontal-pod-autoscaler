# CPU Initialization Period

The equivalent of the Kubernetes HPA `--horizontal-pod-autoscaler-cpu-initialization-period` flag can be set by providing the parameter `cpuInitializationPeriod` in the predicitve config YAML. Unlike the HPA this does not need to be set as a flag for the kube-controller-manager on the master node, and can be autoscaler specific.  

This option is set in seconds.  
This option has a default value of `300` (5 minutes).  
See the [configuration reference for more details](../../reference/configuration#cpuinitializationperiod).