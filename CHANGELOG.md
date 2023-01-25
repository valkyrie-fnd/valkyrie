# Changelog

<!--
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
-->

All notable changes to this project will be documented in this file.

This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Provider config `base_path`. Base path for endpoints related to this provider.
- Provider config `provider_specific`. Can be passed to the provider implementation
- Config `operator_base_path`. General base path for all requests from operator side
- Config `provider_base_path`. General base path for all request from provider side
- Game round render endpoint for provider Caleta

### Changed

### Removed

## [0.5.0] - 2023-01-18

### Added

- Add Vplugin server
- Add BetCode to Transaction in pam api specification
- Added default cpu and memory requests in helm chart
- Added support for "X-Msg-Timestamp"-header for Caleta provider
- Added tracing propagation to vplugin

### Changed

- Split request/response logging into separate log statements
- Only use swagger when using "dev" build tag
- vplugin consistently configured using snake case keys

### Removed


## [0.4.0] - 2023-01-10

### Added
- Add zerolog adapter for hclog
- Add latest tag to built image

### Changed
- Changed swagger package location
- Changed Caleta swagger documentation
- Changed helm repository
- Changed plugin config
### Removed
## [0.3.0] - 2022-12-20

### Added

- Added a provider implementation for Evolution
- Added a provider implementation for Red Tiger
- Added a provider implementation for Caleta

### Changed

### Removed
