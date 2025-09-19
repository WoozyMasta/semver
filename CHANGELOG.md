<!-- markdownlint-disable no-duplicate-heading -->
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.2] - 2025-09-19

### Added

* unify version rendering via `PrintFlags` in `Print()`

## [0.2.0] - 2025-09-18

### Added

* `Bump*()` `Strip*()` `With*()` `Next*()` functions for version modification
* `Flags` type has been added to the version structure and reflects,
  as a bit flag, what the version consists of.

### Changed

* global refactoring of everything

## [0.1.2] - 2025-09-16

### Changed

* Downgraded go directive to `1.18` for better compatibility (no code changes)

## [0.1.1] - 2025-09-16

### Added

* `ParseNoCanon`: SemVer parser without Canonical assembly
  (less allocations, faster where Canonical is not needed).

## [0.1.0] - 2025-08-28

### Added

* Initial release

<!-- links -->

[0.2.2]: <https://github.com/WoozyMasta/semver/compare/v0.2.0...v0.2.2>
[0.2.0]: <https://github.com/WoozyMasta/semver/compare/v0.1.2...v0.2.0>
[0.1.2]: <https://github.com/WoozyMasta/semver/compare/v0.1.1...v0.1.2>
[0.1.1]: <https://github.com/WoozyMasta/semver/compare/v0.1.0...v0.1.1>
[0.1.0]: <https://github.com/WoozyMasta/semver/releases/tag/v0.1.0>
