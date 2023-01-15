# Configuration

Predictive Horizontal Pod Autoscalers have a number of configuration options available.

## minReplicas

```yaml
minReplicas: 2
```

The lower limit for the number of replicas to which the autoscaler can scale down to. `minReplicas` is allowed to be 0
if at least one Object or External metric is configured. Scaling is active as long as at least one metric value is
available.

Default value: `1`.

## maxReplicas

```yaml
maxReplicas: 15
```

The upper limit for the number of replicas to which the autoscaler can scale up.
It cannot be less than minReplicas.

Default value: `10`.

## syncPeriod

```yaml
syncPeriod: 10000
```

Equivalent to `--horizontal-pod-autoscaler-sync-period`; the frequency with which the PHPA calculates replica counts and
scales in milliseconds.

Set in milliseconds.

Default value: `15000` (15 seconds).

Set in milliseconds.

## cpuInitializationPeriod

```yaml
cpuInitializationPeriod: 150
```

Equivalent to `--horizontal-pod-autoscaler-cpu-initialization-period`; the period after pod start when CPU samples
might be skipped.

Set in seconds.

Default value: `300` (5 minutes).

## initialReadinessDelay

```yaml
initialReadinessDelay: 45
```

Equivalent to `--horizontal-pod-autoscaler-initial-readiness-delay`; the period after pod start during which readiness
changes will be treated as initial readiness.

Set in seconds.

Default value: `30` (30 seconds).

## tolerance

```yaml
tolerance: 0.25
```

Equivalent to `--horizontal-pod-autoscaler-tolerance`; the minimum change (from 1.0) in the desired-to-actual metrics
ratio for the horizontal pod autoscaler to consider scaling.

Default value: `0.1`.

## decisionType

```yaml
decisionType: mean
```

Decider on which evaluation to pick if there are multiple models provided.

Possible values:

- **maximum** - pick the highest evaluation of the models.
- **minimum** - pick the lowest evaluation of the models.
- **mean** - calculate the mean number of replicas (rounded to nearest integer) between the models.
- **median** - calculate the median number of replicas between the models.

Default value: `maximum`.

## behavior

Scaling behavior to apply.

Intended to be feature equivalent to Kubernetes HPA behavior.

See the [Horizontal Pod Autoscaler docs
here](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#configurable-scaling-behavior).

## models

List of statistical models to apply.
See [the models section for details](../../user-guide/models).

## metrics

List of metrics to target for evaluating replica counts.
See [the metrics section for details](../../user-guide/metrics).
