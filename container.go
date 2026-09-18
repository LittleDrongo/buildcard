package buildcard

import "strings"

// Detection is best effort. Launchers should explicitly set BUILDCARD_CONTAINER
// for engines which do not expose the common Linux marker files.
func (s *Info) setContainerInfo(goos string, getenv func(string) string, exists func(string) bool) {
	s.InContainer = false
	s.ContainerHost, s.ContainerUser, s.ContainerName, s.ContainerImage = "", "", "", ""
	switch strings.ToLower(strings.TrimSpace(getenv("BUILDCARD_CONTAINER"))) {
	case "true", "1":
		s.InContainer = true
	case "false", "0":
		return
	default:
		s.InContainer = goos == "linux" && (exists("/.dockerenv") || exists("/run/.containerenv"))
	}
	if !s.InContainer {
		return
	}
	s.ContainerHost = strings.TrimSpace(getenv("BUILDCARD_CONTAINER_HOST"))
	s.ContainerUser = strings.TrimSpace(getenv("BUILDCARD_CONTAINER_USER"))
	s.ContainerName = strings.TrimSpace(getenv("BUILDCARD_CONTAINER_NAME"))
	s.ContainerImage = strings.TrimSpace(getenv("BUILDCARD_CONTAINER_IMAGE"))
}
