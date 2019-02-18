package prefab

import (
	"fmt"

	"runtime"
	"strings"

	"github.com/spf13/cast"
)

// Version represents the prefab build version.
type Version struct {
	// Major and minor version.
	Number float32

	// Increment this for bug releases
	PatchLevel int

	// prefabVersionSuffix is the suffix used in the prefab version string.
	// It will be blank for release versions.
	Suffix string
}

func (v Version) String() string {
	return version(v.Number, v.PatchLevel, v.Suffix)
}

// Version returns the prefab version.
func (v Version) Version() VersionString {
	return VersionString(v.String())
}

// VersionString represents a prefab version string.
type VersionString string

func (h VersionString) String() string {
	return string(h)
}

// Compare implements the compare.Comparer interface.
func (h VersionString) Compare(other interface{}) int {
	v := MustParseVersion(h.String())
	return compareVersionsWithSuffix(v.Number, v.PatchLevel, v.Suffix, other)
}

// Eq implements the compare.Eqer interface.
func (h VersionString) Eq(other interface{}) bool {
	s, err := cast.ToStringE(other)
	if err != nil {
		return false
	}
	return s == h.String()
}

var versionSuffixes = []string{"-test", "-DEV"}

// ParseVersion parses a version string.
func ParseVersion(s string) (Version, error) {
	var vv Version
	for _, suffix := range versionSuffixes {
		if strings.HasSuffix(s, suffix) {
			vv.Suffix = suffix
			s = strings.TrimSuffix(s, suffix)
		}
	}

	v, p := parseVersion(s)

	vv.Number = v
	vv.PatchLevel = p

	return vv, nil
}

// MustParseVersion parses a version string
// and panics if any error occurs.
func MustParseVersion(s string) Version {
	vv, err := ParseVersion(s)
	if err != nil {
		panic(err)
	}
	return vv
}

// ReleaseVersion represents the release version.
func (v Version) ReleaseVersion() Version {
	v.Suffix = ""
	return v
}

// Next returns the next prefab release version.
func (v Version) Next() Version {
	return Version{Number: v.Number + 0.01}
}

// Prev returns the previous prefab release version.
func (v Version) Prev() Version {
	return Version{Number: v.Number - 0.01}
}

// NextPatchLevel returns the next patch/bugfix prefab version.
// This will be a patch increment on the previous prefab version.
func (v Version) NextPatchLevel(level int) Version {
	return Version{Number: v.Number - 0.01, PatchLevel: level}
}

// BuildVersionString creates a version string. This is what you see when
// running "prefab version".
func BuildVersionString() string {
	program := "prefab "

	version := "v" + CurrentVersion.String()
	if commitHash != "" {
		version += "-" + strings.ToUpper(commitHash)
	}

	osArch := runtime.GOOS + "/" + runtime.GOARCH

	date := buildDate
	if date == "" {
		date = "unknown"
	}

	return fmt.Sprintf("%s %s %s BuildDate: %s", program, version, osArch, date)

}

func version(version float32, patchVersion int, suffix string) string {
	if patchVersion > 0 || version > 0.53 {
		return fmt.Sprintf("%.2f.%d%s", version, patchVersion, suffix)
	}
	return fmt.Sprintf("%.2f%s", version, suffix)
}

// CompareVersion compares the given version string or number against the
// running prefab version.
// It returns -1 if the given version is less than, 0 if equal and 1 if greater than
// the running version.
func CompareVersion(version interface{}) int {
	return compareVersionsWithSuffix(CurrentVersion.Number, CurrentVersion.PatchLevel, CurrentVersion.Suffix, version)
}

func compareVersions(inVersion float32, inPatchVersion int, in interface{}) int {
	return compareVersionsWithSuffix(inVersion, inPatchVersion, "", in)
}

func compareVersionsWithSuffix(inVersion float32, inPatchVersion int, suffix string, in interface{}) int {
	var c int
	switch d := in.(type) {
	case float64:
		c = compareFloatVersions(inVersion, float32(d))
	case float32:
		c = compareFloatVersions(inVersion, d)
	case int:
		c = compareFloatVersions(inVersion, float32(d))
	case int32:
		c = compareFloatVersions(inVersion, float32(d))
	case int64:
		c = compareFloatVersions(inVersion, float32(d))
	default:
		s, err := cast.ToStringE(in)
		if err != nil {
			return -1
		}

		v, err := ParseVersion(s)
		if err != nil {
			return -1
		}

		if v.Number == inVersion && v.PatchLevel == inPatchVersion {
			return strings.Compare(suffix, v.Suffix)
		}

		if v.Number < inVersion || (v.Number == inVersion && v.PatchLevel < inPatchVersion) {
			return -1
		}

		return 1
	}

	if c == 0 && suffix != "" {
		return 1
	}

	return c
}

func parseVersion(s string) (float32, int) {
	var (
		v float32
		p int
	)

	if strings.Count(s, ".") == 2 {
		li := strings.LastIndex(s, ".")
		p = cast.ToInt(s[li+1:])
		s = s[:li]
	}

	v = float32(cast.ToFloat64(s))

	return v, p
}

func compareFloatVersions(version float32, v float32) int {
	if v == version {
		return 0
	}
	if v < version {
		return -1
	}
	return 1
}