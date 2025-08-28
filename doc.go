// Package semver provides parsing and comparison of Semantic Versioning
// strings (SemVer 2.0.0) with a small set of pragmatic changes:
//
//   - Input may be given with or without the leading "v" (e.g. "1.2.3" or "v1.2.3").
//   - A parsed struct (Semver) exposes numeric fields (Major, Minor, Patch),
//     prerelease and build metadata, and canonical string form.
//   - Comparisons ignore build metadata, as required by SemVer.
//   - Sorting uses semver precedence with a lexicographic tie-breaker on the
//     original input to produce stable, deterministic order.
//
// Canonical form returned by Semver.Canon() always uses the "v" prefix
// and strips build metadata: "vMAJOR.MINOR.PATCH[-PRERELEASE]".
package semver
