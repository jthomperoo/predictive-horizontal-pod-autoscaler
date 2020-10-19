/*
Copyright 2020 The Predictive Horizontal Pod Autoscaler Authors.

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

// Predictive Horizontal Pod Autoscaler provides executable Predictive Horizontal Pod Autoscaler logic, which
// can be built into a Custom Pod Autoscaler.
// The Horizontal Pod Autoscaler has two modes, metric gathering and evaluation.
// Metric mode gathers metrics, taking in a resource to get the metrics for and outputting
// these metrics as serialised JSON.
// Evaluation mode makes decisions on how many replicas a resource should have, taking in
// metrics and outputting evaluation decisions as seralised JSON.
// The predictive element uses past evaluations to predict how to scale in the future.
package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file" // Driver for loading evaluations from file system
	"github.com/jthomperoo/custom-pod-autoscaler/execute"
	"github.com/jthomperoo/custom-pod-autoscaler/execute/http"
	"github.com/jthomperoo/custom-pod-autoscaler/execute/shell"
	cpametric "github.com/jthomperoo/custom-pod-autoscaler/metric"
	hpaevaluate "github.com/jthomperoo/horizontal-pod-autoscaler/evaluate"
	"github.com/jthomperoo/horizontal-pod-autoscaler/metric"
	"github.com/jthomperoo/horizontal-pod-autoscaler/podclient"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/config"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/evaluate"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/prediction"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/prediction/holtwinters"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/prediction/linear"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/stored"
	_ "github.com/mattn/go-sqlite3" // Driver for sqlite3 database	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	cacheddiscovery "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/kubernetes/pkg/controller/podautoscaler/metrics"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
	customclient "k8s.io/metrics/pkg/client/custom_metrics"
	externalclient "k8s.io/metrics/pkg/client/external_metrics"
)

const (
	predictiveConfigEnvName = "predictiveConfig"
)

// EvaluateSpec represents the information fed to the evaluator
type EvaluateSpec struct {
	Metrics              []*cpametric.Metric       `json:"metrics"`
	UnstructuredResource unstructured.Unstructured `json:"resource"`
	Resource             metav1.Object             `json:"-"`
	RunType              string                    `json:"runType"`
}

// MetricSpec represents the information fed to the metric gatherer
type MetricSpec struct {
	UnstructuredResource unstructured.Unstructured `json:"resource"`
	Resource             metav1.Object             `json:"-"`
	RunType              string                    `json:"runType"`
}

func main() {
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	modePtr := flag.String("mode", "no_mode", "command mode, either metric or evaluate")
	flag.Parse()

	// Get config env variable
	predictiveConfigEnv, exists := os.LookupEnv(predictiveConfigEnvName)
	if !exists {
		log.Fatal("Missing required predictiveConfig environment variable.")
	}

	// Parse config
	predictiveConfig, err := config.LoadConfig(strings.NewReader(predictiveConfigEnv))
	if err != nil {
		log.Fatal(err)
	}

	switch *modePtr {
	case "metric":
		getMetrics(bytes.NewReader(stdin), predictiveConfig)
	case "evaluate":
		getEvaluation(bytes.NewReader(stdin), predictiveConfig)
	case "setup":
		setup(predictiveConfig)
	default:
		log.Fatalf("Unknown command mode: %s", *modePtr)
		os.Exit(1)
	}
}

func setup(predictiveConfig *config.Config) {
	// Make folders for storing DB
	err := os.MkdirAll(filepath.Dir(predictiveConfig.DBPath), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	// Open DB connection
	db, err := sql.Open("sqlite3", predictiveConfig.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Get DB driver
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Load migrations
	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file:///%s", predictiveConfig.MigrationPath), "evaluations", driver)
	if err != nil {
		log.Fatal(err)
	}

	// Apply migrations
	err = m.Up()
	if err != nil {
		log.Fatal(err)
	}
}

func getEvaluation(stdin io.Reader, predictiveConfig *config.Config) {
	// Open DB connection
	db, err := sql.Open("sqlite3", predictiveConfig.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var spec EvaluateSpec
	err = yaml.NewYAMLOrJSONDecoder(stdin, 10).Decode(&spec)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Create object from version and kind of piped value
	resourceGVK := spec.UnstructuredResource.GroupVersionKind()
	resourceRuntime, err := scheme.Scheme.New(resourceGVK)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Parse the unstructured k8s resource into the object created, then convert to generic metav1.Object
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(spec.UnstructuredResource.Object, resourceRuntime)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	spec.Resource = resourceRuntime.(metav1.Object)

	var combinedMetrics []*metric.Metric
	err = yaml.NewYAMLOrJSONDecoder(strings.NewReader(spec.Metrics[0].Value), 10).Decode(&combinedMetrics)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Set up shell executer
	shellExec := &shell.Execute{
		Command: exec.Command,
	}

	httpExec := &http.Execute{}

	// Combine executers
	combinedExecute := &execute.CombinedExecute{
		Executers: []execute.Executer{
			shellExec,
			httpExec,
		},
	}

	// Set up evaluator
	evaluator := &evaluate.PredictiveEvaluate{
		HPAEvaluator: hpaevaluate.NewEvaluate(predictiveConfig.Tolerance),
		Store: &stored.LocalStore{
			DB: db,
		},
		Predicters: []prediction.Predicter{
			&linear.Predict{},
			&holtwinters.Predict{
				Execute: combinedExecute,
			},
		},
	}

	// Get evaluation
	result, err := evaluator.GetEvaluation(predictiveConfig, combinedMetrics, spec.RunType)
	if err != nil {
		log.Fatal(err)
	}

	// Marshal evaluation into JSON
	jsonEvaluation, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	fmt.Print(string(jsonEvaluation))
}

func getMetrics(stdin io.Reader, predictiveConfig *config.Config) {
	var spec MetricSpec
	err := yaml.NewYAMLOrJSONDecoder(stdin, 10).Decode(&spec)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Create object from version and kind of piped value
	resourceGVK := spec.UnstructuredResource.GroupVersionKind()
	resourceRuntime, err := scheme.Scheme.New(resourceGVK)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Parse the unstructured k8s resource into the object created, then convert to generic metav1.Object
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(spec.UnstructuredResource.Object, resourceRuntime)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	spec.Resource = resourceRuntime.(metav1.Object)

	if len(predictiveConfig.Metrics) == 0 {
		log.Fatal("Metric specs not supplied")
		os.Exit(1)
	}

	// Create the in-cluster Kubernetes config
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Create the Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Create metric gatherer, with required clients and configuration
	gatherer := metric.NewGather(metrics.NewRESTMetricsClient(
		resourceclient.NewForConfigOrDie(clusterConfig),
		customclient.NewForConfig(
			clusterConfig,
			restmapper.NewDeferredDiscoveryRESTMapper(cacheddiscovery.NewMemCacheClient(clientset.Discovery())),
			customclient.NewAvailableAPIsGetter(clientset.Discovery()),
		),
		externalclient.NewForConfigOrDie(clusterConfig),
	),
		&podclient.OnDemandPodLister{Clientset: clientset},
		time.Duration(predictiveConfig.CPUInitializationPeriod)*time.Second,
		time.Duration(predictiveConfig.InitialReadinessDelay)*time.Second,
	)

	// Get metrics for deployment
	metrics, err := gatherer.GetMetrics(spec.Resource, predictiveConfig.Metrics, spec.Resource.GetNamespace())
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Marshal metrics into JSON
	jsonMetrics, err := json.Marshal(metrics)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Write serialised metrics to stdout
	fmt.Print(string(jsonMetrics))
}
