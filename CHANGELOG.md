# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Documentation as code; configuration reference.
- New decision type `median`, returns the median average of the predictions.
### Changed
- Metric specs now defined in `predictiveConfig` rather than in their own section.

## [v0.3.0] - 2020-02-17
### Added
- Multiplicative method for Holt-Winters time series prediction.
### Changed
- Update Custom Pod Autoscaler version to v0.10.0.
- Update Horizontal Pod Autoscaler version to v0.4.0.
- Holt-Winters no longer additive by default, must specify a method, either `additive` or `multiplicative` in the Holt-Winters configuration.

## [v0.2.0] - 2019-12-19
### Added
- Holt-Winters time series based prediction model.

## [v0.1.0] - 2019-12-9
### Added
- Added the ability to use Linear Regression models to predict future scaling.

[Unreleased]: https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.3.0...HEAD
[v0.3.0]: https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/jthomperoo/predictive-horizontal-pod-autoscaler/releases/tag/v0.1.0