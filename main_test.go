package cctest

import (
	"strings"
	"testing"
)

func TestFilterGoTestVerboseOutput(t *testing.T) {
	input := strings.Join([]string{
		"=== RUN   TestCadastrarOrganiazacao",
		"",
		"cctest cadastrarOrganizacao",
		"=== RUN   TestCadastrarOrganiazacao/cadastrarOrganizacao",
		"=== RUN   TestCadastrarOrganiazacao/cadastrarOrganizacao/creates",
		"  PASS cadastrarOrganizacao > creates (10ms)",
		"",
		"  Test Suites: 1 passed, 1 total",
		"  Tests:       1 passed, 1 total",
		"  Time:        10ms",
		"",
		"--- PASS: TestCadastrarOrganiazacao (0.01s)",
		"    --- PASS: TestCadastrarOrganiazacao/cadastrarOrganizacao (0.01s)",
		"        --- PASS: TestCadastrarOrganiazacao/cadastrarOrganizacao/creates (0.01s)",
		"PASS",
		"ok  example/test 0.020s",
		"",
	}, "\n")

	got := filterGoTestVerboseOutput(input)

	for _, unwanted := range []string{"=== RUN", "--- PASS:", "\nPASS\n"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("expected filtered output to remove %q\noutput:\n%s", unwanted, got)
		}
	}

	for _, wanted := range []string{
		"cctest cadastrarOrganizacao",
		"PASS cadastrarOrganizacao > creates (10ms)",
		"Test Suites: 1 passed, 1 total",
		"ok  example/test 0.020s",
	} {
		if !strings.Contains(got, wanted) {
			t.Fatalf("expected filtered output to contain %q\noutput:\n%s", wanted, got)
		}
	}
}
