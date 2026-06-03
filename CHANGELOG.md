# Changelog

All notable changes to this project will be documented in this file.

The format is based on the Keep a Changelog principles and is intended to work with automated release tooling.

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
