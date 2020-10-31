# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic
Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Holt Winters values can now be fetched at runtime, rather than simply being hardcoded.
### Fixed
- Fixed slow shutdown of PHPA due to ignoring SIGTERM from K8s.
### Changed
- Switched from Golang to Python for calculating statistical predictions for Linear Regression and Holt-Winters.

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
https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.6.0...HEAD
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
