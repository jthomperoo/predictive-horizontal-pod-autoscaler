# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic
Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.13.2] - 2023-07-01
### Changed
- Upgraded `statsmodels` to `0.14.0`.
- Upgraded `dataclasses-json` to `0.5.8`.

## [v0.13.1] - 2023-03-07
### Changed
- Upgraded to Go `v1.20`.
- Upgraded package dependencies.

## [v0.13.0] - 2023-01-16
### Added
- New ability to choose a start interval for a model, e.g. a holt winters model that only starts at the next full hour.
  - `startInterval` is an optional [duration](https://pkg.go.dev/time#ParseDuration) that defines the next interval to
  start at. For example `1m` would start the model recording data and calculating after the next full minute has passed.
- New ability to clear a model's replica history if too much time has passed since it last recorded any data.
  - `resetDuration` is an optional [duration](https://pkg.go.dev/time#ParseDuration) that defines how long should be
  allowed to pass without a model recording any data, if that occurs the replica history is cleared and if a start
  interval is provided a new start time is calculated. If a model is not calculated for an extended period of time
  (e.g. a cluster being powered off) this allows old data to be cleared out and not used in calculations and a new
  start time calculated to respect any provided interval. For example `3m` would clear the model's replica history
  (and reset the start time if interval is provided) if it has been more than 3 minutes since the last data was
  recorded for the model.
### Changed
- **BREAKING CHANGE** typo fixed, `ModelHistory.ReplicaHistroy` renamed to `ModelHistory.ReplicaHistory`.

## [v0.12.0] - 2023-01-15
### Changed
- See the [migration guide from `v0.11.2`
here](https://predictive-horizontal-pod-autoscaler.readthedocs.io/en/latest/user-guide/migration/v0_11_2-to-v0_12_0).
- **BREAKING CHANGE** PHPA spec upgraded from `autoscaling/v2beta2` to `autoscaling/v2` for the following definitions:
  - `CrossVersionObjectReference` in the `scaleTargetRef` field.
  - `MetricSpec` in the `metrics` field.
  - `MetricStatus` in the `currentMetrics` field.
- Upgraded to [k8shorizmetrics `v2.0.0`](https://github.com/jthomperoo/k8shorizmetrics/releases/tag/v2.0.0).
- Upgraded from `autoscaling/v2beta2` to `autoscaling/v2`.
- Upgraded to Go `v1.19`.
### Removed
- **BREAKING CHANGE** Removed `downscaleStabilization`, replaced with `behavior`, `scaleDown`,
`stabilizationWindowSeconds`.
### Added
- Support for [HPA scaling
behaviors](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#configurable-scaling-behavior).

## [v0.11.2] - 2022-11-25
### Fixed
- Bug that would display a statsmodels error when using Holt-Winters and having too few observations (less than 2 *
seasonal periods).

## [v0.11.1] - 2022-07-23
### Fixed
- Bug that caused the autoscaler to fail to scale every time, due to invalid K8s GVR resource fed to the scaling
client.

## [v0.11.0] - 2022-07-23
### Changed
- See the [migration guide from `v0.10.0` here](https://predictive-horizontal-pod-autoscaler.readthedocs.io/en/latest/user-guide/migration/v0_10_0-to-v0_11_0).
- BREAKING CHANGE: Major rewrite converting this project from a Custom Pod Autoscaler to have its own dedicated CRD
and operator.
  - Configuration and deployment has changed completely, no longer need to install the Custom Pod Autoscaler Operator,
  instead you need to install the Predictive Horizontal Pod Autoscaler as an operator.
  - No longer deployed as `CustomPodAutoscaler` custom resources, now deployed as `PredictiveHorizontalPodAutoscaler`
  custom resources.
- BREAKING CHANGES: Several configuration options renamed for clarity.
  - `LinearModel -> storedValues` renamed to `LinearModel -> historySize`
  - `Model -> perInterval` renamed to `Model -> perSyncPeriod`
- BREAKING CHANGES: `perSyncPeriod` behaviour changed slightly, now it will no longer calculate the prediction, but
it will still update the replica history available to make a prediction with and prune the replica history if needed.
- Holt-Winters runtime hooks format changed:
  - `evaluations` renamed to `replicaHistory`.
  - Format change for `replicaHistory`, now in the format `[{"time": "<timestamp>", "replicas": <replica count>}]`.
- Upgrade to Go `v1.18`.
- No longer SQLite based storage, instead using Kubernetes configmaps which give persistent storage by default with
resiliency when the autoscaler operator restarts.
### Removed
- BREAKING CHANGES: Removed some no longer relevant configuration options.
  - `DBPath`
  - `MigrationPath`
- BREAKING CHANGE: Since no longer built as a `CustomPodAutoscaler` the `startTime` configuration is no longer
available: <https://custom-pod-autoscaler.readthedocs.io/en/latest/reference/configuration/#starttime>.
- BREAKING CHANGE: Holt-Winters runtime hooks limited to only support HTTP, shell support dropped.

## [v0.10.0] - 2022-05-14
### Changed
- Removed dependency on `jthomperoo/horizontal-pod-autoscaler` in favour of `jthomperoo/k8shorizmetrics`.
- Bump `jthomperoo/custom-pod-autoscaler` to `v2.6.0`.
- Upgrade to Go `v1.17`.

## [v0.9.0] - 2021-12-28
### Added
- Support for `argoproj.io/v1alpha1` `Rollout` resource.
### Changed
- Bump `jthomperoo/custom-pod-autoscaler` to `v2.3.0`
- Bump `jthomperoo/horizontal-pod-autoscaler` to `v0.8.0`

## [v0.8.0] - 2021-08-15
### Changed
- Updated to Custom Pod Autoscaler `v2.2.0`.
- Updated to Horizontal Pod Autoscaler CPA `v0.7.0`.
### Fixed
- Linear regression no longer fails on first run.

## [v0.7.0] - 2021-04-05
### Added
- Holt Winters values can now be fetched at runtime, rather than simply being hardcoded.
### Fixed
- Fixed slow shutdown of PHPA due to ignoring SIGTERM from K8s.
### Changed
- Switched from Golang to Python for calculating statistical predictions for Linear Regression and Holt-Winters.
- Holt-Winters now calculated using statsmodels, opening up statsmodels configuration options for tuning.
  - `trend` - Either `add`/`additive` or `mul`/`multiplicative`, defines the method for the trend element.
  - `seasonal` - Either `add`/`additive` or `mul`/`multiplicative`, defines the method for the seasonal element.
  - `dampedTrend` - Boolean value to determine if the trend should be damped.
  - `initializationMethod` - Which initialization method to use, see statsmodels for details, either `estimated`,
  `heuristic`, `known`, or `legacy-heuristic`
  - `initialLevel` - The initial level value, required if `initializationMethod` is `known`.
  - `initialTrend` - The initial trend value, required if `initializationMethod` is `known`.
  - `initialSeasonal` - The initial seasonal value, required if `initializationMethod` is `known`.
- Holt-Winters `seasonLength` variable renamed to `seasonalPeriods`.
- Holt-Winters `method` split into `trend` and `seasonal` variables.
- Switched docs theme to material theme.

## [v0.6.0] - 2020-08-31
### Changed
- Update Custom Pod Autoscaler version to `v1.0.0`.
- Update Horizontal Pod Autoscaler version to `v0.6.0`.

## [v0.5.0] - 2020-03-27
### Changed
- Evaluation from HPA now included in list of predicted replica counts, rather than being treated separately at the end.
Now included in mean, median, minimum calculations rather than just the maximum calculation.

## [v0.4.0] - 2020-03-10
### Added
- Documentation as code; configuration reference.
- New decision type `median`, returns the median average of the predictions.
- JSON support for configuration options.
- Can now configure `tolerance`, `initialReadinessDelay` and `initializationPeriod` that are available to be configured
in the K8s HPA.
- Default `downscaleStabilization` set to `300` (5 minutes) to match K8s HPA.
### Changed
- Metric specs now defined in `predictiveConfig` rather than in their own
  section.
- Update Custom Pod Autoscaler version to v0.11.0.
- Update Horizontal Pod Autoscaler version to v0.5.0.
- Default `interval` set to `15000` (15 seconds) to match K8s HPA.
- Default `minReplicas` set to `1` to match K8s HPA.
- Default `maxReplicas` set to `10` to match K8s HPA.

## [v0.3.0] - 2020-02-17
### Added
- Multiplicative method for Holt-Winters time series prediction.
### Changed
- Update Custom Pod Autoscaler version to v0.10.0.
- Update Horizontal Pod Autoscaler version to v0.4.0.
- Holt-Winters no longer additive by default, must specify a method, either `additive` or `multiplicative` in the
Holt-Winters configuration.

## [v0.2.0] - 2019-12-19
### Added
- Holt-Winters time series based prediction model.

## [v0.1.0] - 2019-12-9
### Added
- Added the ability to use Linear Regression models to predict future scaling.

[Unreleased]:
https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.13.2...HEAD
[v0.13.2]:
https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.13.1...v0.13.2
[v0.13.1]:
https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.13.0...v0.13.1
[v0.13.0]:
https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.12.0...v0.13.0
[v0.12.0]:
https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.11.2...v0.12.0
[v0.11.2]:
https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.11.1...v0.11.2
[v0.11.1]:
https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.11.0...v0.11.1
[v0.11.0]:
https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.10.0...v0.11.0
[v0.10.0]:
https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.9.0...v0.10.0
[v0.9.0]:
https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.8.0...v0.9.0
[v0.8.0]:
https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.7.0...v0.8.0
[v0.7.0]:
https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.6.0...v0.7.0
[v0.6.0]:
https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.5.0...v0.6.0
[v0.5.0]:
https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.4.0...v0.5.0
[v0.4.0]:
https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.3.0...v0.4.0
[v0.3.0]:
https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.2.0...v0.3.0
[v0.2.0]:
https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.1.0...v0.2.0
[v0.1.0]:
https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/releases/tag/v0.1.0
