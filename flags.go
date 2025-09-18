package semver

// Flags is a compact bitmask describing which components were explicitly
// present in the input (e.g., MINOR/PATCH for shorthand detection).
type Flags uint8

// Flags represent bitwise flags for Semver parsing state.
const (
	FlagHasV     Flags = 1 << iota // input had leading 'v'/'V'
	FlagHasMajor                   // major component explicitly present (always true for valid)
	FlagHasMinor                   // minor explicitly present in input
	FlagHasPatch                   // patch explicitly present in input
	FlagHasPre                     // prerelease present
	FlagHasBuild                   // build metadata present
)

// HasV reports whether the input had a leading 'v' or 'V'.
// For invalid versions, falls back to inspecting Original.
func (v Semver) HasV() bool {
	if v.Flags&FlagHasV != 0 {
		return true
	}
	return len(v.Original) > 0 && (v.Original[0] == 'v' || v.Original[0] == 'V')
}

// IsRelease reports whether the version is a release (no prerelease/build).
// Always false for invalid versions.
func (v Semver) IsRelease() bool {
	return v.Valid && v.Flags&(FlagHasPre|FlagHasBuild) == 0
}

// HasMajor reports whether the major component was explicitly present in the input.
func (v Semver) HasMajor() bool {
	return v.Valid && v.Flags&FlagHasMajor != 0
}

// HasMinor reports whether the minor component was explicitly present in the input.
func (v Semver) HasMinor() bool {
	return v.Valid && v.Flags&FlagHasMinor != 0
}

// HasPatch reports whether the patch component was explicitly present in the input.
func (v Semver) HasPatch() bool {
	return v.Valid && v.Flags&FlagHasPatch != 0
}

// HasPre reports whether the prerelease component was present in the input.
func (v Semver) HasPre() bool {
	return v.Valid && v.Flags&FlagHasPre != 0
}

// HasBuild reports whether the build metadata component was present in the input.
func (v Semver) HasBuild() bool {
	return v.Valid && v.Flags&FlagHasBuild != 0
}
