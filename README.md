# semver

A minimal and fast Go library for working with
[Semantic Versioning (SemVer2.0.0)][semver].

## Features

* Accepts versions **with or without** the `v` prefix (`1.2.3` / `v1.2.3`).
* Supports shorthands: `MAJOR` and `MAJOR.MINOR` → normalized to
  `vMAJOR.0.0` and `vMAJOR.MINOR.0`. Prerelease/build require full
  `MAJOR.MINOR.PATCH`.
* **Canonical form**: `vMAJOR.MINOR.PATCH[-PRERELEASE]` (build metadata
  stripped).
* `Semver` struct with `Major`, `Minor`, `Patch`, `Prerelease`, `Build`,
  `Original`, and rich helpers.
* Correct comparison/sorting per SemVer (build is ignored). Stable tie-break
  by the original string.
* Mutators: `BumpPatch/Minor/Major`, `WithPre`, `WithBuild`, `StripPre`,
  `StripBuild`, `NextPrerelease`, plus `IsGreater/IsLower/IsEqual`.
* Performance: `Parse`/`Compare` **0 allocs/op** (amd64, Go 1.18+).
  Rendering \~ **1 alloc**.

> `Prerelease` and `Build` are zero-copy slices of `Original` right after
> `Parse`, after mutators they may be standalone strings.

## Differences vs [`golang.org/x/mod/semver`][xsemver]

* Optional leading `v` on input.
* Shorthands `MAJOR`, `MAJOR.MINOR`.
* A convenient `Semver` type + methods (mutators, rendering).
* Numeric components must fit into host `int`; otherwise the version is
  **invalid**.

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

func must(v semver.Semver, ok bool) semver.Semver {
 if !ok { panic("bad semver") }
 return v
}

func main() {
  v1 := must(semver.Parse("1.2.3-rc.1+meta"))
  v2 := must(semver.Parse("v1.2.3"))

  fmt.Println(v1.Canonical()) // v1.2.3-rc.1
  fmt.Println(v2.Major)       // 1
  fmt.Println(v2.MajorStr())  // v1

  // Compare (build ignored; release > prerelease for the same core)
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

## Mutators & Rendering

```go
v := must(semver.Parse("1.2.3-rc.1+build.5"))

p, _ := v.BumpPatch() // v1.2.4
m, _ := v.BumpMinor() // v1.3.0
M, _ := v.BumpMajor() // v2.0.0
fmt.Println(p.Canonical(), m.Canonical(), M.Canonical())

// WithPre / WithBuild (validated per SemVer)
x := must(semver.Parse("1"))
x1, _ := x.WithPre("alpha.1")   // v1.0.0-alpha.1
x2, _ := x1.WithBuild("meta.5") // v1.0.0-alpha.1+meta.5 (in Full)
fmt.Println(x1.Canonical(), x2.Full(true))

// NextPrerelease: bumps last numeric id, or appends ".1"; empty base defaults to "rc"
z := must(semver.Parse("1.2.3-rc.9"))
z1, _ := z.NextPrerelease("rc") // v1.2.3-rc.10
z2 := must(semver.Parse("1.2.3"))
z2, _ = z2.NextPrerelease("")   // v1.2.3-rc.1
fmt.Println(z1.Canonical(), z2.Canonical())

// Strip
r1, _ := v.StripPre()   // canonical: v1.2.3 ; Full can include +build if present
r2, _ := v.StripBuild() // canonical: v1.2.3-rc.1
_ = r1
_ = r2

// Rendering with/without original 'v' preservation
u := must(semver.Parse("V1.2.3-rc.1+meta"))

// Full renders "([v|V]?)MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]":
fmt.Println(u.Full(true))  // "v1.2.3-rc.1+meta"  (force lowercase 'v')
fmt.Println(u.Full(false)) // "V1.2.3-rc.1+meta" (preserve original prefix style)
// Canonical always uses 'v' and drops build:
fmt.Println(u.Canonical()) // "v1.2.3-rc.1"
```

## Compatibility

* **Comparison**: strict SemVer; build metadata does not affect ordering.
* **Canonical**: always starts with `v`, build metadata removed.
* **Shorthands**: `MAJOR`, `MAJOR.MINOR` accepted (pragmatic deviation).
* **Numbers**: must fit into host `int`; too large → invalid.
* Go: tested with Go 1.18+.

## API Cheatsheet

* Compare: `Compare`, `IsGreater`, `IsLower`, `IsEqual`.
* Render:
  * `Canonical()` → `vX.Y.Z[-pre]` (no build),
  * `Full(preserve bool)` → `([v|V]?)X.Y.Z[-pre][+build]`
    * `preserve == true` → force lowercase `'v'`
    * `preserve == false` → preserve original `'v'/'V'` or no prefix
* Slices: `MajorStr()`, `MajorMinorStr()`, `ReleaseStr()`.
* Flags: `HasV()`, `IsRelease()`,
  `HasMajor/HasMinor/HasPatch/HasPre/HasBuild()`.

[semver]: https://semver.org/
[xsemver]: https://pkg.go.dev/golang.org/x/mod/semver
