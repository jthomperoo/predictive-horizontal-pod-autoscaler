# Installation

You can install Predictive Horizontal Pod Autoscalers (PHPAs) on your cluster after you have installed the Predictive
Horizontal Pod Autoscaler Operator (PHPA operator) onto your cluster.

The PHPA operator can be installed using Helm, run this command to install the operator onto your cluster with
cluster-wide scope:

```bash
VERSION=v0.13.2
HELM_CHART=predictive-horizontal-pod-autoscaler-operator
helm install ${HELM_CHART} https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/releases/download/${VERSION}/predictive-horizontal-pod-autoscaler-${VERSION}.tgz
```

After you have done that you can install PHPAs onto your cluster, check out the [examples for PHPAs you can
deploy](https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/tree/master/examples) or follow the [getting
started guide](./getting-started.md).
