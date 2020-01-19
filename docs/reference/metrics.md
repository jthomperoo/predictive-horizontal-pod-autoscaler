# Metrics

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
This works in the same way as metric specs for Horizontal Pod Autoscalers, so any spec configuration could just be copied from an existing Horizontal Pod Autoscaler and put in here.  

See the [Horizontal Pod Autoscaler as a Custom Pod Autoscaler](https://www.github.com/jthomperoo/horizontal-pod-autoscaler) for more information.  
