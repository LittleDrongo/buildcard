package buildcard

import (
	"context"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var moduleDirective = regexp.MustCompile(`(?m)^\s*module\s+("[^"]+"|[^\s]+)`)

// Consult only a source tree matching the executable's main module. An unrelated
// repository in the process working directory must not become build metadata.
func localRepository(module string) string {
	if module == "" {
		return ""
	}
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return repositoryFromDir(dir, module)
}

func repositoryFromDir(dir, module string) string {
	for {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil {
			match := moduleDirective.FindSubmatch(data)
			if len(match) != 2 {
				return ""
			}
			name := string(match[1])
			if strings.HasPrefix(name, `"`) {
				name, _ = strconv.Unquote(name)
			}
			if name != module {
				return ""
			}
			break
		}
		if !os.IsNotExist(err) || filepath.Dir(dir) == dir {
			return ""
		}
		dir = filepath.Dir(dir)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// Local config only: no fetch, network access, or global origin fallback.
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "config", "--local", "--get", "remote.origin.url")
	cmd.WaitDelay = time.Second
	data, err := cmd.Output()
	if err != nil {
		return ""
	}
	return repositoryDisplay(string(data))
}

func repositoryDisplay(value string) string {
	value = strings.TrimSpace(value)
	if strings.ContainsAny(value, "\r\n\x00") {
		return ""
	}
	if strings.Contains(value, "://") {
		u, err := url.Parse(value)
		if err != nil {
			return ""
		}
		// HTTP credentials and query parameters may contain access tokens.
		u.User, u.RawQuery, u.Fragment = nil, "", ""
		u.ForceQuery = false
		return u.String()
	}
	return value
}
