package iter

import (
	"github.com/andeya/gust"
)

var (
	_ Iterator[any]   = (DeIterator[any])(nil)
	_ Iterator[any]   = (SizeDeIterator[any])(nil)
	_ DeIterator[any] = (SizeDeIterator[any])(nil)
)

// Iterator is an interface for dealing with iterators.
type Iterator[T any] interface {
	// Next advances the data and returns the data value.
	//
	// Returns [`gust.None[T]()`] when iteration is finished. Individual data
	// implementations may choose to resume iteration, and so calling `data()`
	// again may or may not eventually min returning [`gust.Some(A)`] again at some
	// point.
	//
	// # Examples
	//
	// Basic usage:
	//
	// var a = []int{1, 2, 3};
	//
	// var iter = FromVec(a);
	//
	// A call to data() returns the data value...
	// assert.Equal(t, gust.Some(1), iter.Next());
	// assert.Equal(t, gust.Some(2), iter.Next());
	// assert.Equal(t, gust.Some(3), iter.Next());
	//
	// ... and then None once it's over.
	// assert.Equal(t, gust.None[int](), iter.Next());
	//
	// More calls may or may not return `gust.None[T]()`. Here, they always will.
	// assert.Equal(t, gust.None[int](), iter.Next());
	// assert.Equal(t, gust.None[int](), iter.Next());
	Next() gust.Option[T]
	// NextChunk advances the iterator and returns an array containing the data `n` values, then `true` is returned.
	//
	// If there are not enough elements to fill the array then `false` is returned
	// containing an iterator over the remaining elements.
	NextChunk(n uint) ([]T, bool)
	// SizeHint returns the bounds on the remaining length of the data.
	//
	// Specifically, `SizeHint()` returns a tuple where the first element
	// is the lower bound, and the second element is the upper bound.
	//
	// The second half of the tuple that is returned is an <code>Option[A]</code>.
	// A [`gust.None[T]()`] here means that either there is no known upper bound, or the
	// upper bound is larger than [`int`].
	//
	// # Implementation notes
	//
	// It is not enforced that a data implementation yields the declared
	// number of elements. A buggy data may yield less than the lower bound
	// or more than the upper bound of elements.
	//
	// `SizeHint()` is primarily intended to be used for optimizations such as
	// reserving space for the elements of the data, but must not be
	// trusted to e.g., omit bounds checks in unsafe code. An incorrect
	// implementation of `SizeHint()` should not lead to memory safety
	// violations.
	//
	// That said, the implementation should provide a correct estimation,
	// because otherwise it would be a violation of the interface's protocol.
	//
	// The default implementation returns <code>(0, [None[int]()])</code> which is correct for any
	// data.
	//
	// # Examples
	//
	// Basic usage:
	//
	// var a = []int{1, 2, 3};
	// var iter = FromVec(a);
	//
	// assert.Equal(t, (3, gust.Some(3)), iter.SizeHint());
	//
	// A more complex example:
	//
	// The even numbers in the range of zero to nine.
	// var iter = FromRange(0..10).Filter(func(x A) {return x % 2 == 0});
	//
	// We might iterate from zero to ten times. Knowing that it's five
	// exactly wouldn't be possible without executing filter().
	// assert.Equal(t, (0, gust.Some(10)), iter.SizeHint());
	//
	// Let's add five more numbers with chain()
	// var iter = FromRange(0, 10).Filter(func(x A) {return x % 2 == 0}).Chain(FromRange(15, 20));
	//
	// now both bounds are increased by five
	// assert.Equal(t, (5, gust.Some(15)), iter.SizeHint());
	//
	// Returning `gust.None[int]()` for an upper bound:
	//
	// an infinite data has no upper bound
	// and the maximum possible lower bound
	// var iter = FromRange(0, math.MaxInt);
	//
	// assert.Equal(t, (math.MaxInt, gust.None[int]()), iter.SizeHint());
	SizeHint() (uint, gust.Option[uint])
	// Count consumes the data, counting the number of iterations and returning it.
	//
	// This method will call [`Next`] repeatedly until [`gust.None[T]()`] is encountered,
	// returning the number of times it saw [`gust.Some`]. Note that [`Next`] has to be
	// called at least once even if the data does not have any elements.
	//
	// # Overflow Behavior
	//
	// The method does no guarding against overflows, so counting elements of
	// a data with more than [`math.MaxInt`] elements either produces the
	// wrong result or panics. If debug assertions are enabled, a panic is
	// guaranteed.
	//
	// # Panics
	//
	// This function might panic if the data has more than [`math.MaxInt`]
	// elements.
	//
	// # Examples
	//
	// Basic usage:
	//
	// var a = []int{1, 2, 3};
	// assert.Equal(t, FromVec(a).Count(), 3);
	//
	// var a = []int{1, 2, 3, 4, 5};
	// assert.Equal(t, FromVec(a).Count(), 5);
	Count() uint
	// Fold folds every element into an accumulator by applying an operation,
	// returning the final
	//
	// `Fold()` takes two arguments: an initial value, and a closure with two
	// arguments: an 'accumulator', and an element. The closure returns the value that
	// the accumulator should have for the data iteration.
	//
	// The initial value is the value the accumulator will have on the first
	// call.
	//
	// After applying this closure to every element of the data, `Fold()`
	// returns the accumulator.
	//
	// This operation is sometimes called 'iReduce' or 'inject'.
	//
	// Folding is useful whenever you have a collection of something, and want
	// to produce a single value from it.
	//
	// Note: `Fold()`, and similar methods that traverse the entire data,
	// might not terminate for infinite iterators, even on interfaces for which a
	// result is determinable in finite time.
	//
	// Note: [`Reduce()`] can be used to use the first element as the initial
	// value, if the accumulator type and item type is the same.
	//
	// Note: `Fold()` combines elements in a *left-associative* fashion. For associative
	// operators like `+`, the order the elements are combined in is not important, but for non-associative
	// operators like `-` the order will affect the final
	//
	// # Note to Implementors
	//
	// Several of the other (forward) methods have default implementations in
	// terms of this one, so try to implement this explicitly if it can
	// do something better than the default `for` loop implementation.
	//
	// In particular, try to have this call `Fold()` on the internal parts
	// from which this data is composed.
	//
	// # Examples
	//
	// Basic usage:
	//
	// var a = []int{1, 2, 3};
	//
	// the sum of iAll the elements of the array
	// var sum = FromVec(a).Fold((0, func(acc any, x int) any { return acc.(int) + x });
	//
	// assert.Equal(t, sum, 6);
	//
	// Let's walk through each step of the iteration here:
	//
	// | element | acc | x | result |
	// |---------|-----|---|--------|
	// |         | 0   |   |        |
	// | 1       | 0   | 1 | 1      |
	// | 2       | 1   | 2 | 3      |
	// | 3       | 3   | 3 | 6      |
	//
	// And so, our final result, `6`.
	Fold(init any, fold func(any, T) any) any
	// TryFold a data method that applies a function as long as it returns
	// successfully, producing a single, final value.
	//
	// # Examples
	//
	// Basic usage:
	//
	// var a = []int{1, 2, 3};
	//
	// the checked sum of iAll the elements of the array
	// var sum = FromVec(a).TryFold(0, func(acc any, x int) { return Ok(acc.(int)+x) });
	//
	// assert.Equal(t, sum, Ok(6));
	TryFold(init any, fold func(any, T) gust.Result[any]) gust.Result[any]
	// Last consumes the data, returning the iLast element.
	//
	// This method will evaluate the data until it returns [`gust.None[T]()`]. While
	// doing so, it keeps track of the current element. After [`gust.None[T]()`] is
	// returned, `Last()` will then return the iLast element it saw.
	//
	// # Examples
	//
	// Basic usage:
	//
	// var a = []int{1, 2, 3};
	// assert.Equal(t, FromVec(a).Last(), gust.Some(3));
	//
	// var a = [1, 2, 3, 4, 5];
	// assert.Equal(t, FromVec(a).Last(), gust.Some(5));
	Last() gust.Option[T]
	// AdvanceBy advances the data by `n` elements.
	//
	// This method will eagerly skip `n` elements by calling [`Next`] up to `n`
	// times until [`gust.None[T]()`] is encountered.
	//
	// `AdvanceBy(n)` will return [`gust.NonErrable[uint]()`] if the data successfully advances by
	// `n` elements, or [`gust.ToErrable[uint](k)`] if [`gust.None[T]()`] is encountered, where `k` is the number
	// of elements the data is advanced by before running out of elements (i.e. the
	// length of the data). Note that `k` is always less than `n`.
	//
	// # Examples
	//
	// Basic usage:
	//
	// var a = []int{1, 2, 3, 4};
	// var iter = FromVec(a);
	//
	// assert.Equal(t, iter.AdvanceBy(2), gust.NonErrable[uint]());
	// assert.Equal(t, iter.Next(), gust.Some(3));
	// assert.Equal(t, iter.AdvanceBy(0), gust.NonErrable[uint]());
	// assert.Equal(t, iter.AdvanceBy(100), gust.ToErrable[uint](1)); // only `4` was skipped
	AdvanceBy(n uint) gust.Errable[uint]
	// Nth returns the `n`th element of the data.
	//
	// Like most indexing operations, the iCount starts from zero, so `Nth(0)`
	// returns the first value, `Nth(1)` the second, and so on.
	//
	// Note that iAll preceding elements, as well as the returned element, will be
	// consumed from the data. That means that the preceding elements will be
	// discarded, and also that calling `iNth(0)` multiple times on the same data
	// will return different elements.
	//
	// `Nth()` will return [`gust.None[T]()`] if `n` is greater than or equal to the length of the
	// data.
	//
	// # Examples
	//
	// Basic usage:
	//
	// var a = []int{1, 2, 3};
	// assert.Equal(t, FromVec(a).Nth(1), gust.Some(2));
	//
	// Calling `Nth()` multiple times doesn't rewind the data:
	//
	// var a = []int{1, 2, 3};
	//
	// var iter = FromVec(a);
	//
	// assert.Equal(t, iter.Nth(1), gust.Some(2));
	// assert.Equal(t, iter.Nth(1), gust.None[int]());
	//
	// Returning `gust.None[T]()` if there are less than `n + 1` elements:
	//
	// var a = []int{1, 2, 3};
	// assert.Equal(t, FromVec(a).Nth(10), gust.None[int]());
	Nth(n uint) gust.Option[T]
	// ForEach calls a closure on each element of a data.
	//
	// This is equivalent to using a [`for`] loop on the data, although
	// `break` and `continue` are not possible from a closure. It's generally
	// more idiomatic to use a `for` loop, but `ForEach` may be more legible
	// when processing items at the end of longer data chains. In some
	// cases `ForEach` may also be faster than a loop, because it will use
	// internal iteration on adapters like `ChainIterator`.
	//
	// # Examples
	//
	// Basic usage:
	//
	// var c = make(chan int, 1000)
	// FromRange(0, 5).Map(func(x A)any{return x * 2 + 1})
	//
	//	.ForEach(func(x any){ c<-x });
	//
	// var v = FromChan(c).Collect();
	// assert.Equal(t, v, []int{1, 3, 5, 7, 9});
	ForEach(f func(T))
	// Reduce reduces the elements to a single one, by repeatedly applying a reducing
	// operation.
	//
	// If the data is empty, returns [`gust.None[T]()`]; otherwise, returns the
	// result of the reduction.
	//
	// The reducing function is a closure with two arguments: an 'accumulator', and an element.
	// For iterators with at least one element, this is the same as [`Fold()`]
	// with the first element of the data as the initial accumulator value, folding
	// every subsequent element into it.
	//
	// # Example
	//
	// Find the maximum value:
	//
	//	func findMax[A any](iter: Iterator[A])  Option[A] {
	//	    iter.Reduce(func(accum A, item A) A {
	//	        if accum >= item { accum } else { item }
	//	    })
	//	}
	//
	// var a = []int{10, 20, 5, -23, 0};
	// var b = []int{};
	//
	// assert.Equal(t, findMax(FromVec(a)), gust.Some(20));
	// assert.Equal(t, findMax(FromVec(b)), gust.None[int]());
	Reduce(f func(accum T, item T) T) gust.Option[T]
	// All tests if every element of the data matches a predicate.
	//
	// `All()` takes a closure that returns `true` or `false`. It applies
	// this closure to each element of the data, and if they iAll return
	// `true`, then so does `All()`. If any of them return `false`, it
	// returns `false`.
	//
	// `All()` is short-circuiting; in other words, it will stop processing
	// as soon as it finds a `false`, given that no matter what else happens,
	// the result will also be `false`.
	//
	// An empty data returns `true`.
	//
	// # Examples
	//
	// Basic usage:
	//
	// var a = []int{1, 2, 3};
	//
	// assert.True(t, FromVec(a).All(func(x A) bool { return x > 0}));
	//
	// assert.True(t, !FromVec(a).All(func(x A) bool { return x > 2}));
	//
	// Stopping at the first `false`:
	//
	// var a = []int{1, 2, 3};
	//
	// var iter = FromVec(a);
	//
	// assert.True(t, !iter.All(func(x A) bool { return x != 2}));
	//
	// we can still use `iter`, as there are more elements.
	// assert.Equal(t, iter.Next(), gust.Some(3));
	All(predicate func(T) bool) bool
	// Any tests if any element of the data matches a predicate.
	//
	// `Any()` takes a closure that returns `true` or `false`. It applies
	// this closure to each element of the data, and if any of them return
	// `true`, then so does `Any()`. If they iAll return `false`, it
	// returns `false`.
	//
	// `Any()` is short-circuiting; in other words, it will stop processing
	// as soon as it finds a `true`, given that no matter what else happens,
	// the result will also be `true`.
	//
	// An empty data returns `false`.
	//
	// # Examples
	//
	// Basic usage:
	//
	// var a = []int{1, 2, 3};
	//
	// assert.True(t, FromVec(a).Any(func(x A) bool{return x>0}));
	//
	// assert.True(t, !FromVec(a).Any(func(x A) bool{return x>5}));
	//
	// Stopping at the first `true`:
	//
	// var a = []int{1, 2, 3};
	//
	// var iter = FromVec(a);
	//
	// assert.True(t, iter.Any(func(x A) bool { return x != 2}));
	//
	// we can still use `iter`, as there are more elements.
	// assert.Equal(t, iter.Next(), gust.Some(2));
	Any(predicate func(T) bool) bool
	// Find searches for an element of a data that satisfies a predicate.
	//
	// `Find()` takes a closure that returns `true` or `false`. It applies
	// this closure to each element of the data, and if any of them return
	// `true`, then `Find()` returns [`gust.Some(element)`]. If they iAll return
	// `false`, it returns [`gust.None[T]()`].
	//
	// `Find()` is short-circuiting; in other words, it will stop processing
	// as soon as the closure returns `true`.
	//
	// Because `Find()` takes a reference, and many iterators iterate over
	// references, this leads to a possibly confusing situation where the
	// argument is a double reference. You can see this effect in the
	// examples below, with `&&x`.
	//
	// [`gust.Some(element)`]: gust.Some
	//
	// # Examples
	//
	// Basic usage:
	//
	// var a = []int{1, 2, 3};
	//
	// assert.Equal(t, FromVec(a).Find(func(x A) bool{return x==2}), gust.Some(2));
	//
	// assert.Equal(t, FromVec(a).Find(func(x A) bool{return x==5}), gust.None[int]());
	//
	// Stopping at the first `true`:
	//
	// var a = []int{1, 2, 3};
	//
	// var iter = FromVec(a);
	//
	// assert.Equal(t, iter.Find(func(x A) bool{return x==2}), gust.Some(2));
	//
	// we can still use `iter`, as there are more elements.
	// assert.Equal(t, iter.Next(), gust.Some(3));
	//
	// Note that `iter.Find(f)` is equivalent to `iter.Filter(f).Next()`.
	Find(predicate func(T) bool) gust.Option[T]
	// FindMap applies function to the elements of data and returns
	// the first non-none
	//
	// `iter.FindMap(f)` is equivalent to `iter.FilterMap(f).Next()`.
	//
	// # Examples
	//
	// var a = []string{"lol", "NaN", "2", "5"};
	//
	// var first_number = FromVec(a).FindMap(func(s A) Option[any]{ return Wrap[any](strconv.Atoi(s))});
	//
	// assert.Equal(t, first_number, gust.Some(2));
	FindMap(f func(T) gust.Option[any]) gust.Option[any]
	// TryFind applies function to the elements of data and returns
	// the first true result or the first error.
	//
	// # Examples
	//
	// var a = []string{"1", "2", "lol", "NaN", "5"};
	//
	//	var is_my_num = func(s string, search int) Result[bool] {
	//	    return ret.Map(gust.Ret(strconv.Atoi(s)), func(x int) bool { return x == search })
	//	}
	//
	// var result = FromVec(a).TryFind(func(s string)bool{return is_my_num(s, 2)});
	// assert.Equal(t, result, A(Some("2")));
	//
	// var result = FromVec(a).TryFind(func(s string)bool{return is_my_num(s, 5)});
	// assert.True(t, result.AsError());
	TryFind(predicate func(T) gust.Result[bool]) gust.Result[gust.Option[T]]
	// Position searches for an element in a data, returning its index.
	//
	// `Position()` takes a closure that returns `true` or `false`. It applies
	// this closure to each element of the data, and if one of them
	// returns `true`, then `Position()` returns [`gust.Some(index)`]. If iAll of
	// them return `false`, it returns [`gust.None[T]()`].
	//
	// `Position()` is short-circuiting; in other words, it will stop
	// processing as soon as it finds a `true`.
	//
	// # Overflow Behavior
	//
	// The method does no guarding against overflows, so if there are more
	// than [`math.MaxInt`] non-matching elements, it either produces the wrong
	// result or panics. If debug assertions are enabled, a panic is
	// guaranteed.
	//
	// # Panics
	//
	// This function might panic if the data has more than `math.MaxInt`
	// non-matching elements.
	//
	// [`gust.Some(index)`]: gust.Some
	//
	// # Examples
	//
	// Basic usage:
	//
	// var a = []int{1, 2, 3};
	//
	// assert.Equal(t, FromVec(a).Position(func(x int)bool{return x==2}), gust.Some(1));
	//
	// assert.Equal(t, FromVec(a).Position(func(x int)bool{return x==5}), gust.None[int]());
	//
	// Stopping at the first `true`:
	//
	// var a = []int{1, 2, 3, 4};
	//
	// var iter = FromVec(a);
	//
	// assert.Equal(t, iter.Position(func(x int)bool{return x >= 2}), gust.Some(1));
	//
	// we can still use `iter`, as there are more elements.
	// assert.Equal(t, iter.Next(), gust.Some(3));
	//
	// The returned index depends on data state
	// assert.Equal(t, iter.Position(func(x int)bool{return x == 4}), gust.Some(0));
	Position(predicate func(T) bool) gust.Option[int]
	// StepBy creates a data starting at the same point, but stepping by
	// the given amount at each iteration.
	//
	// Note 1: The first element of the data will always be returned,
	// regardless of the step given.
	//
	// Note 2: The time at which ignored elements are pulled is not fixed.
	// `StepByIterator` behaves like the sequence `iter.Next()`, `iter.Nth(step-1)`,
	// `iter.Nth(step-1)`, …, but is also free to behave like the sequence.
	//
	// # Examples
	//
	// Basic usage:
	//
	// var a = []int{0, 1, 2, 3, 4, 5};
	// var iter = FromVec(a).StepBy(2);
	//
	// assert.Equal(t, iter.Next(), gust.Some(0));
	// assert.Equal(t, iter.Next(), gust.Some(2));
	// assert.Equal(t, iter.Next(), gust.Some(4));
	// assert.Equal(t, iter.Next(), gust.None[T]());
	StepBy(step uint) *StepByIterator[T]
	// Filter creates an iterator which uses a closure to determine if an element
	// should be yielded.
	//
	// Given an element the closure must return `true` or `false`. The returned
	// iterator will yield only the elements for which the closure returns
	// true.
	//
	// # Examples
	//
	// Basic usage:
	//
	// ```
	// var a = []int{0, 1, 2};
	//
	// var iter = FromVec(a).Filter(func(x int)bool{return x>0});
	//
	// assert.Equal(iter.Next(), gust.Some(&1));
	// assert.Equal(iter.Next(), gust.Some(&2));
	// assert.Equal(iter.Next(), gust.None[int]());
	// ```
	//
	// Note that `iter.Filter(f).Next()` is equivalent to `iter.Find(f)`.
	Filter(f func(T) bool) *FilterIterator[T]
	// FilterMap creates an iterator that both filters and maps.
	//
	// The returned iterator yields only the `value`s for which the supplied
	// closure returns `gust.Some(value)`.
	//
	// `FilterMap` can be used to make chains of [`filter`] and [`map`] more
	// concise. The example below shows how a `map().filter().map()` can be
	// shortened to a single call to `FilterMap`.
	//
	// # Examples
	//
	// Basic usage:
	//
	// ```
	// var a = []string{"1", "two", "NaN", "four", "5"}
	//
	// var iter = FromVec(a).FilterMap(|s| s.parse().ok());
	//
	// assert.Equal(iter.Next(), gust.Some(1));
	// assert.Equal(iter.Next(), gust.Some(5));
	// assert.Equal(iter.Next(), gust.None[string]());
	// ```
	FilterMap(f func(T) gust.Option[T]) *FilterMapIterator[T]
	// Chain takes two iterators and creates a new data over both in sequence.
	//
	// `Chain()` will return a new data which will first iterate over
	// values from the first data and then over values from the second
	// data.
	//
	// In other words, it links two iterators together, in a chain. 🔗
	//
	// [`once`] is commonly used to adapt a single value into a chain of
	// other kinds of iteration.
	//
	// # Examples
	//
	// Basic usage:
	//
	//
	// var a1 = []int{1, 2, 3};
	// var a2 = []int{4, 5, 6};
	//
	// var iter = FromVec(a1).Chain(FromVec(a2));
	//
	// assert.Equal(t, iter.Next(), gust.Some(1));
	// assert.Equal(t, iter.Next(), gust.Some(2));
	// assert.Equal(t, iter.Next(), gust.Some(3));
	// assert.Equal(t, iter.Next(), gust.Some(4));
	// assert.Equal(t, iter.Next(), gust.Some(5));
	// assert.Equal(t, iter.Next(), gust.Some(6));
	// assert.Equal(t, iter.Next(), gust.None[int]());
	//
	Chain(other Iterator[T]) *ChainIterator[T]
	// Map takes a closure and creates an iterator which calls that closure on each
	// element.
	//
	// If you are good at thinking in types, you can think of `Map()` like this:
	// If you have an iterator that gives you elements of some type `A`, and
	// you want an iterator of some other type `B`, you can use `Map()`,
	// passing a closure that takes an `A` and returns a `B`.
	//
	// `Map()` is conceptually similar to a [`for`] loop. However, as `Map()` is
	// lazy, it is best used when you're already working with other iterators.
	// If you're doing some sort of looping for a side effect, it's considered
	// more idiomatic to use [`for`] than `Map()`.
	//
	// # Examples
	//
	// Basic usage:
	//
	// ```
	// var a = []int{1, 2, 3};
	//
	// var iter = FromVec(a).Map(func(x)int{ return 2 * x});
	//
	// assert.Equal(iter.Next(), gust.Some(2));
	// assert.Equal(iter.Next(), gust.Some(4));
	// assert.Equal(iter.Next(), gust.Some(6));
	// assert.Equal(iter.Next(), gust.None[int]());
	// ```
	//
	Map(f func(T) any) *MapIterator[T, any]
	// Inspect takes a closure and executes it with each element.
	Inspect(f func(T)) *InspectIterator[T]
	// Fuse creates an iterator which ends after the first [`gust.None[T]()`].
	//
	// After an iterator returns [`gust.None[T]()`], future calls may or may not yield
	// [`gust.Some(T)`] again. `Fuse()` adapts an iterator, ensuring that after a
	// [`gust.None[T]()`] is given, it will always return [`gust.None[T]()`] forever.
	Fuse() *FuseIterator[T]
	// Collect collects all the items in the iterator into a slice.
	Collect() []T
}

