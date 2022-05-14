/*
Copyright 2022 The Predictive Horizontal Pod Autoscaler Authors.

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
	"context"
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
	cpametric "github.com/jthomperoo/custom-pod-autoscaler/v2/metric"
	"github.com/jthomperoo/k8shorizmetrics"
	"github.com/jthomperoo/k8shorizmetrics/metrics"
	"github.com/jthomperoo/k8shorizmetrics/metricsclient"
	"github.com/jthomperoo/k8shorizmetrics/podsclient"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/algorithm"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/config"
	phpaevaluate "github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/evaluate"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/hook"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/hook/http"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/hook/shell"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/prediction"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/prediction/holtwinters"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/prediction/linear"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/stored"
	_ "github.com/mattn/go-sqlite3" // Driver for sqlite3 database	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	k8sscale "k8s.io/client-go/scale"
)

const (
	predictiveConfigEnvName = "predictiveConfig"
)

// EvaluateSpec represents the information fed to the evaluator
type EvaluateSpec struct {
	Metrics              []*cpametric.ResourceMetric `json:"metrics"`
	UnstructuredResource unstructured.Unstructured   `json:"resource"`
	Resource             metav1.Object               `json:"-"`
	RunType              string                      `json:"runType"`
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
		gather(bytes.NewReader(stdin), predictiveConfig)
	case "evaluate":
		evaluate(bytes.NewReader(stdin), predictiveConfig)
	case "setup":
		setup(predictiveConfig)
	default:
		log.Fatalf("Unknown command mode: %s", *modePtr)
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

func gather(stdin io.Reader, predictiveConfig *config.Config) {
	var spec MetricSpec
	err := yaml.NewYAMLOrJSONDecoder(stdin, 10).Decode(&spec)
	if err != nil {
		log.Fatal(err)
	}

	if len(predictiveConfig.Metrics) == 0 {
		log.Fatal("Metric specs not supplied")
	}

	clusterConfig, clientset, err := getKubernetesClients()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	scale, err := getScaleSubResource(clientset, &spec.UnstructuredResource)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	metricsclient := metricsclient.NewClient(clusterConfig, clientset.Discovery())
	podsclient := &podsclient.OnDemandPodLister{
		Clientset: clientset,
	}
	cpuInitializationPeriod := time.Duration(predictiveConfig.CPUInitializationPeriod) * time.Second
	initialReadinessDelay := time.Duration(predictiveConfig.InitialReadinessDelay) * time.Second

	// Create metric gatherer, with required clients and configuration
	gatherer := k8shorizmetrics.NewGatherer(metricsclient, podsclient, cpuInitializationPeriod, initialReadinessDelay)

	selector, err := labels.Parse(scale.Status.Selector)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Get metrics for deployment
	metrics, err := gatherer.Gather(predictiveConfig.Metrics, scale.GetNamespace(), selector)
	if err != nil {
		log.Fatal(err)
	}

	// Marshal metrics into JSON
	jsonMetrics, err := json.Marshal(metrics)
	if err != nil {
		log.Fatal(err)
	}

	// Write serialised metrics to stdout
	fmt.Print(string(jsonMetrics))
}

func evaluate(stdin io.Reader, predictiveConfig *config.Config) {
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
	}

	_, clientset, err := getKubernetesClients()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	scale, err := getScaleSubResource(clientset, &spec.UnstructuredResource)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	var combinedMetrics []*metrics.Metric
	err = yaml.NewYAMLOrJSONDecoder(strings.NewReader(spec.Metrics[0].Value), 10).Decode(&combinedMetrics)
	if err != nil {
		log.Fatal(err)
	}

	shellExec := &shell.Execute{
		Command: exec.Command,
	}

	httpExec := &http.Execute{}

	// Combine executers
	combinedExecute := &hook.CombinedExecute{
		Executers: []hook.Executer{
			shellExec,
			httpExec,
		},
	}

	algorithmRunner := &algorithm.Run{
		Executer: shellExec,
	}

	// K8s evaluator
	k8sevaluator := k8shorizmetrics.NewEvaluator(predictiveConfig.Tolerance)

	// Set up evaluator
	evaluator := &phpaevaluate.PredictiveEvaluate{
		HPAEvaluator: k8sevaluator,
		Store: &stored.LocalStore{
			DB: db,
		},
		Predicters: []prediction.Predicter{
			&linear.Predict{
				Runner: algorithmRunner,
			},
			&holtwinters.Predict{
				Runner:  algorithmRunner,
				Execute: combinedExecute,
			},
		},
	}

	// Get evaluation
	result, err := evaluator.GetEvaluation(predictiveConfig, combinedMetrics, scale.Spec.Replicas, spec.RunType)
	if err != nil {
		log.Fatal(err)
	}

	// Marshal evaluation into JSON
	jsonEvaluation, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(string(jsonEvaluation))
}

func getScaleSubResource(clientset *kubernetes.Clientset, resource *unstructured.Unstructured) (*autoscalingv1.Scale, error) {
	groupResources, err := restmapper.GetAPIGroupResources(clientset.Discovery())
	if err != nil {
		return nil, err
	}

	scaleClient := k8sscale.New(
		clientset.RESTClient(),
		restmapper.NewDiscoveryRESTMapper(groupResources),
		dynamic.LegacyAPIPathResolverFunc,
		k8sscale.NewDiscoveryScaleKindResolver(
			clientset.Discovery(),
		),
	)

	resourceGV, err := schema.ParseGroupVersion(resource.GetAPIVersion())
	if err != nil {
		return nil, err
	}

	targetGR := schema.GroupResource{
		Group:    resourceGV.Group,
		Resource: resource.GetKind(),
	}

	scale, err := scaleClient.Scales(resource.GetNamespace()).Get(context.Background(), targetGR, resource.GetName(), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return scale, nil
}

func getKubernetesClients() (*rest.Config, *kubernetes.Clientset, error) {
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, err
	}

	clientset, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return nil, nil, err

	}

	return clusterConfig, clientset, nil
}
