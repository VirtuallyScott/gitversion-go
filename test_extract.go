package main

import (
	"fmt"
	"strings"
)

func extractReleaseName(branch string) string {
	parts := strings.Split(branch, "/")
	if len(parts) > 1 {
		versionPart := parts[len(parts)-1]
		if dashIndex := strings.LastIndex(versionPart, "-"); dashIndex != -1 && dashIndex < len(versionPart)-1 {
			prerelease := versionPart[dashIndex+1:]
			return prerelease
		}
	}
	return ""
}

func main() {
	branch := "release/0.0.2-alpha"
	result := extractReleaseName(branch)
	fmt.Printf("Branch: %s -> Release name: '%s'\n", branch, result)
}