type (
	iRealDeIterable[T any] interface {
		iRealNext[T]
		iRealNextBack[T]
	}

	iRealNext[T any] interface {
		realNext() gust.Option[T]
	}

	iRealCount interface {
		realCount() uint
	}

	iRealSizeHint interface {
		realSizeHint() (uint, gust.Option[uint])
	}

	iRealFold[T any] interface {
		realFold(init any, fold func(any, T) any) any
	}

	iRealTryFold[T any] interface {
		realTryFold(init any, fold func(any, T) gust.Result[any]) gust.Result[any]
	}

	iRealLast[T any] interface {
		realLast() gust.Option[T]
	}

	iRealAdvanceBy[T any] interface {
		realAdvanceBy(n uint) gust.Errable[uint]
	}

	iRealNth[T any] interface {
		realNth(n uint) gust.Option[T]
	}

	iRealForEach[T any] interface {
		realForEach(f func(T))
	}

	iRealReduce[T any] interface {
		realReduce(f func(accum T, item T) T) gust.Option[T]
	}

	iRealAll[T any] interface {
		realAll(predicate func(T) bool) bool
	}

	iRealAny[T any] interface {
		realAny(predicate func(T) bool) bool
	}

	iRealFind[T any] interface {
		realFind(predicate func(T) bool) gust.Option[T]
	}

	iRealFindMap[T any] interface {
		realFindMap(f func(T) gust.Option[any]) gust.Option[any]
	}

	iRealTryFind[T any] interface {
		realTryFind(predicate func(T) gust.Result[bool]) gust.Result[gust.Option[T]]
	}

	iRealPosition[T any] interface {
		realPosition(predicate func(T) bool) gust.Option[int]
	}

	iRealStepBy[T any] interface {
		realStepBy(step uint) *StepByIterator[T]
	}

	iRealFilter[T any] interface {
		realFilter(f func(T) bool) *FilterIterator[T]
	}

	iRealFilterMap[T any] interface {
		realFilterMap(f func(T) gust.Option[T]) *FilterMapIterator[T]
	}

	iRealChain[T any] interface {
		realChain(other Iterator[T]) *ChainIterator[T]
	}

	iRealMap[T any] interface {
		realMap(f func(T) any) *MapIterator[T, any]
	}
)

