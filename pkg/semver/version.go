package semver

import (
	"fmt"
	"regexp"
	"strconv"
)

type Version struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
	Build      string
}

var semverPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z0-9.-]+))?(?:\+([a-zA-Z0-9.+-]+))?$`)

func Parse(version string) (*Version, error) {
	matches := semverPattern.FindStringSubmatch(version)
	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid semver format: %s", version)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", matches[1])
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", matches[2])
	}

	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %s", matches[3])
	}

	return &Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		PreRelease: matches[4],
		Build:      matches[5],
	}, nil
}

func (v *Version) String() string {
	version := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	
	if v.PreRelease != "" {
		version += "-" + v.PreRelease
	}
	
	if v.Build != "" {
		version += "+" + v.Build
	}
	
	return version
}

func (v *Version) IncrementMajor() {
	v.Major++
	v.Minor = 0
	v.Patch = 0
}

func (v *Version) IncrementMinor() {
	v.Minor++
	v.Patch = 0
}

func (v *Version) IncrementPatch() {
	v.Patch++
}

func (v *Version) AssemblySemVer() string {
	return fmt.Sprintf("%d.%d.%d.0", v.Major, v.Minor, v.Patch)
}

func (v *Version) AssemblySemFileVer() string {
	return fmt.Sprintf("%d.%d.%d.0", v.Major, v.Minor, v.Patch)
}

func (v *Version) MajorMinorPatch() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func SanitizeBranchName(branch string) string {
	return regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(branch, "-")
}