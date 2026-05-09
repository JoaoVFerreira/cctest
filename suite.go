package cctest

import "testing"

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

	suite.run(t, nil)
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

func (s *Suite) run(t *testing.T, ancestors []*Suite) {
	t.Helper()

	suite := s
	t.Run(suite.name, func(t *testing.T) {
		chain := appendSuite(ancestors, suite)

		for _, node := range suite.nodes {
			node := node
			if node.suite != nil {
				node.suite.run(t, chain)
				continue
			}

			if node.test != nil {
				runTest(t, node.test, chain)
			}
		}
	})
}

func runTest(t *testing.T, tc *testCase, suites []*Suite) {
	t.Helper()

	test := tc
	t.Run(test.name, func(t *testing.T) {
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
