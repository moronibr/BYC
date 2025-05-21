package version

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Version represents a semantic version
type Version struct {
	Major int
	Minor int
	Patch int
	Pre   string
	Build string
}

// VersionInfo represents detailed version information
type VersionInfo struct {
	Version     Version `json:"version"`
	BuildTime   string  `json:"build_time"`
	GitCommit   string  `json:"git_commit"`
	GoVersion   string  `json:"go_version"`
	Environment string  `json:"environment"`
}

// VersionManager handles version-related operations
type VersionManager struct {
	currentVersion Version
	versionInfo    VersionInfo
	upgradePath    map[Version][]Version
}

// NewVersionManager creates a new version manager
func NewVersionManager(current Version, info VersionInfo) *VersionManager {
	return &VersionManager{
		currentVersion: current,
		versionInfo:    info,
		upgradePath:    make(map[Version][]Version),
	}
}

// ParseVersion parses a version string
func ParseVersion(v string) (Version, error) {
	parts := strings.Split(v, ".")
	if len(parts) < 3 {
		return Version{}, fmt.Errorf("invalid version format: %s", v)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version: %s", parts[2])
	}

	version := Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}

	// Handle pre-release version
	if len(parts) > 3 {
		version.Pre = parts[3]
	}

	// Handle build metadata
	if len(parts) > 4 {
		version.Build = parts[4]
	}

	return version, nil
}

// String returns the string representation of a version
func (v Version) String() string {
	version := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Pre != "" {
		version += "-" + v.Pre
	}
	if v.Build != "" {
		version += "+" + v.Build
	}
	return version
}

// Compare compares two versions
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		return v.Major - other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor - other.Minor
	}
	if v.Patch != other.Patch {
		return v.Patch - other.Patch
	}
	return strings.Compare(v.Pre, other.Pre)
}

// IsCompatible checks if two versions are compatible
func (v Version) IsCompatible(other Version) bool {
	// Major version must match for compatibility
	return v.Major == other.Major
}

// GetUpgradePath returns the upgrade path to a target version
func (vm *VersionManager) GetUpgradePath(target Version) ([]Version, error) {
	if !vm.currentVersion.IsCompatible(target) {
		return nil, fmt.Errorf("incompatible version: %s", target)
	}

	if vm.currentVersion.Compare(target) >= 0 {
		return nil, fmt.Errorf("target version %s is not newer than current version %s", target, vm.currentVersion)
	}

	path := vm.upgradePath[vm.currentVersion]
	if len(path) == 0 {
		return nil, fmt.Errorf("no upgrade path found from %s to %s", vm.currentVersion, target)
	}

	return path, nil
}

// RegisterUpgradePath registers an upgrade path
func (vm *VersionManager) RegisterUpgradePath(from Version, to []Version) {
	vm.upgradePath[from] = to
}

// GetVersionInfo returns the current version information
func (vm *VersionManager) GetVersionInfo() VersionInfo {
	return vm.versionInfo
}

// ToJSON returns the version information as JSON
func (vi VersionInfo) ToJSON() (string, error) {
	data, err := json.Marshal(vi)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON parses version information from JSON
func FromJSON(data string) (VersionInfo, error) {
	var vi VersionInfo
	err := json.Unmarshal([]byte(data), &vi)
	return vi, err
}

// Constants for version comparison
const (
	VersionEqual = iota
	VersionGreater
	VersionLess
)

// CompareVersions compares two version strings
func CompareVersions(v1, v2 string) (int, error) {
	ver1, err := ParseVersion(v1)
	if err != nil {
		return 0, err
	}

	ver2, err := ParseVersion(v2)
	if err != nil {
		return 0, err
	}

	return ver1.Compare(ver2), nil
}

// IsVersionCompatible checks if two version strings are compatible
func IsVersionCompatible(v1, v2 string) (bool, error) {
	ver1, err := ParseVersion(v1)
	if err != nil {
		return false, err
	}

	ver2, err := ParseVersion(v2)
	if err != nil {
		return false, err
	}

	return ver1.IsCompatible(ver2), nil
}
