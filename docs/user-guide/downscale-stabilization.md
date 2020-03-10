# Downscale Stabilization

The equivalent of the Kubernetes HPA `--horizontal-pod-autoscaler-downscale-stabilization` flag can be set by providing the parameter `downscaleStabilization` in the autoscaler YAML. Unlike the HPA this does not need to be set as a flag for the kube-controller-manager on the master node, and can be autoscaler specific.  

This option is set in seconds.  
This option is part of the Custom Pod Autoscaler Framework, for more [information please view the Custom Pod Autoscaler Wiki configuration reference](https://custom-pod-autoscaler.readthedocs.io/en/latest/reference/configuration/#downscalestabilization).