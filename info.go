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
	InContainer    bool
	ContainerHost  string // Host of the container engine, supplied by the launcher.
	ContainerUser  string // User who launched deployment, not the process identity.
	ContainerName  string
	ContainerImage string
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
	info.setContainerInfo(runtime.GOOS, os.Getenv, func(path string) bool {
		_, err := os.Stat(path)
		return err == nil
	})
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
	rows := [][2]string{
		{"Версия", normalizeText(s.Version)},
		{"Коммит", normalizeText(s.ShortHash)},
		{"Дата коммита", normalizeText(s.CommitDate)},
		{"Грязный коммит", strconv.FormatBool(s.Dirty)},
		{"Дата сборки", normalizeText(s.BuildTime)},
		{"Репозиторий", normalizeText(s.Repository)},
		{},
		{"Запуск:", ""},
		{"Машина", normalizeText(s.Hostname)},
		{"Операционная система", normalizeText(s.GOOS)},
		{"Пользователь", normalizeText(s.Username)},
		{"Исполняемый путь", normalizeText(s.ExecutablePath)},
		{"Исполняемый файл", normalizeText(s.ExecutableName)},
		{"Архитектура", normalizeText(s.GOARCH)},
		{"В работе", normalizeText(s.Uptime)},
		{"Запущено", normalizeText(s.StartedAt)},
		{"Процесс PID", strconv.Itoa(s.PID)},
		{"Родитель PID", strconv.Itoa(s.ParentPID)},
		{"Аргументы запуска", formatArgs(s.Args)},
	}
	if s.InContainer {
		rows = append(rows,
			[2]string{},
			[2]string{"Контейнеризация:", ""},
			[2]string{"Приложение в контейнере", strconv.FormatBool(s.InContainer)},
			[2]string{"Хост контейнера", s.ContainerHost},
			[2]string{"Пользователь запуска контейнера", s.ContainerUser},
			[2]string{"Имя контейнера", s.ContainerName},
			[2]string{"Образ контейнера", s.ContainerImage},
		)
	}
	return rows
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
	return usernameFrom(u, err, os.Getenv, os.Geteuid())
}

func usernameFrom(u *user.User, err error, getenv func(string) string, uid int) string {
	if err == nil && u != nil {
		if strings.TrimSpace(u.Username) != "" {
			return u.Username
		}

		if strings.TrimSpace(u.Name) != "" {
			return u.Name
		}
	}

	for _, key := range []string{"USER", "LOGNAME", "USERNAME"} {
		if name := strings.TrimSpace(getenv(key)); name != "" {
			return name
		}
	}
	if uid >= 0 {
		return "uid=" + strconv.Itoa(uid)
	}
	return ""
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
