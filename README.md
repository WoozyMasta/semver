# semver

A minimal and fast Go library for working with [Semantic Versioning (SemVer 2.0.0)][semver].

## Features

- Accepts versions **with or without** the `v` prefix (`1.2.3` and `v1.2.3`).
- Provides a **canonical form**: `vMAJOR.MINOR.PATCH[-PRERELEASE]`
  (build metadata is stripped).
- Exposes a structured `Semver` type with `Major`, `Minor`, `Patch`,
  `Prerelease`, `Build`, `Canonical`, and `Original` fields.
- Correct comparison and sorting according to SemVer precedence rules.
- Stable deterministic ordering (ties broken by the original input string).

## Differences from [`golang.org/x/mod/semver`][xsemver]

- Leading `v` is optional on input.
- Offers a convenient `Semver` struct and methods instead of plain
  string functions.
- API designed for practical use in CLI tools and automation.

## Installation

```bash
go get github.com/woozymasta/semver
```

## Quick Start

```go
package main

import (
 "fmt"

 "github.com/woozymasta/semver"
)

func main() {
  v1, _ := semver.Parse("1.2.3-rc.1+meta")
  v2, _ := semver.Parse("v1.2.3")

  fmt.Println(v1.Canon())         // v1.2.3-rc.1
  fmt.Println(v2.Major)           // 1
  fmt.Println(v2.MajorStr())      // v1
  fmt.Println(v2.MajorMinorStr()) // v1.2

  switch v1.Compare(v2) {
  case -1:
    fmt.Println("v1 < v2")
  case 0:
    fmt.Println("v1 == v2")
  case 1:
    fmt.Println("v1 > v2")
  }
}
```

### Sorting

```go
versions := []semver.Semver{
  must(semver.Parse("1.2.3")),
  must(semver.Parse("1.2.3-rc.1")),
  must(semver.Parse("2.0.0")),
}
semver.List(versions).Sort()
```

## Compatibility

- Comparison: strictly follows SemVer rules, ignoring build metadata.
- Canonical form: always prefixed with v; build metadata is removed.
- Go version: tested on Go 1.21+, but should work with earlier releases
  (except for slices in tests).

[semver]: https://semver.org/
[xsemver]: https://pkg.go.dev/golang.org/x/mod/semver
