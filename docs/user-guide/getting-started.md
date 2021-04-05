# Getting Started

This guide will walk through the first steps for deploying a simple Predictive Horizontal Pod Autoscaler (PHPA). This
guide will demonstrate how to deploy a PHPA that uses a linear regression to predict future load based on CPU usage.

To see the final result of this guide, check out the [Simple Linear Regression
example](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/tree/master/examples/simple-linear).

## Prerequisites

This guide requires the following tools installed:

- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) == `v1.X`
- [helm](https://helm.sh/docs/intro/install/) == `v3.X`
- [k3d](https://k3d.io/#installation) == `v4.X`

## Set up the cluster

This guide uses [k3d](https://k3d.io/) to handle provisioning a local K8s server, but you can use any K8s server (you
may already have one set up). If you already have a K8s server configured with the metrics server enabled, skip this
step and move on to the next step.

To provision a new cluster using k3d run the following command:

```bash
k3d cluster create phpa-test-cluster
```

## Install the Custom Pod Autoscaler Operator

The PHPA requires the [Custom Pod Autoscaler Operator
(CPAO)](https://github.com/jthomperoo/custom-pod-autoscaler-operator) to handle management of autoscalers.

In this guide we are using `v1.0.3` of the CPAO, but check out the [installation
guide](https://github.com/jthomperoo/custom-pod-autoscaler-operator/blob/master/INSTALL.md) for more up to date
instructions for later releases.

Run the following commands to install `v1.0.3` of the CPAO:

```bash
VERSION=v1.0.3
HELM_CHART=custom-pod-autoscaler-operator
helm install ${HELM_CHART} https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/download/${VERSION}/custom-pod-autoscaler-operator-${VERSION}.tgz
```

You can check the CPAO has been deployed properly by running:

```bash
kubectl get pods
```

## Create a deployment to autoscale

We now need to create a test application to scale up and down based on load. In this guide we are using an example
container provided by the Kubernetes docs for testing the Horizontal Pod Autoscaler; the test application will simply
respond `OK!` to any request sent to it. This will allow us to adjust how many requests we are sending to the
application to simulate greater and lesser load.

Create a new file called `deployment.yaml` and copy the following YAML into the file:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    run: php-apache
  name: php-apache
spec:
  replicas: 1
  selector:
    matchLabels:
      run: php-apache
  template:
    metadata:
      labels:
        run: php-apache
    spec:
      containers:
      - image: k8s.gcr.io/hpa-example
        imagePullPolicy: Always
        name: php-apache
        ports:
        - containerPort: 80
          protocol: TCP
        resources:
          limits:
            cpu: 500m
          requests:
            cpu: 200m
      restartPolicy: Always
---
apiVersion: v1
kind: Service
metadata:
  name: php-apache
  namespace: default
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    run: php-apache
  sessionAffinity: None
  type: ClusterIP
```

This YAML sets up two K8s resources:

- A [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) to provision some containers to
run our test application that we will scale up and down.
- A [Service](https://kubernetes.io/docs/concepts/services-networking/service/) to expose our test application so we
can send it HTTP requests to affect the CPU load.

Now deploy the application to the K8s cluster by running:

```bash
kubectl apply -f deployment.yaml
```

You can check the test application has been deployed by running:

```bash
kubectl get pods
```

## Create a linear regression autoscaler

Now we need to set up the autoscaler. This autoscaler will be configured to watch our test application's CPU usage and
apply a linear regression to predict ahead of time what the replica count should be.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: simple-linear-example
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - replicationcontrollers
  - replicationcontrollers/scale
  verbs:
  - '*'
- apiGroups:
  - apps
  resources:
  - deployments
  - deployments/scale
  - replicasets
  - replicasets/scale
  - statefulsets
  - statefulsets/scale
  verbs:
  - '*'
- apiGroups:
  - metrics.k8s.io
  resources:
  - '*'
  verbs:
  - '*'
---
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
      value: "10"
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
    - name: downscaleStabilization
      value: "0"
```

This autoscaler works by using the same logic that the Horizontal Pod Autoscaler uses to calculate the number of
replicas a target deployment should have, in this example it tries to make sure that the average CPU utilization is
`50%`. Once it calculates this Horizontal Pod Autoscaler target value, it then stores it and combines it with previous
calculations, feeding it into a linear regression model to try and fit a better prediction.

This example is not hugely practical, it serves primarily as a demonstration, as such it only stores the last 60 seconds
worth of replica target values and tries to fit this into a linear regression. You can see some sample results in
this graph:

![Calculated HPA values vs linear regression predicted values](../img/getting_started_linear_regression.svg)

This shows how as the calculated value drops rapidly from `10` target replicas to `0`, the linear regression results in
a smoothing effect on the actual scaling that takes place; instead it drops from `10` to `5` to `2` and finally to `1`.

The predictive elements are not only for scaling downwards, they could also predict ahead of time an increase in the
required number of replicas, for example with a sequence of increasing calculated replicas (`[1, 3, 5]`) it could
preemptively scale to `7` after applying a linear regression.

The key elements of the PHPA YAML defined above are:

- The role defined at the top allows the autoscaler to have access to the metrics server.

- The autoscaler is targeting our test application; identifying it by the fact it is a `Deployment` with the name
`php-apache`:

```yaml
scaleTargetRef:
  apiVersion: apps/v1
  kind: Deployment
  name: php-apache
```

- The minimum and maximum replicas that the deployment can be autoscaled to are set to the range `0-10`:

```yaml
- name: minReplicas
  value: "1"
- name: maxReplicas
  value: "10"
```

- The frequency that the autoscaler calculates a new target replica value is set to 10 seconds (`10000 ms`).

```yaml
- name: interval
  value: "10000"
```

- The *downscale stabilization* value for the autoscaler is set to `0`, meaning it will only use the latest autoscaling
target and will not pick the highest across a window of time. For more information around this check out the [Custom Pod
Autoscaler wiki](https://custom-pod-autoscaler.readthedocs.io/en/stable/user-guide/cooldown/).

```yaml
- name: downscaleStabilization
  value: "0"
```

- The actual autoscaling decisions are defined in the *predictive config*.
  - The *model* is configured as a linear regression model.
    - The linear regression is set to run every time the autoscaler is run (every interval), in this example it is
    every 10 seconds (`perInterval: 1`).
    - The linear regression is predicting 10 seconds into the future (`lookAhead: 10000`).
    - The linear regression uses a maximum of `6` previous target values for predicting (`storedValues: 6`).
  - The `decisionType` is set to be `maximum`, meaning that the target replicas will be set to whichever is higher
  between the calculated HPA value and the predicted model value.
  - The *metrics* defines the normal Horizontal Pod Autoscaler rules to apply for autoscaling, the results of which
  will have the models applied to for prediction.
    - The metric targeted is the CPU resource of the deployment.
    - The targeted value is that CPU utilization across the test application's containers should be `50%`, if it goes
    too far above this there are not enough pods, and if it goes too far below this there are too many pods.

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
    metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          averageUtilization: 50
          type: Utilization
```

Now deploy the autoscaler to the K8s cluster by running:

```bash
kubectl apply -f phpa.yaml
```

You can check the autoscaler has been deployed by running:

```bash
kubectl get pods
```

## Apply load and monitor the autoscaling process

You can monitor the autoscaling process by running:

```bash
kubectl logs simple-linear-example --follow
```

You can see the targets calculated by the HPA logic before the linear regression has been applied to them by querying
the autoscaler's internal database:

```bash
kubectl exec -it simple-linear-example -- sqlite3 /store/predictive-horizontal-pod-autoscaler.db 'SELECT * FROM evaluation;'
```

You can increase the load by starting a new container, and looping to send a bunch of HTTP requests to our test
application:

```bash
kubectl run -it --rm load-generator --image=busybox /bin/sh
```

To start making requests from this container, run:

```bash
while true; do wget -q -O- http://php-apache.default.svc.cluster.local; done
```

You can stop this request loop by hitting *Ctrl+c*.

Try and start increasing the load, then stopping, you should be able to see a difference between the calculated HPA
values and the target values predicted by the linear regression.

## Delete the cluster and clean up

Once you have finished testing the autoscaler, you can clean up any K8s resources by running:

```bash
HELM_CHART=custom-pod-autoscaler-operator
kubectl delete -f deployment.yaml
kubectl delete -f phpa.yaml
helm uninstall ${HELM_CHART}
```

If you are using k3d you can clean up the entire cluster by running:

```bash
k3d cluster delete phpa-test-cluster
```

## Conclusion

This guide is intended to provide a simple walkthrough of how to install and use the PHPA, the concepts outlined here
can be used to deploy autoscalers with different predictive models. Check out the [examples in the project Git
repository to see more
samples](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/tree/master/examples).
