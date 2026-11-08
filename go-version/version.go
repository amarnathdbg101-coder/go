package goversion

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a parsed Semantic Version (SemVer 2.0.0).
type Version struct {
	Major         int
	Minor         int
	Patch         int
	Prerelease    string
	BuildMetadata string
	raw           string
}

// Parse parses a semantic version string (e.g. "1.2.3", "v2.0.0-beta.1+exp.sha.5114f85").
func Parse(v string) (*Version, error) {
	raw := strings.TrimSpace(v)
	s := strings.TrimPrefix(raw, "v")
	s = strings.TrimPrefix(s, "V")

	// Split build metadata
	var build string
	if idx := strings.Index(s, "+"); idx != -1 {
		build = s[idx+1:]
		s = s[:idx]
	}

	// Split prerelease
	var pre string
	if idx := strings.Index(s, "-"); idx != -1 {
		pre = s[idx+1:]
		s = s[:idx]
	}

	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("%w: %s", ErrInvalidVersion, raw)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil || major < 0 {
		return nil, fmt.Errorf("%w: major %q", ErrInvalidVersion, parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil || minor < 0 {
		return nil, fmt.Errorf("%w: minor %q", ErrInvalidVersion, parts[1])
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil || patch < 0 {
		return nil, fmt.Errorf("%w: patch %q", ErrInvalidVersion, parts[2])
	}

	return &Version{
		Major:         major,
		Minor:         minor,
		Patch:         patch,
		Prerelease:    pre,
		BuildMetadata: build,
		raw:           raw,
	}, nil
}

// MustParse parses a version string and panics if invalid.
func MustParse(v string) *Version {
	ver, err := Parse(v)
	if err != nil {
		panic(err)
	}
	return ver
}

// String returns formatted SemVer string.
func (v *Version) String() string {
	res := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		res = fmt.Sprintf("%s-%s", res, v.Prerelease)
	}
	if v.BuildMetadata != "" {
		res = fmt.Sprintf("%s+%s", res, v.BuildMetadata)
	}
	return res
}

// Compare compares v to other following SemVer 2.0.0 precedence rules.
// Returns -1 if v < other, 0 if v == other, 1 if v > other.
func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}

	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}

	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}

	// Precedence rule: normal version > pre-release version
	if v.Prerelease == "" && other.Prerelease != "" {
		return 1
	}
	if v.Prerelease != "" && other.Prerelease == "" {
		return -1
	}
	if v.Prerelease == "" && other.Prerelease == "" {
		return 0
	}

	return comparePrerelease(v.Prerelease, other.Prerelease)
}

func comparePrerelease(preA, preB string) int {
	partsA := strings.Split(preA, ".")
	partsB := strings.Split(preB, ".")

	minLen := len(partsA)
	if len(partsB) < minLen {
		minLen = len(partsB)
	}

	for i := 0; i < minLen; i++ {
		a := partsA[i]
		b := partsB[i]

		numA, errA := strconv.Atoi(a)
		numB, errB := strconv.Atoi(b)

		if errA == nil && errB == nil {
			if numA != numB {
				if numA < numB {
					return -1
				}
				return 1
			}
		} else if errA == nil && errB != nil {
			// Numeric identifiers have lower precedence than non-numeric
			return -1
		} else if errA != nil && errB == nil {
			return 1
		} else {
			// Both are non-numeric strings
			if a != b {
				if a < b {
					return -1
				}
				return 1
			}
		}
	}

	if len(partsA) < len(partsB) {
		return -1
	}
	if len(partsA) > len(partsB) {
		return 1
	}
	return 0
}

// GreaterThan checks if v > other.
func (v *Version) GreaterThan(other *Version) bool {
	return v.Compare(other) > 0
}

// GreaterThanOrEqual checks if v >= other.
func (v *Version) GreaterThanOrEqual(other *Version) bool {
	return v.Compare(other) >= 0
}

// LessThan checks if v < other.
func (v *Version) LessThan(other *Version) bool {
	return v.Compare(other) < 0
}

// LessThanOrEqual checks if v <= other.
func (v *Version) LessThanOrEqual(other *Version) bool {
	return v.Compare(other) <= 0
}

// Equal checks if v == other (ignores build metadata as per SemVer 2.0.0).
func (v *Version) Equal(other *Version) bool {
	return v.Compare(other) == 0
}
