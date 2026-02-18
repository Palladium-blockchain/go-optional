package optional

import "testing"

func TestZeroValueIsEmpty(t *testing.T) {
	var o Optional[int]
	if !o.IsEmpty() {
		t.Fatalf("zero value Optional should be empty")
	}
	if v, ok := o.Get(); ok || v != 0 {
		t.Fatalf("Get on empty: got (v=%v, ok=%v), want (0, false)", v, ok)
	}
	if p := o.ToPtr(); p != nil {
		t.Fatalf("ToPtr on empty: got %v, want nil", *p)
	}
	if got := o.Or(42); got != 42 {
		t.Fatalf("Or on empty: got %v, want 42", got)
	}
}

func TestNewHasValue(t *testing.T) {
	o := New(10)

	if o.IsEmpty() {
		t.Fatalf("New should create non-empty Optional")
	}

	v, ok := o.Get()
	if !ok || v != 10 {
		t.Fatalf("Get: got (v=%v, ok=%v), want (10, true)", v, ok)
	}

	if got := o.Or(42); got != 10 {
		t.Fatalf("Or: got %v, want 10", got)
	}
}

func TestFromPtrNil(t *testing.T) {
	o := FromPtr[int](nil)

	if !o.IsEmpty() {
		t.Fatalf("FromPtr(nil) should be empty")
	}

	v, ok := o.Get()
	if ok || v != 0 {
		t.Fatalf("Get: got (v=%v, ok=%v), want (0, false)", v, ok)
	}

	if p := o.ToPtr(); p != nil {
		t.Fatalf("ToPtr: got %v, want nil", *p)
	}

	if got := o.Or(7); got != 7 {
		t.Fatalf("Or: got %v, want 7", got)
	}
}

func TestFromPtrNonNil(t *testing.T) {
	x := 5
	o := FromPtr(&x)

	if o.IsEmpty() {
		t.Fatalf("FromPtr(non-nil) should be non-empty")
	}

	v, ok := o.Get()
	if !ok || v != 5 {
		t.Fatalf("Get: got (v=%v, ok=%v), want (5, true)", v, ok)
	}

	p := o.ToPtr()
	if p == nil || *p != 5 {
		t.Fatalf("ToPtr: got %v, want pointer to 5", p)
	}
}

func TestEmptyFunction(t *testing.T) {
	o := Empty[string]()

	if !o.IsEmpty() {
		t.Fatalf("Empty() should return empty Optional")
	}

	v, ok := o.Get()
	if ok || v != "" {
		t.Fatalf("Get: got (v=%q, ok=%v), want (\"\", false)", v, ok)
	}
}

func TestSetAndUnset(t *testing.T) {
	var o Optional[int]

	o.Set(99)
	if o.IsEmpty() {
		t.Fatalf("after Set, Optional should be non-empty")
	}
	v, ok := o.Get()
	if !ok || v != 99 {
		t.Fatalf("Get after Set: got (v=%v, ok=%v), want (99, true)", v, ok)
	}

	o.Unset()
	if !o.IsEmpty() {
		t.Fatalf("after Unset, Optional should be empty")
	}
	v, ok = o.Get()
	if ok || v != 0 {
		t.Fatalf("Get after Unset: got (v=%v, ok=%v), want (0, false)", v, ok)
	}
}

func TestToPtrReturnsCopy(t *testing.T) {
	o := New(1)
	p := o.ToPtr()

	if p == nil {
		t.Fatalf("ToPtr: got nil, want non-nil")
	}
	if *p != 1 {
		t.Fatalf("ToPtr: got %v, want 1", *p)
	}

	// Changing the returned pointer must not affect the Optional (copy semantics).
	*p = 2

	v, ok := o.Get()
	if !ok || v != 1 {
		t.Fatalf("Optional value changed via ToPtr result: got (v=%v, ok=%v), want (1, true)", v, ok)
	}
}

