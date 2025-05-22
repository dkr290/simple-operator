package controller

import (
	"sort"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
)

// sortVersions sorts the version strings (assuming formats like "v21", "v22") Only thoose are accepted in this application
func sortVersions(versions map[string]*appsv1.Deployment) []string {
	var vers []string
	for ver := range versions {
		vers = append(vers, ver)
	}
	sort.Slice(vers, func(i, j int) bool {
		return parseVersion(vers[i]) < parseVersion(vers[j])
	})
	return vers
}

// parseVersion converts a version string (e.g., "v22") by stripping the "v" and returning its integer part.
func parseVersion(ver string) int {
	trimmed := strings.TrimPrefix(ver, "v")
	num, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0
	}
	return num
}
