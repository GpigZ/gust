package iter

import (
	"github.com/andeya/gust"
)

var (
	_ Iterator[any]       = (*ChainIterator[any])(nil)
	_ iRealLast[any]      = (*ChainIterator[any])(nil)
	_ iRealFind[any]      = (*ChainIterator[any])(nil)
	_ iRealNext[any]      = (*ChainIterator[any])(nil)
	_ iRealSizeHint       = (*ChainIterator[any])(nil)
	_ iRealCount          = (*ChainIterator[any])(nil)
	_ iRealTryFold[any]   = (*ChainIterator[any])(nil)
	_ iRealFold[any]      = (*ChainIterator[any])(nil)
	_ iRealAdvanceBy[any] = (*ChainIterator[any])(nil)
	_ iRealNth[any]       = (*ChainIterator[any])(nil)
)

func newChainIterator[T any](inner Iterator[T], other Iterator[T]) *ChainIterator[T] {
	iter := &ChainIterator[T]{inner: inner, other: other}
	iter.setFacade(iter)
	return iter
}

type ChainIterator[T any] struct {
	iterTrait[T]
	inner Iterator[T]
	other Iterator[T]
}

func (s *ChainIterator[T]) realLast() gust.Option[T] {
	// Must exhaust a before b.
	var aLast gust.Option[T]
	var bLast gust.Option[T]
	if s.inner != nil {
		aLast = s.inner.Last()
	}
	if s.other != nil {
		bLast = s.other.Last()
	}
	if bLast.IsSome() {
		return bLast
	}
	return aLast
}

func (s *ChainIterator[T]) realFind(predicate func(T) bool) gust.Option[T] {
	if s.inner != nil {
		item := s.inner.Find(predicate)
		if item.IsSome() {
			return item
		}
		s.inner = nil
	}
	if s.other != nil {
		return s.other.Find(predicate)
	}
	return gust.None[T]()
}

func (s *ChainIterator[T]) realNth(n uint) gust.Option[T] {
	if s.inner != nil {
		r := s.inner.AdvanceBy(n)
		if r.AsError() {
			n -= r.Unwrap()
		} else {
			item := s.inner.Next()
			if item.IsSome() {
				return item
			}
			n = 0
		}
		s.inner = nil
	}
	if s.other != nil {
		return s.other.Nth(n)
	}
	return gust.None[T]()
}

func (s *ChainIterator[T]) realAdvanceBy(n uint) gust.Errable[uint] {
	var rem = n
	if s.inner != nil {
		r := s.inner.AdvanceBy(rem)
		if !r.AsError() {
			return r
		}
		rem -= r.Unwrap()
		s.inner = nil
	}
	if s.other != nil {
		r := s.other.AdvanceBy(rem)
		if !r.AsError() {
			return r
		}
		rem -= r.Unwrap()
		// we don't fuse the second iterator
	}
	if rem == 0 {
		return gust.NonErrable[uint]()
	}
	return gust.ToErrable(n - rem)
}

func (s *ChainIterator[T]) realFold(acc any, f func(any, T) any) any {
	if s.inner != nil {
		acc = s.inner.Fold(acc, f)
	}
	if s.other != nil {
		acc = s.other.Fold(acc, f)
	}
	return acc
}

func (s *ChainIterator[T]) realTryFold(acc any, f func(any, T) gust.Result[any]) gust.Result[any] {
	if s.inner != nil {
		r := s.inner.TryFold(acc, f)
		if r.IsErr() {
			return r
		}
		acc = r.Unwrap()
		s.inner = nil
	}
	if s.other != nil {
		r := s.other.TryFold(acc, f)
		if r.IsErr() {
			return r
		}
		acc = r.Unwrap()
		// we don't fuse the second iterator
	}
	return gust.Ok(acc)
}

func (s *ChainIterator[T]) realNext() gust.Option[T] {
	if s.inner != nil {
		item := s.inner.Next()
		if item.IsSome() {
			return item
		}
		s.inner = nil
	}
	if s.other != nil {
		return s.other.Next()
	}
	return gust.None[T]()
}

func (s *ChainIterator[T]) realSizeHint() (uint, gust.Option[uint]) {
	if s.inner != nil && s.other != nil {
		var aLower, aUpper = s.inner.SizeHint()
		var bLower, bUpper = s.other.SizeHint()
		var lower = saturatingAdd(aLower, bLower)
		var upper gust.Option[uint]
		if aUpper.IsSome() && bUpper.IsSome() {
			upper = checkedAdd(aUpper.Unwrap(), bUpper.Unwrap())
		}
		return lower, upper
	}
	if s.inner != nil && s.other == nil {
		return s.inner.SizeHint()
	}
	if s.inner == nil && s.other != nil {
		return s.other.SizeHint()
	}
	return 0, gust.Some[uint](0)
}

func (s *ChainIterator[T]) realCount() uint {
	var aCount uint
	if s.inner != nil {
		aCount = s.inner.Count()
	}
	var bCount uint
	if s.other != nil {
		bCount = s.other.Count()
	}
	return aCount + bCount
}