func TestOptionalWithStructType(t *testing.T) {
	type S struct {
		A int
		B string
	}

	s := S{A: 7, B: "x"}
	o := New(s)

	v, ok := o.Get()
	if !ok || v != s {
		t.Fatalf("Get: got (v=%v, ok=%v), want (%v, true)", v, ok, s)
	}

	p := o.ToPtr()
	if p == nil || *p != s {
		t.Fatalf("ToPtr: got %v, want pointer to %v", p, s)
	}

	// Ensure it's a copy.
	p.A = 8
	v2, ok := o.Get()
	if !ok || v2.A != 7 {
		t.Fatalf("Optional struct changed via ToPtr result: got A=%v, want 7", v2.A)
	}
}
func TestToPtr_ShallowCopyReferenceTypes_TableDriven(t *testing.T) {
	type testCase[T any] struct {
		name   string
		makeV  func() T
		mutate func(p *T)
		// If true: after mutate(*ptr), value inside Optional is expected to change too (shared backing data)
		expectOptionalChanged bool
		check                 func(t *testing.T, got T)
	}

	t.Run("slice", func(t *testing.T) {
		cases := []testCase[[]int]{
			{
				name: "append may or may not reallocate; we mutate element to guarantee shared backing",
				makeV: func() []int {
					return []int{1, 2, 3}
				},
				mutate: func(p *[]int) {
					// modify underlying array element
					(*p)[0] = 99
				},
				expectOptionalChanged: true,
				check: func(t *testing.T, got []int) {
					if len(got) != 3 || got[0] != 99 {
						t.Fatalf("got %v, want first element 99", got)
					}
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				v := tc.makeV()
				o := New(v)

				p := o.ToPtr()
				if p == nil {
					t.Fatalf("ToPtr returned nil")
				}

				tc.mutate(p)

				got, ok := o.Get()
				if !ok {
					t.Fatalf("expected ok=true")
				}

				if !tc.expectOptionalChanged {
					// keep for symmetry; currently not used in slice case
					orig, _ := New(v).Get()
					if len(got) != len(orig) {
						t.Fatalf("unexpected change: got %v, want %v", got, orig)
					}
				}

				tc.check(t, got)
			})
		}
	})

	t.Run("map", func(t *testing.T) {
		cases := []testCase[map[string]int]{
			{
				name: "mutating map via *ptr affects optional value (shared map header)",
				makeV: func() map[string]int {
					return map[string]int{"a": 1}
				},
				mutate: func(p *map[string]int) {
					(*p)["b"] = 2
				},
				expectOptionalChanged: true,
				check: func(t *testing.T, got map[string]int) {
					if got["a"] != 1 || got["b"] != 2 {
						t.Fatalf("got %v, want keys a=1 and b=2", got)
					}
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				v := tc.makeV()
				o := New(v)

				p := o.ToPtr()
				if p == nil {
					t.Fatalf("ToPtr returned nil")
				}

				tc.mutate(p)

				got, ok := o.Get()
				if !ok {
					t.Fatalf("expected ok=true")
				}
				tc.check(t, got)
			})
		}
	})

	t.Run("pointer", func(t *testing.T) {
		cases := []testCase[*int]{
			{
				name: "mutating pointed value via *ptr affects optional value (shared pointee)",
				makeV: func() *int {
					x := 10
					return &x
				},
				mutate: func(p **int) {
					**p = 77
				},
				expectOptionalChanged: true,
				check: func(t *testing.T, got *int) {
					if got == nil || *got != 77 {
						t.Fatalf("got %v, want pointer to 77", got)
					}
				},
			},
			{
				name: "replacing pointer itself via *ptr does NOT affect optional value (pointer copied)",
				makeV: func() *int {
					x := 10
					return &x
				},
				mutate: func(p **int) {
					y := 5
					*p = &y // replace pointer variable itself
				},
				expectOptionalChanged: false,
				check: func(t *testing.T, got *int) {
					// original pointee should remain 10
					if got == nil || *got != 10 {
						t.Fatalf("got %v, want pointer to 10", got)
					}
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				v := tc.makeV()
				o := New(v)

				p := o.ToPtr()
				if p == nil {
					t.Fatalf("ToPtr returned nil")
				}

				tc.mutate(p)

				got, ok := o.Get()
				if !ok {
					t.Fatalf("expected ok=true")
				}

				tc.check(t, got)
			})
		}
	})

	t.Run("chan", func(t *testing.T) {
		cases := []testCase[chan int]{
			{
				name: "sending via *ptr affects same channel (shared channel value)",
				makeV: func() chan int {
					return make(chan int, 1)
				},
				mutate: func(p *chan int) {
					*p <- 123
				},
				expectOptionalChanged: true,
				check: func(t *testing.T, got chan int) {
					select {
					case x := <-got:
						if x != 123 {
							t.Fatalf("got %v, want 123", x)
						}
					default:
						t.Fatalf("expected value in channel")
					}
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				v := tc.makeV()
				o := New(v)

				p := o.ToPtr()
				if p == nil {
					t.Fatalf("ToPtr returned nil")
				}

				tc.mutate(p)

				got, ok := o.Get()
				if !ok {
					t.Fatalf("expected ok=true")
				}
				tc.check(t, got)
			})
		}
	})

	t.Run("func", func(t *testing.T) {
		calls := 0
		f := func() { calls++ }

		cases := []testCase[func()]{
			{
				name:  "function value copy calls same underlying function",
				makeV: func() func() { return f },
				mutate: func(p *func()) {
					(*p)()
				},
				expectOptionalChanged: true, // "changed" here means: calling via ptr affects external state
				check: func(t *testing.T, got func()) {
					got()
					if calls != 2 {
						t.Fatalf("calls=%d, want 2", calls)
					}
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				o := New(tc.makeV())
				p := o.ToPtr()
				if p == nil {
					t.Fatalf("ToPtr returned nil")
				}

				tc.mutate(p)

				got, ok := o.Get()
				if !ok {
					t.Fatalf("expected ok=true")
				}
				tc.check(t, got)
			})
		}
	})
}

