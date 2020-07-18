# Predictive Configuration

Beyond specifying models, other configuration options can be set in the `predictiveConfig` YAML:

## Providing Predictive Configuration

Predictive Configuration is provided through environment variables, which can be supplied through the Custom Pod Autoscaler YAML shorthand:

```yaml
apiVersion: custompodautoscaler.com/v1
kind: CustomPodAutoscaler
metadata:
  name: simple-linear-example
spec:
  template:
    spec:
      containers:
      - name: simple-linear-example
        image: jthomperoo/predictive-horizontal-pod-autoscaler:latest
        imagePullPolicy: Always
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: php-apache
  provisionRole: false
  config: 
    - name: minReplicas
      value: "1"
    - name: maxReplicas
      value: "5"
    - name: predictiveConfig
      value: |
        models:
        - type: Linear
          name: LinearPrediction
          perInterval: 1
          linear:
            lookAhead: 10000
            storedValues: 6
        decisionType: "maximum"
        metrics:
        - type: Resource
          resource:
            name: cpu
            target:
              averageUtilization: 50
              type: Utilization
    - name: interval
      value: "10000"
```

The predictiveConfig is provided through this environment variable, and represented in YAML:
```yaml
    - name: predictiveConfig
      value: |
        models:
        - type: Linear
          name: LinearPrediction
          perInterval: 1
          linear:
            lookAhead: 10000
            storedValues: 6
        decisionType: "maximum"
```

## decisionType

Example:  
```yaml
decisionType: mean
```

Default value: `maximum`.  
Possible values:

- **maximum** - pick the highest evaluation of the models.
- **minimum** - pick the lowest evaluation of the models.
- **mean** - calculate the mean number of replicas between the models.
- **median** - calculate the median number of replicas between the models.

Decider on which evaluation to pick if there are multiple models provided.

## dbPath

Example:  
```yaml
dbPath: "/tmp/path/store.db"
```
Default value: `/store/predictive-horizontal-pod-autoscaler.db`.  

The path to store the SQLite3 database, e.g. for storing the DB in a volume to persist it.

## migrationPath

Example:
```yaml
migrationPath: "/tmp/migrations/sql"
```

Default value: `/app/sql`.  

The path of the SQL migrations for the SQLite3 database.

## models

List of statistical models to apply.  
See [the models section for details](../../user-guide/models).

## metrics

List of metrics to target for evaluating replica counts.  
See [the metrics section for details](../../user-guide/metrics).

## cpuInitializationPeriod

Example:
```yaml
cpuInitializationPeriod: 150
```
Default value: `300` (5 minutes).  
Set in seconds.  
Equivalent to `--horizontal-pod-autoscaler-cpu-initialization-period`; the period after pod start when CPU samples might be skipped.  

## initialReadinessDelay

Example:
```yaml
initialReadinessDelay: 45
```
Default value: `30` (30 seconds).  
Set in seconds.  
Equivalent to `--horizontal-pod-autoscaler-initial-readiness-delay`; the period after pod start during which readiness changes will be treated as initial readiness.

## tolerance

Example:
```yaml
tolerance: 0.25
```
Default value: `0.1`.  
Equivalent to `--horizontal-pod-autoscaler-tolerance`; the minimum change (from 1.0) in the desired-to-actual metrics ratio for the horizontal pod autoscaler to consider scaling.
