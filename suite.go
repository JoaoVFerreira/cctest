package cctest

import (
	"os"
	"testing"
	"time"
)

type hookFunc func(*Context)

type suiteNode struct {
	suite *Suite
	test  *testCase
}

type testCase struct {
	name string
	fn   func(*Context)
	skip bool
}

// Suite records nested suites, test cases, and hooks before running them as
// ordinary Go subtests.
type Suite struct {
	name       string
	config     Config
	nodes      []suiteNode
	beforeEach []hookFunc
	afterEach  []hookFunc
}

// Describe creates a root suite and runs it with testing.T subtests.
func Describe(t *testing.T, name string, fn func(*Suite), opts ...Option) {
	t.Helper()

	suite := newSuite(name)
	suite.config = defaultSuiteConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&suite.config)
		}
	}

	if fn != nil {
		fn(suite)
	}

	reporter := newPrettyReporter(os.Stdout, reporterEnabled(suite.config), reporterColorEnabled(suite.config), suite.name)
	started := time.Now()
	reporter.Start()
	suite.run(t, nil, reporter)
	reporter.Summary(time.Since(started), t.Failed())
}

func newSuite(name string) *Suite {
	return &Suite{
		name: name,
	}
}

// Describe records a nested suite.
func (s *Suite) Describe(name string, fn func(*Suite)) {
	child := newSuite(name)
	s.nodes = append(s.nodes, suiteNode{suite: child})

	if fn != nil {
		fn(child)
	}
}

// It records a test case.
func (s *Suite) It(name string, fn func(*Context)) {
	s.nodes = append(s.nodes, suiteNode{test: &testCase{
		name: name,
		fn:   fn,
	}})
}

// Skip records a skipped test case.
func (s *Suite) Skip(name string, fn func(*Context)) {
	s.nodes = append(s.nodes, suiteNode{test: &testCase{
		name: name,
		fn:   fn,
		skip: true,
	}})
}

// BeforeEach records a hook that runs before every It in this suite and its
// nested suites.
func (s *Suite) BeforeEach(fn func(*Context)) {
	if fn != nil {
		s.beforeEach = append(s.beforeEach, fn)
	}
}

// AfterEach records a hook that runs after every It in this suite and its
// nested suites.
func (s *Suite) AfterEach(fn func(*Context)) {
	if fn != nil {
		s.afterEach = append(s.afterEach, fn)
	}
}

func (s *Suite) run(t *testing.T, ancestors []*Suite, reporter *prettyReporter) {
	t.Helper()

	suite := s
	t.Run(suite.name, func(t *testing.T) {
		chain := appendSuite(ancestors, suite)

		for _, node := range suite.nodes {
			node := node
			if node.suite != nil {
				node.suite.run(t, chain, reporter)
				continue
			}

			if node.test != nil {
				runTest(t, node.test, chain, reporter)
			}
		}
	})
}

func runTest(t *testing.T, tc *testCase, suites []*Suite, reporter *prettyReporter) {
	t.Helper()

	test := tc
	t.Run(test.name, func(t *testing.T) {
		started := time.Now()
		defer func() {
			reporter.TestDone(testPath(suites, test), testStatusFromT(t), time.Since(started))
		}()

		if test.skip {
			t.Skip("skipped")
		}

		ctx := newContext(t, rootConfig(suites))
		defer runAfterEach(ctx, suites)

		runBeforeEach(ctx, suites)

		if test.fn != nil {
			test.fn(ctx)
		}
	})
}

func testStatusFromT(t *testing.T) prettyTestStatus {
	switch {
	case t.Skipped():
		return prettyTestSkipped
	case t.Failed():
		return prettyTestFailed
	default:
		return prettyTestPassed
	}
}

func testPath(suites []*Suite, test *testCase) string {
	parts := make([]string, 0, len(suites)+1)
	for _, suite := range suites {
		parts = append(parts, suite.name)
	}
	parts = append(parts, test.name)
	return joinTestPath(parts)
}

func rootConfig(suites []*Suite) Config {
	if len(suites) == 0 {
		return defaultSuiteConfig()
	}
	return suites[0].config
}

func runBeforeEach(ctx *Context, suites []*Suite) {
	ctx.t.Helper()

	for _, suite := range suites {
		for _, hook := range suite.beforeEach {
			hook(ctx)
		}
	}
}

func runAfterEach(ctx *Context, suites []*Suite) {
	ctx.t.Helper()

	hooks := collectAfterEach(suites)
	for i := len(hooks) - 1; i >= 0; i-- {
		hook := hooks[i]
		defer hook(ctx)
	}
}

func collectAfterEach(suites []*Suite) []hookFunc {
	var hooks []hookFunc
	for i := len(suites) - 1; i >= 0; i-- {
		hooks = append(hooks, suites[i].afterEach...)
	}
	return hooks
}

func appendSuite(suites []*Suite, suite *Suite) []*Suite {
	next := make([]*Suite, 0, len(suites)+1)
	next = append(next, suites...)
	next = append(next, suite)
	return next
}
