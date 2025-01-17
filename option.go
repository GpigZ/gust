package gust

import (
	"encoding/json"
	"fmt"
)

// Ptr wraps a pointer value.
// NOTE:
//
//	non-nil pointer is wrapped as Some,
//	and nil pointer is wrapped as None.
func Ptr[U any, T *U](ptr T) Option[T] {
	if ptr == nil {
		return Option[T]{value: nil}
	}
	v := &ptr
	return Option[T]{value: &v}
}

// Opt wraps a value as an Option.
// NOTE:
//
//	non-zero T is wrapped as Some,
//	and zero T is wrapped as None.
func Opt[T comparable](val T) Option[T] {
	var zero T
	if zero == val {
		return Option[T]{value: nil}
	}
	v := &val
	return Option[T]{value: &v}
}

// Some wraps a non-none value.
// NOTE:
//
//	Option[T].IsSome() returns true.
//	and Option[T].IsNone() returns false.
func Some[T any](value T) Option[T] {
	v := &value
	return Option[T]{value: &v}
}

// None returns a none.
// NOTE:
//
//	Option[T].IsNone() returns true,
//	and Option[T].IsSome() returns false.
func None[T any]() Option[T] {
	return Option[T]{value: nil}
}

// Option can be used to avoid `(T, bool)` and `if *U != nil`,
// represents an optional value:
//
//	every [`Option`] is either [`Some`](which is non-none T), or [`None`](which is none).
type Option[T any] struct {
	value **T
}

// String returns the string representation.
func (o Option[T]) String() string {
	if o.IsNone() {
		return "None"
	}
	return fmt.Sprintf("Some(%v)", o.unwrapUnchecked())
}

// ToX converts to `Option[any]`.
func (o Option[T]) ToX() Option[any] {
	if o.IsNone() {
		return None[any]()
	}
	return Some[any](o.unwrapUnchecked())
}

// IsSome returns `true` if the option has value.
func (o Option[T]) IsSome() bool {
	return !o.IsNone()
}

// IsSomeAnd returns `true` if the option has value and the value inside it matches a predicate.
func (o Option[T]) IsSomeAnd(f func(T) bool) bool {
	if o.IsSome() {
		return f(o.unwrapUnchecked())
	}
	return false
}

// IsNone returns `true` if the option is none.
func (o Option[T]) IsNone() bool {
	return o.value == nil || *o.value == nil
}

// Expect returns the contained [`Some`] value.
// Panics if the value is none with a custom panic message provided by `msg`.
func (o Option[T]) Expect(msg string) T {
	if o.IsNone() {
		panic(msg)
	}
	return o.unwrapUnchecked()
}

// Unwrap returns the contained value.
// Panics if the value is none.
func (o Option[T]) Unwrap() T {
	if o.IsSome() {
		return o.unwrapUnchecked()
	}
	var t T
	panic(fmt.Sprintf("call Option[%T].Unwrap() on none", t))
}

// UnwrapOr returns the contained value or a provided default.
func (o Option[T]) UnwrapOr(defaultSome T) T {
	if o.IsSome() {
		return o.unwrapUnchecked()
	}
	return defaultSome
}

// UnwrapOrElse returns the contained value or computes it from a closure.
func (o Option[T]) UnwrapOrElse(defaultSome func() T) T {
	if o.IsSome() {
		return o.unwrapUnchecked()
	}
	return defaultSome()
}

// unwrapUnchecked returns the contained value.
func (o Option[T]) unwrapUnchecked() T {
	return **o.value
}

// Map maps an `Option[T]` to `Option[T]` by applying a function to a contained value.
func (o Option[T]) Map(f func(T) T) Option[T] {
	if o.IsSome() {
		return Some[T](f(o.unwrapUnchecked()))
	}
	return None[T]()
}

// XMap maps an `Option[T]` to `Option[any]` by applying a function to a contained value.
func (o Option[T]) XMap(f func(T) any) Option[any] {
	if o.IsSome() {
		return Some[any](f(o.unwrapUnchecked()))
	}
	return None[any]()
}

// Inspect calls the provided closure with a reference to the contained value (if it has value).
func (o Option[T]) Inspect(f func(T)) Option[T] {
	if o.IsSome() {
		f(o.unwrapUnchecked())
	}
	return o
}

// InspectNone calls the provided closure (if it is none).
func (o Option[T]) InspectNone(f func()) Option[T] {
	if o.IsNone() {
		f()
	}
	return o
}

// MapOr returns the provided default value (if none),
// or applies a function to the contained value (if any).
func (o Option[T]) MapOr(defaultSome T, f func(T) T) T {
	if o.IsSome() {
		return f(o.unwrapUnchecked())
	}
	return defaultSome
}

// XMapOr returns the provided default value (if none),
// or applies a function to the contained value (if any).
func (o Option[T]) XMapOr(defaultSome any, f func(T) any) any {
	if o.IsSome() {
		return f(o.unwrapUnchecked())
	}
	return defaultSome
}

