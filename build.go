package buildcard

import (
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

const (
	defaultVersion = "DEV"
	unknownValue   = "unknown"
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

func readBuildInfo() Info {
	info, _ := debug.ReadBuildInfo()
	return buildInfoFromMetadata(info)
}

func buildInfoFromMetadata(info *debug.BuildInfo) Info {
	b := Info{Version: defaultVersion, CommitHash: unknownValue, Repository: unknownValue, CommitDate: unknownValue, BuildTime: unknownValue}
	if info != nil {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			b.Version = info.Main.Version
		}
		b.Repository = normalizeText(info.Main.Path)
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				b.CommitHash = normalizeText(setting.Value)
			case "vcs.time":
				b.CommitDate = formatTime(normalizeDate(setting.Value))
			case "vcs.modified":
				b.Dirty = parseBool(setting.Value)
			}
		}
	}
	if v := strings.TrimSpace(buildVersion); v != "" {
		b.Version = v
	}
	if v := strings.TrimSpace(buildCommitHash); v != "" {
		b.CommitHash = v
	}
	if v := strings.TrimSpace(buildCommitShort); v != "" && v != unknownValue {
		b.ShortHash = v
	}
	if strings.TrimSpace(buildCommitDate) != "" {
		b.CommitDate = formatTime(normalizeDate(buildCommitDate))
	}
	if strings.TrimSpace(buildDirty) != "" {
		b.Dirty = parseBool(buildDirty)
	}
	if strings.TrimSpace(buildRepository) != "" {
		b.Repository = normalizeText(buildRepository)
	}
	b.BuildTime = formatTime(normalizeDate(buildTime))
	if b.ShortHash == "" {
		b.ShortHash = b.CommitHash[:min(7, len(b.CommitHash))]
	}
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
