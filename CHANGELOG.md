# Changelog

All notable changes to this project will be documented in this file.

The format is based on the Keep a Changelog principles and is intended to work with automated release tooling.

## [2.1.0](https://github.com/TerrorSquad/forge/compare/v2.0.1...v2.1.0) (2026-07-18)


### ⚠ BREAKING CHANGES

* a hook that references an uninstalled tool now fails instead of skipping it. Set SKIP_<TOOL>=1 or remove the tool from forge.toml to restore the previous pass-through behavior.

### Features

* detect down containers in preflight (+ pin release to 2.1.0) ([69e505f](https://github.com/TerrorSquad/forge/commit/69e505f211dee19f9c9af2df7be279307dbbf19b))
* detect down containers in preflight with a clear error ([b78861a](https://github.com/TerrorSquad/forge/commit/b78861a29dac44c1b214158169e836fdd36a2b3b))
* fail the hook when a configured tool binary is missing ([2d44d80](https://github.com/TerrorSquad/forge/commit/2d44d8022b153195ba72e41aee62ef8023a652c9))


### Bug Fixes

* honor run options on pre-push and only stage outputs on success ([1e351ca](https://github.com/TerrorSquad/forge/commit/1e351caea8e17b422995381f777c2ee27e69ba7d))
* remove `when` from config schema to match decoder ([36a6277](https://github.com/TerrorSquad/forge/commit/36a6277b03b0338d2228ca486e1ada8a638e5753))
* runner correctness bugs, cleanup, tests, and architecture docs ([8bed5cf](https://github.com/TerrorSquad/forge/commit/8bed5cfee20ad90655de5359e9f1be0758ff7786))


### Performance Improvements

* memoize container-running checks within a run ([b18b901](https://github.com/TerrorSquad/forge/commit/b18b901fa019601f7dc58952a3f7747fc92bd54e))

## [2.0.1](https://github.com/TerrorSquad/forge/compare/v2.0.0...v2.0.1) (2026-07-17)


### Bug Fixes

* resolve bare command for container backends ([1fe3781](https://github.com/TerrorSquad/forge/commit/1fe378164e5cbc8107c401cb3371814e045d5350))
* resolve bare command for container backends ([99773d8](https://github.com/TerrorSquad/forge/commit/99773d8c630f276b91ed6592d87371902c90dc4a))

## [2.0.0](https://github.com/TerrorSquad/forge/compare/v1.2.4...v2.0.0) (2026-06-13)


### ⚠ BREAKING CHANGES

* removed the `safe_stash` hook config option and the FORGE_NO_STASH environment variable.

### Code Refactoring

* remove pre-commit safety stash ([0a6445b](https://github.com/TerrorSquad/forge/commit/0a6445b708f568ca07c298806d4cfb004ef759ca))

## [1.2.4](https://github.com/TerrorSquad/forge/compare/v1.2.3...v1.2.4) (2026-06-09)


### Bug Fixes

* ddev workdir, dead BinaryExists, configurable ticket/type policy ([4c0f19d](https://github.com/TerrorSquad/forge/commit/4c0f19dedca44cf235750b7dfe77eff8c3882df7))
* **install:** resolve mise shims path via MISE_DATA_DIR and XDG_DATA_HOME ([b23a143](https://github.com/TerrorSquad/forge/commit/b23a143d0eac47019cefa7d1763b3a8b96b337f4))

## [1.2.3](https://github.com/TerrorSquad/forge/compare/v1.2.2...v1.2.3) (2026-06-09)


### Bug Fixes

* six bug fixes in runner, cache, workspace, and tool filter ([d6f2067](https://github.com/TerrorSquad/forge/commit/d6f2067d615aee2e90dcd5ada4251a2136cf9d58))
* six bug fixes in runner, cache, workspace, and tool filter ([5ab5850](https://github.com/TerrorSquad/forge/commit/5ab5850a96eebe53a0bdbae0397b9c00ba4792c1))

## [1.2.2](https://github.com/TerrorSquad/forge/compare/v1.2.1...v1.2.2) (2026-06-09)


### Bug Fixes

* support ** glob patterns in exclude_patterns via doublestar ([30aaed2](https://github.com/TerrorSquad/forge/commit/30aaed22ba413137a554f23a2dcbb618e3c2d416))

## [1.2.1](https://github.com/TerrorSquad/forge/compare/v1.2.0...v1.2.1) (2026-06-08)


### Bug Fixes

* load mise shims ([c578d1b](https://github.com/TerrorSquad/forge/commit/c578d1b49926980390aa8c3aaef353975973d8ea))
* load mise shims ([dde9821](https://github.com/TerrorSquad/forge/commit/dde9821de93aebe46124fa403090363c41611794))

## [1.2.0](https://github.com/TerrorSquad/forge/compare/v1.1.1...v1.2.0) (2026-06-03)


### Features

* add --skip-group and --verbose to forge run, implement skip-group filtering ([5b1f13b](https://github.com/TerrorSquad/forge/commit/5b1f13b6a2d9ada76db2395387da5b0692bc0a5f))
* **config:** enforce strict forge.toml validation and document declaration-order tool execution ([de16fd9](https://github.com/TerrorSquad/forge/commit/de16fd9e461ea4ee2f872d3290f279a1094ed474))
* **doctor:** report invalid repo config and validation issues in doctor diagnostics ([ac44811](https://github.com/TerrorSquad/forge/commit/ac448118b9a1249d76a68e1ea3fe4cdc2fe22595))


### Bug Fixes

* **config:** preserve tool declaration order from forge.toml ([3013dc3](https://github.com/TerrorSquad/forge/commit/3013dc312111f84bbb2b065b82fb6ab5c52d4be5))
* skip parallel pre-commit tools when no matching staged files ([affccab](https://github.com/TerrorSquad/forge/commit/affccab28f3f3a12b3d42c19fd0ed803764cee03))

## [1.1.1](https://github.com/TerrorSquad/forge/compare/v1.1.0...v1.1.1) (2026-06-03)


### Bug Fixes

* skip hooks during git merge/rebase/cherry-pick/revert operations ([0fa2852](https://github.com/TerrorSquad/forge/commit/0fa28527e48dbe95a2c6115883734c9b9c491789))
* skip pre-commit tools when no staged files match extensions even with pass_files=false ([874fd52](https://github.com/TerrorSquad/forge/commit/874fd52f374d4def2a2173dee1095d4651d1e6de))

## [1.1.0](https://github.com/TerrorSquad/forge/compare/v1.0.11...v1.1.0) (2026-05-27)


### Features

* default installer path to ~/.local/bin for non-root installs ([ef51e63](https://github.com/TerrorSquad/forge/commit/ef51e639fb4b086d69f736d11876cf898d59d41c))


### Bug Fixes

* create install directory when installing forge binary ([298729d](https://github.com/TerrorSquad/forge/commit/298729d4a44283189ce266a4df46013e07c6ccfa))
* create install directory when installing forge binary ([0e9f4c5](https://github.com/TerrorSquad/forge/commit/0e9f4c5219f395a729592f000447b46edb04bba6))

## [1.0.11](https://github.com/TerrorSquad/forge/compare/v1.0.10...v1.0.11) (2026-05-26)


### Bug Fixes

* install smoke tests and dist build output ([423417e](https://github.com/TerrorSquad/forge/commit/423417e4ccd27f3c017174b640ad28c4ed950753))

## [1.0.10](https://github.com/TerrorSquad/forge/compare/v1.0.9...v1.0.10) (2026-05-26)

### Bug Fixes

* install smoke tests and dist build output ([423417e](https://github.com/TerrorSquad/forge/commit/423417e4ccd27f3c017174b640ad28c4ed950753))