// MapOrElse computes a default function value (if none), or
// applies a different function to the contained value (if any).
func (o Option[T]) MapOrElse(defaultFn func() T, f func(T) T) T {
	if o.IsSome() {
		return f(o.unwrapUnchecked())
	}
	return defaultFn()
}

// XMapOrElse computes a default function value (if none), or
// applies a different function to the contained value (if any).
func (o Option[T]) XMapOrElse(defaultFn func() any, f func(T) any) any {
	if o.IsSome() {
		return f(o.unwrapUnchecked())
	}
	return defaultFn()
}

// OkOr transforms the `Option[T]` into a [`Result[T]`], mapping [`Some(v)`] to
// [`Ok(v)`] and [`None`] to [`Err(err)`].
func (o Option[T]) OkOr(err any) Result[T] {
	if o.IsSome() {
		return Ok(o.Unwrap())
	}
	return Err[T](err)
}

// OkOrElse transforms the `Option[T]` into a [`Result[T]`], mapping [`Some(v)`] to
// [`Ok(v)`] and [`None`] to [`Err(errFn())`].
func (o Option[T]) OkOrElse(errFn func() any) Result[T] {
	if o.IsSome() {
		return Ok(o.Unwrap())
	}
	return Err[T](errFn())
}

// And returns [`None`] if the option is [`None`], otherwise returns `optb`.
func (o Option[T]) And(optb Option[T]) Option[T] {
	if o.IsSome() {
		return optb
	}
	return o
}

// XAnd returns [`None`] if the option is [`None`], otherwise returns `optb`.
func (o Option[T]) XAnd(optb Option[any]) Option[any] {
	if o.IsSome() {
		return optb
	}
	return None[any]()
}

// AndThen returns [`None`] if the option is [`None`], otherwise calls `f` with the
func (o Option[T]) AndThen(f func(T) Option[T]) Option[T] {
	if o.IsNone() {
		return o
	}
	return f(o.unwrapUnchecked())
}

// XAndThen returns [`None`] if the option is [`None`], otherwise calls `f` with the
func (o Option[T]) XAndThen(f func(T) Option[any]) Option[any] {
	if o.IsNone() {
		return None[any]()
	}
	return f(o.unwrapUnchecked())
}

// Filter returns [`None`] if the option is [`None`], otherwise calls `predicate`
// with the wrapped value and returns.
func (o Option[T]) Filter(predicate func(T) bool) Option[T] {
	if o.IsSome() {
		if predicate(o.unwrapUnchecked()) {
			return o
		}
	}
	return None[T]()
}

// Or returns the option if it contains a value, otherwise returns `optb`.
func (o Option[T]) Or(optb Option[T]) Option[T] {
	if o.IsNone() {
		return optb
	}
	return o
}

// OrElse returns [`None`] if the option is [`None`], otherwise calls `f` with the returns the result.
func (o Option[T]) OrElse(f func() Option[T]) Option[T] {
	if o.IsNone() {
		return f()
	}
	return o
}

// Xor [`Some`] if exactly one of `self`, `optb` is [`Some`], otherwise returns [`None`].
func (o Option[T]) Xor(optb Option[T]) Option[T] {
	if o.IsSome() && optb.IsNone() {
		return o
	}
	if o.IsNone() && optb.IsSome() {
		return optb
	}
	return None[T]()
}

// Insert inserts `value` into the option, then returns its pointer.
func (o *Option[T]) Insert(some T) *T {
	v := &some
	o.value = &v
	return v
}

// GetOrInsert inserts `value` into the option if it is [`None`], then
// returns the contained value pointer.
func (o *Option[T]) GetOrInsert(some T) *T {
	if o.IsNone() {
		v := &some
		o.value = &v
	}
	return *o.value
}

// GetOrInsertWith inserts a value computed from `f` into the option if it is [`None`],
// then returns the contained value.
func (o *Option[T]) GetOrInsertWith(f func() T) *T {
	if o.IsNone() {
		var some = f()
		v := &some
		o.value = &v
	}
	return *o.value
}

// Replace replaces the actual value in the option by the value given in parameter,
// returning the old value if present,
// leaving a [`Some`] in its place without deinitializing either one.
func (o *Option[T]) Replace(some T) (old Option[T]) {
	old.value = o.value
	v := &some
	o.value = &v
	return old
}

func (o Option[T]) MarshalJSON() ([]byte, error) {
	if o.IsNone() {
		return []byte("null"), nil
	}
	return json.Marshal(o.value)
}

func (o *Option[T]) UnmarshalJSON(b []byte) error {
	var value = new(T)
	err := json.Unmarshal(b, value)
	if err == nil {
		o.value = &value
	}
	return err
}

var (
	_ Iterable[any]   = Option[any]{}
	_ DeIterable[any] = Option[any]{}
)

func (o Option[T]) Next() Option[T] {
	if o.IsNone() {
		return o
	}
	v := o.Unwrap()
	*o.value = nil
	return Some(v)
}

func (o Option[T]) NextBack() Option[T] {
	return o.Next()
}

func (o Option[T]) Remaining() uint {
	if o.IsNone() {
		return 0
	}
	return 1
}
