package buildcard

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"
)

var startedAt = time.Now()

// Info is a point-in-time snapshot of build and process information.
// StartedAt measures package initialization time, not operating system process start.
type Info struct {
	Version        string
	CommitHash     string
	ShortHash      string
	CommitDate     string
	Dirty          bool
	BuildTime      string
	Repository     string
	Hostname       string
	Username       string
	ExecutablePath string
	ExecutableName string
	PID            int
	ParentPID      int
	GOOS           string
	GOARCH         string
	StartedAt      string
	Uptime         string
	Args           []string // Command-line arguments excluding the executable name.
}

// Snapshot collects build metadata and runtime information for the current process.
func Snapshot() Info {
	execPath := executablePath()
	info := readBuildInfo()
	args := os.Args
	if len(args) > 0 {
		args = args[1:]
	}

	info.Hostname = normalizeText(hostname())
	info.Username = normalizeText(username())
	info.ExecutablePath = execPath
	info.ExecutableName = normalizeText(filepath.Base(execPath))
	info.PID = os.Getpid()
	info.ParentPID = os.Getppid()
	info.GOOS = runtime.GOOS
	info.GOARCH = runtime.GOARCH
	info.StartedAt = startedAt.Local().Format(time.RFC3339)
	info.Uptime = formatUptime(startedAt)
	info.Args = slices.Clone(args)
	return info
}

// String returns the snapshot formatted as a table.
func (s Info) String() string { return formatTable(s.tableRows()) }

// Preview displays the captured snapshot as a table on standard output.
// It does not refresh the snapshot's data.
func (s Info) Preview() { fmt.Print(s.String()) }

// PrintTable prints the snapshot to standard output.
// Deprecated: use Preview instead.
func (s Info) PrintTable() { s.Preview() }

// WriteTable writes the snapshot to w.
func (s Info) WriteTable(w io.Writer) error {
	_, err := io.WriteString(w, s.String())
	return err
}

func (s Info) tableRows() [][2]string {
	return [][2]string{
		{"Версия", normalizeText(s.Version)},
		{"Коммит", normalizeText(s.ShortHash)},
		{"Дата коммита", normalizeText(s.CommitDate)},
		{"Грязный коммит", strconv.FormatBool(s.Dirty)},
		{"Дата сборки", normalizeText(s.BuildTime)},
		{"Репозиторий", normalizeText(s.Repository)},
		{"Машина", normalizeText(s.Hostname)},
		{"Пользователь", normalizeText(s.Username)},
		{"Исполняемый путь", normalizeText(s.ExecutablePath)},
		{"Исполняемый файл", normalizeText(s.ExecutableName)},
		{"Процесс PID", strconv.Itoa(s.PID)},
		{"Родитель PID", strconv.Itoa(s.ParentPID)},
		{"Операционная система", normalizeText(s.GOOS)},
		{"Архитектура", normalizeText(s.GOARCH)},
		{"Запущено", normalizeText(s.StartedAt)},
		{"В работе", normalizeText(s.Uptime)},
		{"Аргументы запуска", formatArgs(s.Args)},
	}
}

func hostname() string {
	name, err := os.Hostname()
	if err != nil {
		return ""
	}

	return name
}

func username() string {
	u, err := user.Current()
	if err == nil && u != nil {
		if strings.TrimSpace(u.Username) != "" {
			return u.Username
		}

		if strings.TrimSpace(u.Name) != "" {
			return u.Name
		}
	}

	if name := os.Getenv("USER"); strings.TrimSpace(name) != "" {
		return name
	}
	return os.Getenv("USERNAME")
}

func executablePath() string {
	path, err := os.Executable()
	if err != nil {
		return unknownValue
	}

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return normalizeText(path)
	}

	return normalizeText(resolved)
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return unknownValue
	}

	return t.Local().Format(time.RFC3339)
}

func formatUptime(startedAt time.Time) string {
	if startedAt.IsZero() {
		return unknownValue
	}

	uptime := time.Since(startedAt).Round(time.Second)

	if uptime < 0 {
		uptime = 0
	}

	return durationString(uptime)
}

func formatArgs(args []string) string {
	if len(args) == 0 {
		return unknownValue
	}

	return strings.Join(args, ", ")
}
