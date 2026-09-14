package buildcard

import (
	"errors"
	"runtime/debug"
	"testing"
	"time"
)

func TestBuildMetadata(t *testing.T) {
	old := []string{buildVersion, buildCommitShort, buildCommitHash, buildCommitDate, buildDirty, buildTime, buildRepository}
	t.Cleanup(func() {
		buildVersion, buildCommitShort, buildCommitHash, buildCommitDate, buildDirty, buildTime, buildRepository = old[0], old[1], old[2], old[3], old[4], old[5], old[6]
	})
	buildVersion, buildCommitShort, buildCommitHash, buildCommitDate, buildDirty, buildTime, buildRepository = "", "", "", "", "", "", ""

	b := buildInfoFromMetadata(nil)
	if b.Version != "DEV" || b.CommitHash != "unknown" || b.BuildTime != unknownValue {
		t.Fatalf("unexpected defaults: %s", b)
	}
	metadata := &debug.BuildInfo{
		Main: debug.Module{Path: "example.com/application", Version: "v1.2.3"},
		Deps: []*debug.Module{{Path: "github.com/LittleDrongo/buildcard", Version: "v9.9.9"}},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "123456789abcdef"},
			{Key: "vcs.time", Value: "2026-09-14T10:00:00Z"},
			{Key: "vcs.modified", Value: "true"},
		},
	}
	b = buildInfoFromMetadata(metadata)
	if b.Version != "v1.2.3" || b.Repository != "example.com/application" || b.ShortHash != "1234567" || !b.Dirty || b.CommitDate == unknownValue || b.BuildTime != unknownValue {
		t.Fatalf("unexpected metadata: %+v", b)
	}
	buildVersion, buildCommitShort, buildCommitHash = " v2.0.0 ", " release ", "abcdef123456789"
	buildCommitDate, buildDirty, buildTime, buildRepository = "0", "false", "2026-09-14T12:00:00Z", "https://example.com/repo"
	overridden := buildInfoFromMetadata(metadata)
	if overridden.Version != "v2.0.0" || overridden.ShortHash != "release" || overridden.CommitHash != buildCommitHash || overridden.Dirty || overridden.CommitDate != formatTime(time.Unix(0, 0)) || overridden.BuildTime != formatTime(normalizeDate(buildTime)) || overridden.Repository != buildRepository {
		t.Fatalf("overrides not applied: %+v", overridden)
	}
	if b.ShortHash != "1234567" {
		t.Fatal("existing metadata changed after changing linker variables")
	}
}

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestTables(t *testing.T) {
	errWrite := errors.New("write failed")
	if err := Snapshot().WriteTable(failingWriter{errWrite}); !errors.Is(err, errWrite) {
		t.Fatalf("lost write error: %v", err)
	}
	got := formatTable([][2]string{{"Я", "1"}, {"AB", "2"}})
	if got != "Я      1\nAB     2\n" {
		t.Fatalf("bad Unicode alignment: %q", got)
	}
}
