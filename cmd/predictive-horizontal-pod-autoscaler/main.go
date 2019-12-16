/*
Copyright 2019 The Predictive Horizontal Pod Autoscaler Authors.

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
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file" // Driver for loading evaluations from file system
	cpametric "github.com/jthomperoo/custom-pod-autoscaler/metric"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/prediction/holtwinters"
	hpaevaluate "github.com/jthomperoo/horizontal-pod-autoscaler/evaluate"
	"github.com/jthomperoo/horizontal-pod-autoscaler/metric"
	"github.com/jthomperoo/horizontal-pod-autoscaler/podclient"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/config"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/evaluate"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/prediction"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/prediction/linear"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/stored"
	_ "github.com/mattn/go-sqlite3" // Driver for sqlite3 database	appsv1 "k8s.io/api/apps/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	"k8s.io/apimachinery/pkg/util/yaml"
	cacheddiscovery "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/kubernetes"
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
	predictiveConfig, err := config.LoadConfig([]byte(predictiveConfigEnv))
	if err != nil {
		log.Fatal(err)
	}

	switch *modePtr {
	case "metric":
		getMetrics(bytes.NewReader(stdin))
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

	// Read in resource metrics provided
	var resourceMetrics cpametric.ResourceMetrics
	err = yaml.NewYAMLOrJSONDecoder(stdin, 10).Decode(&resourceMetrics)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	if len(resourceMetrics.Metrics) != 1 {
		log.Fatalf("Expected 1 CPA metric, got %d", len(resourceMetrics.Metrics))
		os.Exit(1)
	}

	var combinedMetrics []*metric.Metric
	err = yaml.NewYAMLOrJSONDecoder(strings.NewReader(resourceMetrics.Metrics[0].Value), 10).Decode(&combinedMetrics)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Set up evaluator
	evaluator := &evaluate.PredictiveEvaluate{
		HPAEvaluator: hpaevaluate.NewEvaluate(0.1),
		Store: &stored.LocalStore{
			DB: db,
		},
		Predicters: []prediction.Predicter{
			&linear.Predict{},
			&holtwinters.Predict{},
		},
	}

	// Get evaluation
	result, err := evaluator.GetEvaluation(predictiveConfig, combinedMetrics, resourceMetrics.RunType)
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

func getMetrics(stdin io.Reader) {
	var deployment appsv1.Deployment
	err := yaml.NewYAMLOrJSONDecoder(stdin, 10).Decode(&deployment)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	metricSpecsValue, exists := os.LookupEnv("metrics")
	if !exists {
		log.Fatal("Metric specs not supplied")
		os.Exit(1)
	}

	// Read in metric specs to evaluate
	var metricSpecs []autoscalingv2.MetricSpec
	err = yaml.NewYAMLOrJSONDecoder(strings.NewReader(metricSpecsValue), 10).Decode(&metricSpecs)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	if len(metricSpecs) == 0 {
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
	), &podclient.OnDemandPodLister{Clientset: clientset}, 5*time.Minute, 30*time.Second)

	// Get metrics for deployment
	metrics, err := gatherer.GetMetrics(&deployment, metricSpecs, deployment.ObjectMeta.Namespace)
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