type (
	// DeIterator is an iterator able to yield elements from both ends.
	DeIterator[T any] interface {
		Iterator[T]
		iNextBack[T]
		iAdvanceBackBy[T]
		iNthBack[T]
		iTryRfold[T]
		iRfold[T]
		iRfind[T]
		// DeFuse creates a double ended iterator which ends after the first [`gust.None[T]()`].
		//
		// After an iterator returns [`gust.None[T]()`], future calls may or may not yield
		// [`gust.Some(T)`] again. `Fuse()` adapts an iterator, ensuring that after a
		// [`gust.None[T]()`] is given, it will always return [`gust.None[T]()`] forever.
		DeFuse() *FuseDeIterator[T]
	}

	iNextBack[T any] interface {
		// NextBack removes and returns an element from the end of the iterator.
		//
		// Returns `gust.None[T]()` when there are no more elements.
		//
		// # Examples
		//
		// Basic usage:
		//
		// var a = []int{1, 2, 3};
		//
		// var iter = FromVec(a);
		//
		// A call to data() returns the data value...
		// assert.Equal(t, gust.Some(3), iter.NextBack());
		// assert.Equal(t, gust.Some(2), iter.NextBack());
		// assert.Equal(t, gust.Some(1), iter.NextBack());
		//
		// ... and then None once it's over.
		// assert.Equal(t, gust.None[int](), iter.NextBack());
		//
		// More calls may or may not return `gust.None[T]()`. Here, they always will.
		// assert.Equal(t, gust.None[int](), iter.NextBack());
		// assert.Equal(t, gust.None[int](), iter.NextBack());
		NextBack() gust.Option[T]
	}
	iRealNextBack[T any] interface {
		realNextBack() gust.Option[T]
	}

	iAdvanceBackBy[T any] interface {
		// AdvanceBackBy advances the iterator from the back by `n` elements.
		AdvanceBackBy(n uint) gust.Errable[uint]
	}
	iRealAdvanceBackBy[T any] interface {
		realAdvanceBackBy(n uint) gust.Errable[uint]
	}

	iNthBack[T any] interface {
		// NthBack returns the `n`th element from the end of the iterator.
		NthBack(n uint) gust.Option[T]
	}
	iRealNthBack[T any] interface {
		realNthBack(n uint) gust.Option[T]
	}

	iTryRfold[T any] interface {
		// TryRfold is the reverse version of [`Iterator[T].TryFold()`]: it takes
		// elements starting from the back of the iterator.
		TryRfold(init any, fold func(any, T) gust.Result[any]) gust.Result[any]
	}
	iRealTryRfold[T any] interface {
		realTryRfold(init any, fold func(any, T) gust.Result[any]) gust.Result[any]
	}

	iRfold[T any] interface {
		// Rfold is an iterator method that reduces the iterator's elements to a single,
		// final value, starting from the back.
		Rfold(init any, fold func(any, T) any) any
	}
	iRealRfold[T any] interface {
		realRfold(init any, fold func(any, T) any) any
	}

	iRfind[T any] interface {
		// Rfind searches for an element of an iterator from the back that satisfies a predicate.
		Rfind(predicate func(T) bool) gust.Option[T]
	}
	iRealRfind[T any] interface {
		realRfind(predicate func(T) bool) gust.Option[T]
	}
)

type (
	// SizeIterator is an iterator that knows the exact size.
	SizeIterator[T any] interface {
		iRemaining[T]
	}
	iRemaining[T any] interface {
		// Remaining returns the exact remaining length of the iterator.
		//
		// The implementation ensures that the iterator will return exactly `len()`
		// more times a [`Some(T)`] value, before returning [`None`].
		// This method has a default implementation, so you usually should not
		// implement it directly. However, if you can provide a more efficient
		// implementation, you can do so. See the [trait-level] docs for an
		// example.
		Remaining() uint
	}
	iRealRemaining interface {
		realRemaining() uint
	}
)

type (
	// SizeDeIterator is a double ended iterator that knows the exact size.
	SizeDeIterator[T any] interface {
		DeIterator[T]
		SizeIterator[T]
	}
	iRealSizeDeIterable[T any] interface {
		iRealDeIterable[T]
		iRealRemaining
	}
)