func TestFromPtrReferenceTypes_TableDriven(t *testing.T) {
	t.Run("nil pointers produce empty", func(t *testing.T) {
		if o := FromPtr[[]int](nil); !o.IsEmpty() {
			t.Fatalf("FromPtr(nil slice ptr) should be empty")
		}
		if o := FromPtr[map[string]int](nil); !o.IsEmpty() {
			t.Fatalf("FromPtr(nil map ptr) should be empty")
		}
		if o := FromPtr[*int](nil); !o.IsEmpty() {
			t.Fatalf("FromPtr(nil *int ptr) should be empty")
		}
	})

	t.Run("non-nil pointers copy the value", func(t *testing.T) {
		s := []int{1, 2}
		m := map[string]int{"a": 1}
		x := 5
		p := &x

		cases := []struct {
			name string
			run  func(t *testing.T)
		}{
			{
				name: "slice",
				run: func(t *testing.T) {
					o := FromPtr(&s)
					got, ok := o.Get()
					if !ok || len(got) != 2 || got[0] != 1 {
						t.Fatalf("got %v (ok=%v), want [1 2] (true)", got, ok)
					}
				},
			},
			{
				name: "map",
				run: func(t *testing.T) {
					o := FromPtr(&m)
					got, ok := o.Get()
					if !ok || got["a"] != 1 {
						t.Fatalf("got %v (ok=%v), want map[a:1] (true)", got, ok)
					}
				},
			},
			{
				name: "pointer-to-int as T",
				run: func(t *testing.T) {
					o := FromPtr(&p) // *T where T is *int
					got, ok := o.Get()
					if !ok || got == nil || *got != 5 {
						t.Fatalf("got %v (ok=%v), want pointer to 5 (true)", got, ok)
					}
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, tc.run)
		}
	})
}
