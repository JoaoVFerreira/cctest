package cctest

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestPrettyReporterFormatsJestLikeOutput(t *testing.T) {
	var out bytes.Buffer
	reporter := newPrettyReporter(&out, true, false, "Root")

	reporter.Start()
	reporter.TestDone("Root > creates an asset", prettyTestPassed, 2*time.Millisecond)
	reporter.TestDone("Root > rejects invalid input", prettyTestFailed, 3*time.Millisecond)
	reporter.TestDone("Root > pending case", prettyTestSkipped, time.Microsecond)
	reporter.Summary(10*time.Millisecond, true)

	got := out.String()

	for _, want := range []string{
		"cctest Root",
		"PASS Root > creates an asset (2ms)",
		"FAIL Root > rejects invalid input (3ms)",
		"SKIP Root > pending case (<1ms)",
		"Test Suites: 1 failed, 1 total",
		"Tests:       1 failed, 1 passed, 1 skipped, 3 total",
		"Time:        10ms",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected reporter output to contain %q\noutput:\n%s", want, got)
		}
	}
}

func TestPrettyReporterCanColorStatusLabels(t *testing.T) {
	var out bytes.Buffer
	reporter := newPrettyReporter(&out, true, true, "Root")

	reporter.TestDone("Root > green", prettyTestPassed, time.Millisecond)
	reporter.TestDone("Root > red", prettyTestFailed, time.Millisecond)

	got := out.String()
	if !strings.Contains(got, ansiGreen+"PASS"+ansiReset) {
		t.Fatalf("expected green PASS label in output:\n%s", got)
	}
	if !strings.Contains(got, ansiRed+"FAIL"+ansiReset) {
		t.Fatalf("expected red FAIL label in output:\n%s", got)
	}
}

func TestPrettyReporterDoesNothingWhenDisabled(t *testing.T) {
	var out bytes.Buffer
	reporter := newPrettyReporter(&out, false, true, "Root")

	reporter.Start()
	reporter.TestDone("Root > hidden", prettyTestPassed, time.Millisecond)
	reporter.Summary(time.Millisecond, false)

	if got := out.String(); got != "" {
		t.Fatalf("expected disabled reporter to stay silent, got %q", got)
	}
}
