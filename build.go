package buildcard

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

const (
	defaultVersion = "DEV"
	unknownValue   = "unknown"
	timeLayout     = "2006-01-02 15:04:05 -0700 MST"
)

var (
	buildVersion     = ""
	buildCommitShort = "unknown"
	buildCommitHash  = ""
	buildCommitDate  = ""
	buildDirty       = ""
	buildTime        = ""
	buildRepository  = ""
)

// Build contains metadata for the running application. Treat it as read-only.
var Build = readBuildInfo()

// BuildInfo describes the build of the running application.
type BuildInfo struct {
	shortHash     string
	version       string
	commitHash    string
	commitDate    time.Time
	buildTime     time.Time
	repository    string
	isDirtyCommit bool
}

func readBuildInfo() BuildInfo {
	info, _ := debug.ReadBuildInfo()
	return buildInfoFromMetadata(info)
}

func buildInfoFromMetadata(info *debug.BuildInfo) BuildInfo {
	b := BuildInfo{version: defaultVersion, commitHash: unknownValue, repository: unknownValue}
	if info != nil {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			b.version = info.Main.Version
		}
		b.repository = normalizeText(info.Main.Path)
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				b.commitHash = normalizeText(setting.Value)
			case "vcs.time":
				b.commitDate = normalizeDate(setting.Value)
			case "vcs.modified":
				b.isDirtyCommit = parseBool(setting.Value)
			}
		}
	}
	if v := strings.TrimSpace(buildVersion); v != "" {
		b.version = v
	}
	if v := strings.TrimSpace(buildCommitHash); v != "" {
		b.commitHash = v
	}
	if v := strings.TrimSpace(buildCommitShort); v != "" && v != unknownValue {
		b.shortHash = v
	}
	if strings.TrimSpace(buildCommitDate) != "" {
		b.commitDate = normalizeDate(buildCommitDate)
	}
	if strings.TrimSpace(buildDirty) != "" {
		b.isDirtyCommit = parseBool(buildDirty)
	}
	if strings.TrimSpace(buildRepository) != "" {
		b.repository = normalizeText(buildRepository)
	}
	b.buildTime = normalizeDate(buildTime)
	return b
}

func normalizeText(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return unknownValue
	}
	return v
}

func parseBool(v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	return v == "1" || v == "true" || v == "yes" || v == "y"
}

func normalizeDate(v string) time.Time {
	v = strings.TrimSpace(v)
	if v == "" || v == "unknown" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t
	}
	if unix, err := strconv.ParseInt(v, 10, 64); err == nil {
		return time.Unix(unix, 0).Local()
	}
	return time.Time{}
}

func formatTimeWithTZ(t time.Time) string {
	if t.IsZero() {
		return unknownValue
	}
	return t.Local().Format(timeLayout)
}

// MarshalJSON implements json.Marshaler with UTC timestamps.
func (b BuildInfo) MarshalJSON() ([]byte, error) {
	type out struct {
		Version       string `json:"version"`
		CommitHash    string `json:"commit_hash"`
		CommitDate    string `json:"commit_date"`
		IsDirtyCommit bool   `json:"is_dirty_commit"`
		BuildTime     string `json:"build_time"`
		Repository    string `json:"repository"`
	}
	o := out{
		Version:       b.version,
		CommitHash:    b.commitHash,
		IsDirtyCommit: b.isDirtyCommit,
		Repository:    b.repository,
	}
	if !b.commitDate.IsZero() {
		o.CommitDate = b.commitDate.UTC().Format(time.RFC3339)
	}
	if !b.buildTime.IsZero() {
		o.BuildTime = b.buildTime.UTC().Format(time.RFC3339)
	}
	return json.Marshal(o)
}

// Version returns the application version, or DEV.
func (b BuildInfo) Version() string { return b.version }

// CommitHash returns the full commit hash, or unknown.
func (b BuildInfo) CommitHash() string { return b.commitHash }

// ShortCommitHash returns the explicit short hash or the first seven characters of the full hash.
func (b BuildInfo) ShortCommitHash() string {
	if b.shortHash != "" {
		return b.shortHash
	}
	if len(b.commitHash) <= 7 {
		return b.commitHash
	}
	return b.commitHash[:7]
}

// CommitDate returns the commit timestamp, or zero when unavailable.
func (b BuildInfo) CommitDate() time.Time { return b.commitDate }

// IsDirty reports whether the application was built with uncommitted changes.
func (b BuildInfo) IsDirty() bool { return b.isDirtyCommit }

// BuildTime returns the explicit build timestamp, or zero when unavailable.
func (b BuildInfo) BuildTime() time.Time { return b.buildTime }

// Repository returns the repository override or main module path.
func (b BuildInfo) Repository() string { return b.repository }

// String returns a compact build description.
func (b BuildInfo) String() string {
	date := formatTimeWithTZ(b.commitDate)
	buildTime := formatTimeWithTZ(b.buildTime)
	dirty := ""
	if b.isDirtyCommit {
		dirty = "*"
	}
	return fmt.Sprintf("%s - %s | %s%s | build %s", b.version, date, b.ShortCommitHash(), dirty, buildTime)
}
