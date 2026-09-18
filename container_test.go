package buildcard

import (
	"bytes"
	"strings"
	"testing"
)

func TestContainerInfo(t *testing.T) {
	for _, tc := range []struct {
		name, goos, explicit, marker string
		want                         bool
	}{
		{"host", "linux", "", "", false},
		{"docker", "linux", "", "/.dockerenv", true},
		{"podman", "linux", "", "/run/.containerenv", true},
		{"explicit", "windows", "true", "", true},
		{"disabled", "linux", "false", "/.dockerenv", false},
		{"non Linux", "windows", "", "/.dockerenv", false},
		{"invalid flag", "linux", "invalid", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := map[string]string{"BUILDCARD_CONTAINER": tc.explicit, "BUILDCARD_CONTAINER_HOST": "prod01", "BUILDCARD_CONTAINER_USER": "deployer", "BUILDCARD_CONTAINER_NAME": "service", "BUILDCARD_CONTAINER_IMAGE": "service:v1"}
			info := Info{InContainer: true, ContainerHost: "stale"}
			info.setContainerInfo(tc.goos, func(k string) string { return env[k] }, func(p string) bool { return p == tc.marker })
			if info.InContainer != tc.want {
				t.Fatalf("container=%t", info.InContainer)
			}
			if tc.want {
				if info.ContainerHost != "prod01" || info.ContainerUser != "deployer" || info.ContainerName != "service" || info.ContainerImage != "service:v1" {
					t.Fatalf("missing metadata: %+v", info)
				}
				for _, value := range []string{"prod01", "deployer", "service:v1"} {
					if !strings.Contains(info.String(), value) {
						t.Fatalf("missing %s in table", value)
					}
				}
			} else if info.ContainerHost != "" || info.ContainerUser != "" || info.ContainerName != "" || info.ContainerImage != "" {
				t.Fatalf("nonzero host fields: %+v", info)
			}
			var out bytes.Buffer
			if err := info.WriteTable(&out); err != nil || out.String() != info.String() {
				t.Fatal("inconsistent table output", err)
			}
		})
	}
}

func TestContainerWithoutLauncherMetadata(t *testing.T) {
	var info Info
	info.setContainerInfo("linux", func(string) string { return "" }, func(p string) bool { return p == "/.dockerenv" })
	if !info.InContainer || info.ContainerHost != "" || info.ContainerUser != "" || info.ContainerName != "" || info.ContainerImage != "" {
		t.Fatalf("invented metadata: %+v", info)
	}
}
