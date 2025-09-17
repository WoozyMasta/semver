package semver

// Flags is a bitmask describing parse state / presence of components.
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

// Convenience helpers (no branches in hot paths):
func (v Semver) HasV() bool {
	return v.Flags&FlagHasV != 0
}

func (v Semver) IsRelease() bool {
	return v.Flags&(FlagHasPre|FlagHasBuild) == 0
}

func (v Semver) HasMajor() bool {
	return v.Flags&FlagHasMajor != 0
}

func (v Semver) HasMinor() bool {
	return v.Flags&FlagHasMinor != 0
}

func (v Semver) HasPatch() bool {
	return v.Flags&FlagHasPatch != 0
}

func (v Semver) HasPre() bool {
	return v.Flags&FlagHasPre != 0
}

func (v Semver) HasBuild() bool {
	return v.Flags&FlagHasBuild != 0
}
