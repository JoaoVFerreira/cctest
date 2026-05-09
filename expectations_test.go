package cctest

import (
	"errors"
	"testing"
)

func TestExpectationsPassForCommonMatchers(t *testing.T) {
	Describe(t, "Expectations", func(s *Suite) {
		s.It("matches common values", func(ctx *Context) {
			ctx.Expect("abc").ToEqual("abc")
			ctx.Expect(map[string]any{"a": float64(1)}).ToDeepEqual(map[string]any{"a": float64(1)})
			ctx.Expect(nil).ToBeNil()
			ctx.Expect(errors.New("boom")).ToNotBeNil()
			ctx.Expect(true).ToBeTrue()
			ctx.Expect(false).ToBeFalse()
			ctx.Expect("hello world").ToContain("world")
			ctx.Expect([]string{"a", "b"}).ToContain("b")
			ctx.Expect(map[string]int{"a": 1}).ToContain("a")
			ctx.Expect([]int{1, 2, 3}).ToHaveLen(3)
			ctx.Expect(`{"b":2,"a":1}`).ToMatchJSON(map[string]any{"a": 1, "b": 2})
			ctx.Expect(errors.New("wrapped boom")).ToError()
			ctx.Expect(errors.New("wrapped boom")).ToErrorContain("boom")
		})
	})
}

func TestExpectationNegation(t *testing.T) {
	Describe(t, "Expectations", func(s *Suite) {
		s.It("negates matchers", func(ctx *Context) {
			ctx.Expect("abc").Not().ToEqual("def")
			ctx.Expect(errors.New("boom")).Not().ToBeNil()
			ctx.Expect([]string{"a", "b"}).Not().ToContain("c")
			ctx.Expect([]int{1, 2, 3}).Not().ToHaveLen(2)
		})
	})
}
