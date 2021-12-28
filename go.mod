module github.com/jthomperoo/predictive-horizontal-pod-autoscaler

go 1.16

require (
	github.com/argoproj/argo-rollouts v1.0.7
	github.com/golang-migrate/migrate/v4 v4.7.0
	github.com/google/go-cmp v0.5.5
	github.com/jthomperoo/custom-pod-autoscaler/v2 v2.3.0
	github.com/jthomperoo/horizontal-pod-autoscaler v0.8.0
	github.com/mattn/go-sqlite3 v2.0.1+incompatible
	k8s.io/api v0.21.8
	k8s.io/apimachinery v0.21.8
	k8s.io/client-go v0.21.8
	k8s.io/kubernetes v1.21.8
	k8s.io/metrics v0.21.8
)

replace (
	k8s.io/api => k8s.io/api v0.21.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.8
	k8s.io/apiserver => k8s.io/apiserver v0.21.8
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.8
	k8s.io/client-go => k8s.io/client-go v0.21.8
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.21.8
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.8
	k8s.io/code-generator => k8s.io/code-generator v0.21.8
	k8s.io/component-base => k8s.io/component-base v0.21.8
	k8s.io/component-helpers => k8s.io/component-helpers v0.21.8
	k8s.io/controller-manager => k8s.io/controller-manager v0.21.8
	k8s.io/cri-api => k8s.io/cri-api v0.21.8
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.21.8
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.8
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.21.8
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.21.8
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.21.8
	k8s.io/kubectl => k8s.io/kubectl v0.21.8
	k8s.io/kubelet => k8s.io/kubelet v0.21.8
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.21.8
	k8s.io/metrics => k8s.io/metrics v0.21.8
	k8s.io/mount-utils => k8s.io/mount-utils v0.21.8
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.21.8
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.21.8
	k8s.io/sample-controller => k8s.io/sample-controller v0.21.8
)
