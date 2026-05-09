package cctest

import (
	"reflect"
	"strings"
	"testing"
)

func TestDescribeRunsSuitesAndHooksInOrder(t *testing.T) {
	var got []string

	Describe(t, "Root", func(s *Suite) {
		s.BeforeEach(func(ctx *Context) {
			got = append(got, "root before")
		})
		s.AfterEach(func(ctx *Context) {
			got = append(got, "root after")
		})

		s.It("first", func(ctx *Context) {
			got = append(got, "first")
		})

		s.Describe("Child", func(s *Suite) {
			s.BeforeEach(func(ctx *Context) {
				got = append(got, "child before")
			})
			s.AfterEach(func(ctx *Context) {
				got = append(got, "child after")
			})

			s.It("second", func(ctx *Context) {
				got = append(got, "second")
			})
		})
	})

	want := []string{
		"root before",
		"first",
		"root after",
		"root before",
		"child before",
		"second",
		"child after",
		"root after",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("hook order mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestAfterEachRunsWhenTestSkips(t *testing.T) {
	var got []string

	Describe(t, "Root", func(s *Suite) {
		s.BeforeEach(func(ctx *Context) {
			got = append(got, "before")
		})
		s.AfterEach(func(ctx *Context) {
			got = append(got, "after")
		})

		s.It("skips", func(ctx *Context) {
			got = append(got, "test")
			ctx.T().SkipNow()
		})
	})

	want := []string{"before", "test", "after"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("skip cleanup mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestOuterAfterEachRunsWhenInnerAfterEachSkips(t *testing.T) {
	var got []string

	Describe(t, "Root", func(s *Suite) {
		s.AfterEach(func(ctx *Context) {
			got = append(got, "root after")
		})

		s.Describe("Child", func(s *Suite) {
			s.AfterEach(func(ctx *Context) {
				got = append(got, "child after")
				ctx.T().SkipNow()
			})

			s.It("test", func(ctx *Context) {
				got = append(got, "test")
			})
		})
	})

	want := []string{"test", "child after", "root after"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("after hook cleanup mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestSkipDoesNotRunHooksOrBody(t *testing.T) {
	var got []string

	Describe(t, "Root", func(s *Suite) {
		s.BeforeEach(func(ctx *Context) {
			got = append(got, "before")
		})
		s.AfterEach(func(ctx *Context) {
			got = append(got, "after")
		})
		s.Skip("skipped", func(ctx *Context) {
			got = append(got, "test")
		})
	})

	if len(got) != 0 {
		t.Fatalf("skipped test ran unexpectedly: %#v", got)
	}
}

func TestEachItReceivesIsolatedContext(t *testing.T) {
	var contexts []*Context
	var names []string

	Describe(t, "Root", func(s *Suite) {
		s.It("first", func(ctx *Context) {
			contexts = append(contexts, ctx)
			names = append(names, ctx.Name())
		})

		s.It("second", func(ctx *Context) {
			contexts = append(contexts, ctx)
			names = append(names, ctx.Name())
		})
	})

	if len(contexts) != 2 {
		t.Fatalf("expected two contexts, got %d", len(contexts))
	}
	if contexts[0] == contexts[1] {
		t.Fatal("expected each It to receive a different context")
	}
	if !strings.HasSuffix(names[0], "/Root/first") {
		t.Fatalf("unexpected first test name: %q", names[0])
	}
	if !strings.HasSuffix(names[1], "/Root/second") {
		t.Fatalf("unexpected second test name: %q", names[1])
	}
}


Commite as alterações com EXECEÇÃO dos arquivos
- cc_integration_v1.md 
- cc_tech_plan.md 
- cc_unit_v1.md 
- proposal_solution.md 