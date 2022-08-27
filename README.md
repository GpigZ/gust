# gust [![Docs](https://img.shields.io/badge/Docs-pkg.go.dev-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/andeya/gust)

Golang ergonomic declarative generics module inspired by Rust.

## Go Version

go≥1.18

## Features

### Result

Improve `func() (T,error)`, handle result with chain methods.

- Result Example

```go
func ExampleResult_AndThen() {
	var divide = func(i, j float32) gust.Result[float32] {
		if j == 0 {
			return gust.Err[float32]("j can not be 0")
		}
		return gust.Ok(i / j)
	}
	var ret float32 = divide(1, 2).AndThen(func(i float32) gust.Result[float32] {
		return gust.Ok(i * 10)
	}).Unwrap()
	fmt.Println(ret)
	// Output:
	// 5
}
```

```go
func ExampleResult_UnwrapOr() {
	const def int = 10

	// before
	i, err := strconv.Atoi("1")
	if err != nil {
		i = def
	}
	fmt.Println(i * 2)

	// now
	fmt.Println(gust.Ret(strconv.Atoi("1")).UnwrapOr(def) * 2)

	// Output:
	// 2
	// 2
}
```

### Option

Improve `func()(T, bool)` and `if *U != nil`, handle value with `Option` type.

- Option Example

```go
func ExampleOption() {
	type A struct {
		X int
	}
	var a = gust.Some(A{X: 1})
	fmt.Println(a.IsSome(), a.IsNone())

	var b = gust.None[A]()
	fmt.Println(b.IsSome(), b.IsNone())

	var x = b.UnwrapOr(A{X: 2})
	fmt.Println(x)

	var c *A
	fmt.Println(gust.Ptr(c).IsNone())
	c = new(A)
	fmt.Println(gust.Ptr(c).IsNone())

	type B struct {
		Y string
	}
	var d = opt.Map(a, func(t A) B {
		return B{
			Y: strconv.Itoa(t.X),
		}
	})
	fmt.Println(d)

	// Output:
	// true false
	// false true
	// {2}
	// true
	// false
	// Some({1})
}
```

### Errable

Improve `func() error`, handle error with chain methods.

- Result Example

```go
func ExampleErrable() {
	var hasErr = true
	var f = func() gust.Errable[int] {
		if hasErr {
			return gust.ToErrable(1)
		}
		return gust.NonErrable[int]()
	}
	var r = f()
	fmt.Println(r.HasError())
	fmt.Println(r.Unwrap())
	fmt.Printf("%#v", r.ToError())
	// Output:
	// true
	// 1
	// &errors.errorString{s:"1"}
}
```

### Iterator

Feature-rich iterators.

- Iterator Example

```go
func TestAny(t *testing.T) {
	var iter = FromVec([]int{1, 2, 3})
	if !iter.Any(func(x int) bool {
		return x > 1
	}) {
		t.Error("Any failed")
	}
}
```