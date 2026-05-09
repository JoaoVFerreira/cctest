package cctest

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type prettyTestStatus int

const (
	prettyTestPassed prettyTestStatus = iota
	prettyTestFailed
	prettyTestSkipped
)

const (
	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiDim    = "\x1b[2m"
	ansiGreen  = "\x1b[32m"
	ansiRed    = "\x1b[31m"
	ansiYellow = "\x1b[33m"
)

type prettyReporter struct {
	out       io.Writer
	enabled   bool
	color     bool
	suiteName string
	started   bool
	mu        sync.Mutex
	passed    int
	failed    int
	skipped   int
}

func newPrettyReporter(out io.Writer, enabled bool, color bool, suiteName string) *prettyReporter {
	return &prettyReporter{
		out:       out,
		enabled:   enabled,
		color:     color,
		suiteName: suiteName,
	}
}

func reporterEnabled(config Config) bool {
	if !config.prettyOutput || !testing.Verbose() {
		return false
	}
	return os.Getenv("CCTEST_PRETTY") != "0"
}

func reporterColorEnabled(config Config) bool {
	if config.colorOutput != nil {
		return *config.colorOutput
	}
	if os.Getenv("NO_COLOR") != "" || os.Getenv("CCTEST_COLOR") == "0" {
		return false
	}
	return true
}

func (r *prettyReporter) Start() {
	if !r.enabled {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.started {
		return
	}
	r.started = true

	fmt.Fprintf(r.out, "\n%s %s\n", r.paint(ansiBold, "cctest"), r.paint(ansiDim, r.suiteName))
}

func (r *prettyReporter) TestDone(path string, status prettyTestStatus, duration time.Duration) {
	if !r.enabled {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	label, color := statusLabel(status)
	switch status {
	case prettyTestPassed:
		r.passed++
	case prettyTestFailed:
		r.failed++
	case prettyTestSkipped:
		r.skipped++
	}

	fmt.Fprintf(
		r.out,
		"  %s %s %s\n",
		r.paint(color, label),
		path,
		r.paint(ansiDim, "("+formatPrettyDuration(duration)+")"),
	)
}

func (r *prettyReporter) Summary(duration time.Duration, failed bool) {
	if !r.enabled {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	total := r.passed + r.failed + r.skipped
	suiteStatus := "passed"
	suiteColor := ansiGreen
	if failed || r.failed > 0 {
		suiteStatus = "failed"
		suiteColor = ansiRed
	}

	fmt.Fprintln(r.out)
	fmt.Fprintf(r.out, "  %s 1 %s, 1 total\n", r.paint(ansiBold, "Test Suites:"), r.paint(suiteColor, suiteStatus))
	fmt.Fprintf(r.out, "  %s %s\n", r.paint(ansiBold, "Tests:      "), r.testSummary(total))
	fmt.Fprintf(r.out, "  %s %s\n\n", r.paint(ansiBold, "Time:       "), formatPrettyDuration(duration))
}

func (r *prettyReporter) testSummary(total int) string {
	parts := make([]string, 0, 4)
	if r.failed > 0 {
		parts = append(parts, r.paint(ansiRed, fmt.Sprintf("%d failed", r.failed)))
	}
	if r.passed > 0 {
		parts = append(parts, r.paint(ansiGreen, fmt.Sprintf("%d passed", r.passed)))
	}
	if r.skipped > 0 {
		parts = append(parts, r.paint(ansiYellow, fmt.Sprintf("%d skipped", r.skipped)))
	}
	parts = append(parts, fmt.Sprintf("%d total", total))
	return strings.Join(parts, ", ")
}

func (r *prettyReporter) paint(color string, value string) string {
	if !r.color || color == "" {
		return value
	}
	return color + value + ansiReset
}

func statusLabel(status prettyTestStatus) (string, string) {
	switch status {
	case prettyTestFailed:
		return "FAIL", ansiRed
	case prettyTestSkipped:
		return "SKIP", ansiYellow
	default:
		return "PASS", ansiGreen
	}
}

func formatPrettyDuration(duration time.Duration) string {
	if duration < time.Millisecond {
		return "<1ms"
	}
	if duration < time.Second {
		return fmt.Sprintf("%dms", duration.Milliseconds())
	}
	return duration.Round(time.Millisecond).String()
}

func joinTestPath(parts []string) string {
	return strings.Join(parts, " > ")
}
