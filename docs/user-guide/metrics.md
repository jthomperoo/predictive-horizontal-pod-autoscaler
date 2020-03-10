# Metrics

Deciding which metrics to use is done by using `MetricSpecs`, which are a key part of HPAs, and look like this:
```yaml
- type: Resource
  resource:
    name: cpu
    target:
      type: Utilization
      averageUtilization: 50
```

To send these specs to the Predictive HPA, add a config option called `metrics` to the CPA, with a multiline string containing the metric list. For example:
```yaml
- name: predictiveConfig
  value: |
    ...
    metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          averageUtilization: 50
          type: Utilization
```

This allows porting over existing Kubernetes HPA metric configurations to the Predictive Horizontal Pod Autoscaler.  
Equivalent to K8s HPA metric specs; which are [demonstrated in this HPA walkthrough](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/#autoscaling-on-multiple-metrics-and-custom-metrics).  
Can hold multiple values as it is an array.