package buildcard

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

func durationString(d time.Duration) string {
	if d < 0 {
		d = -d
	}

	const (
		day  = 24 * time.Hour
		year = 365 * day
	)

	years := d / year
	d %= year

	days := d / day
	d %= day

	hours := d / time.Hour
	d %= time.Hour

	minutes := d / time.Minute
	d %= time.Minute

	seconds := d / time.Second

	var parts []string

	if years > 0 {
		parts = append(parts, fmt.Sprintf("%d г.", years))
	}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d дн.", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d ч.", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%d мин.", minutes))
	}
	if seconds > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%d сек.", seconds))
	}

	return strings.Join(parts, " ")
}

// formatTable aligns labels and values in the snapshot table.
// Empty rows separate sections; labels ending in a colon are section headings.
func formatTable(rows [][2]string) string {
	width := 0
	for _, row := range rows {
		width = max(width, utf8.RuneCountInString(row[0]))
	}
	var out strings.Builder
	for _, row := range rows {
		if row == [2]string{} || strings.HasSuffix(row[0], ":") {
			out.WriteString(row[0])
			out.WriteByte('\n')
			continue
		}
		fmt.Fprintf(&out, "%-*s     %s\n", width, row[0], row[1])
	}
	return out.String()
}
