## [3.0.0](https://github.com/ivannovak/glide-plugin-docker/compare/v2.4.0...v3.0.0) (2025-12-01)


### ⚠ BREAKING CHANGES

* This plugin now requires Glide v2.4.0+ with SDK v2.

Migration to SDK v2:
- Replace v1.BasePlugin with v2.BasePlugin[Config]
- Add type-safe Config struct with defaultProfile, autoDetectFiles,
  and projectName options
- Update main.go to use v2.Serve()
- Create new plugin.go with SDK v2 patterns
- Remove legacy grpc_plugin.go (v1 implementation)
- Add Docker health check in HealthCheck()
- Declare RequiresDocker capability in metadata

The plugin now uses the declarative SDK v2 pattern with:
- Type-safe configuration via Go generics
- Unified lifecycle management with Docker availability check
- Declarative commands via Commands() method

Note: go.mod includes a replace directive pointing to local glide
repo until v2.4.0 with SDK v2 is released.

* feat!: upgrade to glide SDK v3.0.0
* Updates module dependency from glide/v2 to glide/v3 v3.0.0.
This aligns with the SDK v2 type-safe configuration system released in glide v3.0.0.

- Update go.mod to require github.com/ivannovak/glide/v3 v3.0.0
- Remove local replace directive (now using published version)
- Update all imports from /v2/ to /v3/

* ci: add CI workflow for PR validation

### Features

* upgrade to glide SDK v3.0.0 ([#1](https://github.com/ivannovak/glide-plugin-docker/issues/1)) ([4cd1535](https://github.com/ivannovak/glide-plugin-docker/commit/4cd15356d89752d089ec21172c58eb535db2f00f))

## [2.4.0](https://github.com/ivannovak/glide-plugin-docker/compare/v2.3.2...v2.4.0) (2025-11-25)


### Features

* add main entry point for plugin binary ([b6cb709](https://github.com/ivannovak/glide-plugin-docker/commit/b6cb70945c1bd6e09749865472eda00d322a1ac3))

## [2.3.2](https://github.com/ivannovak/glide-plugin-docker/compare/v2.3.1...v2.3.2) (2025-11-25)


### Bug Fixes

* correct build path in release workflow ([eb7c4b0](https://github.com/ivannovak/glide-plugin-docker/commit/eb7c4b0e87569aba0b32383e2adec651023bc7cd))

## [2.3.1](https://github.com/ivannovak/glide-plugin-docker/compare/v2.3.0...v2.3.1) (2025-11-25)


### Bug Fixes

* remove CI dependency from release workflow ([4c0160d](https://github.com/ivannovak/glide-plugin-docker/commit/4c0160decacd6f0052aac6a7577db13475b821ff))

## [2.3.0](https://github.com/ivannovak/glide-plugin-docker/compare/v2.2.0...v2.3.0) (2025-11-25)


### Features

* use published Glide v2.2.0 ([46c90cb](https://github.com/ivannovak/glide-plugin-docker/commit/46c90cbdce196ab6c9de9db25628ab0f3849c039))

## [2.2.0](https://github.com/ivannovak/glide-plugin-docker/compare/v2.1.0...v2.2.0) (2025-11-24)


### Features

* migrate to Glide v2 module path ([f8a7254](https://github.com/ivannovak/glide-plugin-docker/commit/f8a7254cd6959ca72625b2916353e1b98a4b8cdc))

## [2.1.0](https://github.com/ivannovak/glide-plugin-docker/compare/v2.0.0...v2.1.0) (2025-11-24)


### Features

* add release workflow for cross-platform binaries ([47ad3b5](https://github.com/ivannovak/glide-plugin-docker/commit/47ad3b59e9913f104502a647d9b10375a8465ee8))

## [2.0.0](https://github.com/ivannovak/glide-plugin-docker/compare/v1.0.0...v2.0.0) (2025-11-24)


### ⚠ BREAKING CHANGES

* Plugin now uses gRPC instead of library architecture

### Code Refactoring

* migrate to gRPC architecture and cleanup legacy code ([013e1be](https://github.com/ivannovak/glide-plugin-docker/commit/013e1be41336bfc75adb336c5a77c15a9d5fb2b0))

## 1.0.0 (2025-11-21)


### Features

* **plugin:** initial Docker plugin extraction (Phase 6) ([545ae53](https://github.com/ivannovak/glide-plugin-docker/commit/545ae5308df59fbc0e446339fbafbce719b74892))
