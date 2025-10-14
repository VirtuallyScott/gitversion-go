package gitversion

import (
	"encoding/json"
	"fmt"

	"github.com/VirtuallyScott/gitversion-go/pkg/semver"
)

type OutputFormat string

const (
	Text               OutputFormat = "text"
	JSON               OutputFormat = "json"
	AssemblySemVer     OutputFormat = "AssemblySemVer"
	AssemblySemFileVer OutputFormat = "AssemblySemFileVer"
)

type JSONOutput struct {
	Major                     int    `json:"Major"`
	Minor                     int    `json:"Minor"`
	Patch                     int    `json:"Patch"`
	PreReleaseTag             string `json:"PreReleaseTag"`
	PreReleaseTagWithDash     string `json:"PreReleaseTagWithDash"`
	BuildMetaData             string `json:"BuildMetaData"`
	BuildMetaDataPadded       string `json:"BuildMetaDataPadded"`
	FullBuildMetaData         string `json:"FullBuildMetaData"`
	MajorMinorPatch           string `json:"MajorMinorPatch"`
	SemVer                    string `json:"SemVer"`
	AssemblySemVer            string `json:"AssemblySemVer"`
	AssemblySemFileVer        string `json:"AssemblySemFileVer"`
	FullSemVer                string `json:"FullSemVer"`
	InformationalVersion      string `json:"InformationalVersion"`
	BranchName                string `json:"BranchName"`
	EscapedBranchName         string `json:"EscapedBranchName"`
	Sha                       string `json:"Sha"`
	ShortSha                  string `json:"ShortSha"`
	NuGetVersionV2            string `json:"NuGetVersionV2"`
	NuGetVersion              string `json:"NuGetVersion"`
	VersionSourceSha          string `json:"VersionSourceSha"`
	CommitsSinceVersionSource int    `json:"CommitsSinceVersionSource"`
	CommitDate                string `json:"CommitDate"`
}

type Formatter struct {
	repo Repository
}

func NewFormatter(repo Repository) *Formatter {
	return &Formatter{repo: repo}
}

func (f *Formatter) Format(version *semver.Version, format OutputFormat, branch string) (string, error) {
	switch format {
	case Text:
		return version.String(), nil
	case AssemblySemVer:
		return version.AssemblySemVer(), nil
	case AssemblySemFileVer:
		return version.AssemblySemFileVer(), nil
	case JSON:
		return f.formatJSON(version, branch)
	default:
		return "", fmt.Errorf("unknown output format: %s", format)
	}
}

func (f *Formatter) formatJSON(version *semver.Version, branch string) (string, error) {
	sha, _ := f.repo.GetSHA()
	shortSha, _ := f.repo.GetShortSHA()
	commitDate, _ := f.repo.GetCommitDate()
	latestTag, _ := f.repo.GetLatestTag()
	commitCount, _ := f.repo.GetCommitCountSinceTag(latestTag)

	preReleaseWithDash := ""
	if version.PreRelease != "" {
		preReleaseWithDash = "-" + version.PreRelease
	}

	buildMetaDataPadded := ""
	if version.Build != "" {
		buildMetaDataPadded = "+" + version.Build
	}

	output := JSONOutput{
		Major:                     version.Major,
		Minor:                     version.Minor,
		Patch:                     version.Patch,
		PreReleaseTag:             version.PreRelease,
		PreReleaseTagWithDash:     preReleaseWithDash,
		BuildMetaData:             version.Build,
		BuildMetaDataPadded:       buildMetaDataPadded,
		FullBuildMetaData:         version.Build,
		MajorMinorPatch:           version.MajorMinorPatch(),
		SemVer:                    version.String(),
		AssemblySemVer:            version.AssemblySemVer(),
		AssemblySemFileVer:        version.AssemblySemFileVer(),
		FullSemVer:                version.String(),
		InformationalVersion:      version.String(),
		BranchName:                branch,
		EscapedBranchName:         semver.SanitizeBranchName(branch),
		Sha:                       sha,
		ShortSha:                  shortSha,
		NuGetVersionV2:            version.String(),
		NuGetVersion:              version.String(),
		VersionSourceSha:          sha,
		CommitsSinceVersionSource: commitCount,
		CommitDate:                commitDate,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(data), nil
}
