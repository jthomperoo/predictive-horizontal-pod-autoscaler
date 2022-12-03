/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/scale"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	cached "k8s.io/client-go/discovery/cached"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/jthomperoo/k8shorizmetrics/v2"
	"github.com/jthomperoo/k8shorizmetrics/v2/metricsclient"
	"github.com/jthomperoo/k8shorizmetrics/v2/podsclient"

	jamiethompsonmev1alpha1 "github.com/jthomperoo/predictive-horizontal-pod-autoscaler/api/v1alpha1"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/algorithm"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/controllers"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/hook/http"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/prediction"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/prediction/holtwinters"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/prediction/linear"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(jamiethompsonmev1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "a07591f8.jamiethompson.me",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	clientSet, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to get clientset from config")
		os.Exit(1)
	}

	scaleClient, err := scale.NewForConfig(mgr.GetConfig(),
		restmapper.NewDeferredDiscoveryRESTMapper(cached.NewMemCacheClient(clientSet.Discovery())),
		dynamic.LegacyAPIPathResolverFunc,
		scale.NewDiscoveryScaleKindResolver(clientSet.Discovery()))
	if err != nil {
		setupLog.Error(err, "unable to set up scale client from config")
		os.Exit(1)
	}

	metricsclient := metricsclient.NewClient(mgr.GetConfig(), clientSet.Discovery())
	podsclient := &podsclient.OnDemandPodLister{Clientset: clientSet}
	cpuInitializationPeriod := time.Duration(300) * time.Second
	initialReadinessDelay := time.Duration(30) * time.Second
	tolerance := 0.1
	pyRunner := algorithm.NewAlgorithmPython()
	httpExec := &http.Execute{}

	if err = (&controllers.PredictiveHorizontalPodAutoscalerReconciler{
		Client:      mgr.GetClient(),
		Scheme:      mgr.GetScheme(),
		ScaleClient: scaleClient,
		Gatherer:    *k8shorizmetrics.NewGatherer(metricsclient, podsclient, cpuInitializationPeriod, initialReadinessDelay),
		Evaluator:   *k8shorizmetrics.NewEvaluator(tolerance),
		Predicter: &prediction.ModelPredict{
			Predicters: []prediction.Predicter{
				&linear.Predict{
					Runner: pyRunner,
				},
				&holtwinters.Predict{
					HookExecute: httpExec,
					Runner:      pyRunner,
				},
			},
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PredictiveHorizontalPodAutoscaler")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
