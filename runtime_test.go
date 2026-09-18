package buildcard

import (
	"errors"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"testing"
)

func TestUsernameFallback(t *testing.T) {
	for _, tc := range []struct {
		name string
		u    *user.User
		env  map[string]string
		uid  int
		want string
	}{
		{"system", &user.User{Username: "alice"}, map[string]string{"USER": "other"}, 12, "alice"},
		{"environment", nil, map[string]string{"USER": "app"}, 12, "app"},
		{"login", nil, map[string]string{"LOGNAME": "login"}, 12, "login"},
		{"windows", nil, map[string]string{"USERNAME": "win"}, -1, "win"},
		{"container", nil, nil, 789010, "uid=789010"},
		{"root", nil, nil, 0, "uid=0"},
		{"unavailable", nil, nil, -1, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			if tc.u == nil {
				err = errors.New("no passwd entry")
			}
			if got := usernameFrom(tc.u, err, func(k string) string { return tc.env[k] }, tc.uid); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLocalRepository(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git unavailable")
	}
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
	}
	git("init")
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module emal_processing\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := repositoryFromDir(dir, "emal_processing"); got != "" {
		t.Fatalf("without origin: %s", got)
	}
	want := "git@gitlab.services.mts.ru:clr_group_msk/email_invoice_stamp_cleaner.git"
	git("remote", "add", "origin", want)
	if got := repositoryFromDir(dir, "emal_processing"); got != want {
		t.Fatalf("got %q", got)
	}
	if got := repositoryFromDir(dir, "other"); got != "" {
		t.Fatalf("unrelated module: %s", got)
	}
	sub := filepath.Join(dir, "cmd")
	if err := os.Mkdir(sub, 0700); err != nil {
		t.Fatal(err)
	}
	if got := repositoryFromDir(sub, "emal_processing"); got != want {
		t.Fatalf("subdirectory: %s", got)
	}
	if err := os.WriteFile(filepath.Join(sub, "go.mod"), []byte("module nested\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := repositoryFromDir(sub, "emal_processing"); got != "" {
		t.Fatalf("crossed nested module: %s", got)
	}
}

func TestRepositoryDisplay(t *testing.T) {
	for input, want := range map[string]string{
		"https://user:secret@example.com/team/repo.git?token=secret#secret": "https://example.com/team/repo.git",
		"git@example.com:team/repo.git":                                     "git@example.com:team/repo.git",
		"ssh://git@example.com/team/repo.git":                               "ssh://example.com/team/repo.git",
		"bad\nvalue":                                                        "",
	} {
		if got := repositoryDisplay(input); got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}
