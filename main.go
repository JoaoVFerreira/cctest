package cctest

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// Main runs m and filters Go's native verbose subtest lines so cctest's reporter
// can be the primary terminal output for packages that opt in through TestMain.
func Main(m *testing.M) {
	os.Exit(RunMain(m))
}

// RunMain is the return-code form of Main for packages that need custom
// TestMain setup or teardown around cctest's output filtering.
func RunMain(m *testing.M) int {
	if !flag.Parsed() {
		flag.Parse()
	}

	if !testing.Verbose() || os.Getenv("CCTEST_CLEAN") == "0" {
		return m.Run()
	}

	output, code, err := captureStdout(m.Run)
	if err != nil {
		fmt.Fprint(os.Stdout, output)
		fmt.Fprintf(os.Stderr, "cctest: unable to filter go test output: %v\n", err)
		return code
	}

	fmt.Fprint(os.Stdout, filterGoTestVerboseOutput(output))
	return code
}

func captureStdout(run func() int) (string, int, error) {
	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		return "", run(), err
	}

	os.Stdout = writer

	output := make(chan string, 1)
	go func() {
		var buffer bytes.Buffer
		_, _ = io.Copy(&buffer, reader)
		output <- buffer.String()
	}()

	var code int
	func() {
		defer func() {
			os.Stdout = original
			_ = writer.Close()
		}()
		code = run()
	}()

	captured := <-output
	_ = reader.Close()

	return captured, code, nil
}

func filterGoTestVerboseOutput(output string) string {
	scanner := bufio.NewScanner(strings.NewReader(output))
	var lines []string

	for scanner.Scan() {
		line := scanner.Text()
		if isGoVerboseLine(line) {
			continue
		}
		lines = append(lines, line)
	}

	if len(lines) == 0 {
		return ""
	}

	filtered := strings.TrimLeft(strings.Join(lines, "\n"), "\n")
	if strings.HasSuffix(output, "\n") {
		filtered += "\n"
	}
	return filtered
}

func isGoVerboseLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	switch {
	case strings.HasPrefix(trimmed, "=== RUN"):
		return true
	case strings.HasPrefix(trimmed, "=== PAUSE"):
		return true
	case strings.HasPrefix(trimmed, "=== CONT"):
		return true
	case strings.HasPrefix(trimmed, "--- PASS:"):
		return true
	case strings.HasPrefix(trimmed, "--- FAIL:"):
		return true
	case strings.HasPrefix(trimmed, "--- SKIP:"):
		return true
	case trimmed == "PASS":
		return true
	case trimmed == "FAIL":
		return true
	default:
		return false
	}
}
