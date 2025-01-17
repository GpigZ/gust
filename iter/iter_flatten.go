package iter

import (
	"github.com/andeya/gust"
	"github.com/andeya/gust/opt"
)

var (
	_ Iterator[any]       = (*FlattenIterator[any, gust.Iterable[any]])(nil)
	_ iRealNext[any]      = (*FlattenIterator[any, gust.Iterable[any]])(nil)
	_ iRealSizeHint       = (*FlattenIterator[any, gust.Iterable[any]])(nil)
	_ iRealTryFold[any]   = (*FlattenIterator[any, gust.Iterable[any]])(nil)
	_ iRealFold[any]      = (*FlattenIterator[any, gust.Iterable[any]])(nil)
	_ iRealAdvanceBy[any] = (*FlattenIterator[any, gust.Iterable[any]])(nil)
	_ iRealCount          = (*FlattenIterator[any, gust.Iterable[any]])(nil)
	_ iRealLast[any]      = (*FlattenIterator[any, gust.Iterable[any]])(nil)
)

func newFlattenIterator[T any, D gust.Iterable[T]](iter Iterator[D]) *FlattenIterator[T, D] {
	p := &FlattenIterator[T, D]{iter: iter.Fuse()}
	p.setFacade(p)
	return p
}

type FlattenIterator[T any, D gust.Iterable[T]] struct {
	iterTrait[T]
	iter      *FuseIterator[D]
	frontiter gust.Option[Iterator[T]]
	backiter  gust.Option[Iterator[T]]
}

func (f FlattenIterator[T, D]) realNext() gust.Option[T] {
	for {
		if f.frontiter.IsSome() {
			x := f.frontiter.Unwrap().Next().InspectNone(func() {
				f.frontiter = gust.None[Iterator[T]]()
			})
			if x.IsSome() {
				return x
			}
		} else {
			x := f.iter.Next().Inspect(func(t D) {
				f.frontiter = gust.Some(FromIterable[T](t))
			})
			if x.IsNone() {
				if f.backiter.IsSome() {
					return f.backiter.Unwrap().Next().InspectNone(func() {
						f.backiter = gust.None[Iterator[T]]()
					})
				}
				return gust.None[T]()
			}
		}
	}
}

func (f FlattenIterator[T, D]) realSizeHint() (uint, gust.Option[uint]) {
	var fl = opt.MapOr(f.frontiter, gust.Pair[uint, gust.Option[uint]]{0, gust.Some[uint](0)}, func(i Iterator[T]) (x gust.Pair[uint, gust.Option[uint]]) {
		x.A, x.B = i.SizeHint()
		return x
	})
	var bl = opt.MapOr(f.backiter, gust.Pair[uint, gust.Option[uint]]{0, gust.Some[uint](0)}, func(i Iterator[T]) (x gust.Pair[uint, gust.Option[uint]]) {
		x.A, x.B = i.SizeHint()
		return x
	})
	var lo = saturatingAdd(fl.A, bl.A)
	// TODO: check fixed size
	var a, b = f.iter.SizeHint()
	if a == 0 && b.IsSome() && b.Unwrap() == 0 && fl.B.IsSome() && bl.B.IsSome() {
		return lo, checkedAdd(fl.B.Unwrap(), bl.B.Unwrap())
	}
	return lo, gust.None[uint]()
}

func (f FlattenIterator[T, D]) realTryFold(init any, fold func(any, T) gust.Result[any]) gust.Result[any] {
	return f.iterTryFold(init, func(acc any, item Iterator[T]) gust.Result[any] {
		return item.TryFold(acc, fold)
	})
}

func (f FlattenIterator[T, D]) realFold(init any, fold func(any, T) any) any {
	return f.iterFold(init, func(acc any, item Iterator[T]) any {
		return item.Fold(acc, fold)
	})
}

func (f FlattenIterator[T, D]) iterTryFold(acc any, fold func(any, Iterator[T]) gust.Result[any]) gust.Result[any] {
	var flatten = func(frontiter *gust.Option[Iterator[T]], fold func(any, Iterator[T]) gust.Result[any]) func(any, D) gust.Result[any] {
		return func(acc any, iter D) gust.Result[any] {
			return fold(acc, *frontiter.Insert(FromIterable[T](iter)))
		}
	}
	if f.frontiter.IsSome() {
		x := fold(acc, f.frontiter.Unwrap())
		if x.IsErr() {
			return x
		}
		acc = x
		f.frontiter = gust.None[Iterator[T]]()
	}

	acc = f.iter.TryFold(acc, flatten(&f.frontiter, fold))
	f.frontiter = gust.None[Iterator[T]]()

	if f.backiter.IsSome() {
		x := fold(acc, f.backiter.Unwrap())
		if x.IsErr() {
			return x
		}
		acc = x
		f.backiter = gust.None[Iterator[T]]()
	}
	return gust.Ok(acc)
}

func (f FlattenIterator[T, D]) iterFold(acc any, fold func(any, Iterator[T]) any) any {
	f.frontiter.Inspect(func(i Iterator[T]) {
		acc = fold(acc, i)
	})
	acc = f.iter.Fold(acc, func(acc any, i D) any {
		return fold(acc, FromIterable[T](i))
	})
	f.backiter.Inspect(func(i Iterator[T]) {
		acc = fold(acc, i)
	})
	return acc
}

func (f FlattenIterator[T, D]) realAdvanceBy(n uint) gust.Errable[uint] {
	var advance = func(n any, iter Iterator[T]) gust.Result[any] {
		x := iter.AdvanceBy(n.(uint))
		if x.AsError() {
			return gust.Err[any](n.(uint) - x.Unwrap())
		}
		return gust.Ok[any](nil)
	}
	x := f.iterTryFold(n, advance)
	if x.IsErr() {
		remaining := x.ErrVal().(uint)
		if remaining > 0 {
			return gust.ToErrable(n - remaining)
		}
	}
	return gust.NonErrable[uint]()
}

func (f FlattenIterator[T, D]) realCount() uint {
	var count = func(acc any, iter Iterator[T]) any {
		return acc.(uint) + iter.Count()
	}
	return f.iterFold(uint(0), count).(uint)
}

func (f FlattenIterator[T, D]) realLast() gust.Option[T] {
	var last = func(last any, iter Iterator[T]) any {
		return iter.Last().Or(last.(gust.Option[T]))
	}
	return f.iterFold(gust.None[T](), last).(gust.Option[T])
}

var (
	_ DeIterator[any]         = (*FlattenDeIterator[any, gust.DeIterable[any]])(nil)
	_ iRealNext[any]          = (*FlattenDeIterator[any, gust.DeIterable[any]])(nil)
	_ iRealNextBack[any]      = (*FlattenDeIterator[any, gust.DeIterable[any]])(nil)
	_ iRealSizeHint           = (*FlattenDeIterator[any, gust.DeIterable[any]])(nil)
	_ iRealTryFold[any]       = (*FlattenDeIterator[any, gust.DeIterable[any]])(nil)
	_ iRealTryRfold[any]      = (*FlattenDeIterator[any, gust.DeIterable[any]])(nil)
	_ iRealFold[any]          = (*FlattenDeIterator[any, gust.DeIterable[any]])(nil)
	_ iRealRfold[any]         = (*FlattenDeIterator[any, gust.DeIterable[any]])(nil)
	_ iRealAdvanceBy[any]     = (*FlattenDeIterator[any, gust.DeIterable[any]])(nil)
	_ iRealAdvanceBackBy[any] = (*FlattenDeIterator[any, gust.DeIterable[any]])(nil)
	_ iRealCount              = (*FlattenDeIterator[any, gust.DeIterable[any]])(nil)
	_ iRealLast[any]          = (*FlattenDeIterator[any, gust.DeIterable[any]])(nil)
)

func newFlattenDeIterator[T any, D gust.DeIterable[T]](iter DeIterator[D]) *FlattenDeIterator[T, D] {
	p := &FlattenDeIterator[T, D]{iter: iter.DeFuse()}
	p.setFacade(p)
	return p
}

type FlattenDeIterator[T any, D gust.DeIterable[T]] struct {
	sizeDeIterTrait[T]
	iter      *FuseDeIterator[D]
	frontiter gust.Option[DeIterator[T]]
	backiter  gust.Option[DeIterator[T]]
}

func (f FlattenDeIterator[T, D]) realNextBack() gust.Option[T] {
	for {
		if f.backiter.IsSome() {
			x := f.backiter.Unwrap().NextBack().InspectNone(func() {
				f.backiter = gust.None[DeIterator[T]]()
			})
			if x.IsSome() {
				return x
			}
		} else {
			x := f.iter.NextBack().Inspect(func(t D) {
				f.backiter = gust.Some(FromDeIterable[T](t))
			})
			if x.IsNone() {
				if f.frontiter.IsSome() {
					return f.frontiter.Unwrap().NextBack().InspectNone(func() {
						f.frontiter = gust.None[DeIterator[T]]()
					})
				}
				return gust.None[T]()
			}
		}
	}
}

func (f FlattenDeIterator[T, D]) realNext() gust.Option[T] {
	for {
		if f.frontiter.IsSome() {
			x := f.frontiter.Unwrap().Next().InspectNone(func() {
				f.frontiter = gust.None[DeIterator[T]]()
			})
			if x.IsSome() {
				return x
			}
		} else {
			x := f.iter.Next().Inspect(func(t D) {
				f.frontiter = gust.Some(FromDeIterable[T](t))
			})
			if x.IsNone() {
				if f.backiter.IsSome() {
					return f.backiter.Unwrap().Next().InspectNone(func() {
						f.backiter = gust.None[DeIterator[T]]()
					})
				}
				return gust.None[T]()
			}
		}
	}
}

func (f FlattenDeIterator[T, D]) realSizeHint() (uint, gust.Option[uint]) {
	var fl = opt.MapOr(f.frontiter, gust.Pair[uint, gust.Option[uint]]{0, gust.Some[uint](0)}, func(i DeIterator[T]) (x gust.Pair[uint, gust.Option[uint]]) {
		x.A, x.B = i.SizeHint()
		return x
	})
	var bl = opt.MapOr(f.backiter, gust.Pair[uint, gust.Option[uint]]{0, gust.Some[uint](0)}, func(i DeIterator[T]) (x gust.Pair[uint, gust.Option[uint]]) {
		x.A, x.B = i.SizeHint()
		return x
	})
	var lo = saturatingAdd(fl.A, bl.A)
	// TODO: check fixed size
	var a, b = f.iter.SizeHint()
	if a == 0 && b.IsSome() && b.Unwrap() == 0 && fl.B.IsSome() && bl.B.IsSome() {
		return lo, checkedAdd(fl.B.Unwrap(), bl.B.Unwrap())
	}
	return lo, gust.None[uint]()
}

func (f FlattenDeIterator[T, D]) realTryFold(init any, fold func(any, T) gust.Result[any]) gust.Result[any] {
	return f.iterTryFold(init, func(acc any, item DeIterator[T]) gust.Result[any] {
		return item.TryFold(acc, fold)
	})
}

func (f FlattenDeIterator[T, D]) realTryRfold(init any, fold func(any, T) gust.Result[any]) gust.Result[any] {
	return f.iterTryRfold(init, func(acc any, item DeIterator[T]) gust.Result[any] {
		return item.TryRfold(acc, fold)
	})
}

func (f FlattenDeIterator[T, D]) realFold(init any, fold func(any, T) any) any {
	return f.iterFold(init, func(acc any, item DeIterator[T]) any {
		return item.Fold(acc, fold)
	})
}

func (f FlattenDeIterator[T, D]) realRfold(init any, fold func(any, T) any) any {
	return f.iterRfold(init, func(acc any, item DeIterator[T]) any {
		return item.Rfold(acc, fold)
	})
}

func (f FlattenDeIterator[T, D]) realAdvanceBy(n uint) gust.Errable[uint] {
	var advance = func(n any, iter DeIterator[T]) gust.Result[any] {
		x := iter.AdvanceBy(n.(uint))
		if x.AsError() {
			return gust.Err[any](n.(uint) - x.Unwrap())
		}
		return gust.Ok[any](nil)
	}
	x := f.iterTryFold(n, advance)
	if x.IsErr() {
		remaining := x.ErrVal().(uint)
		if remaining > 0 {
			return gust.ToErrable(n - remaining)
		}
	}
	return gust.NonErrable[uint]()
}

func (f FlattenDeIterator[T, D]) realAdvanceBackBy(n uint) gust.Errable[uint] {
	var advance = func(n any, iter DeIterator[T]) gust.Result[any] {
		x := iter.AdvanceBackBy(n.(uint))
		if x.AsError() {
			return gust.Err[any](n.(uint) - x.Unwrap())
		}
		return gust.Ok[any](nil)
	}
	x := f.iterTryRfold(n, advance)
	if x.IsErr() {
		remaining := x.ErrVal().(uint)
		if remaining > 0 {
			return gust.ToErrable(n - remaining)
		}
	}
	return gust.NonErrable[uint]()
}

func (f FlattenDeIterator[T, D]) realCount() uint {
	var count = func(acc any, iter DeIterator[T]) any {
		return acc.(uint) + iter.Count()
	}
	return f.iterFold(uint(0), count).(uint)
}

func (f FlattenDeIterator[T, D]) realLast() gust.Option[T] {
	var last = func(last any, iter DeIterator[T]) any {
		return iter.Last().Or(last.(gust.Option[T]))
	}
	return f.iterFold(gust.None[T](), last).(gust.Option[T])
}

func (f FlattenDeIterator[T, D]) iterTryFold(acc any, fold func(any, DeIterator[T]) gust.Result[any]) gust.Result[any] {
	var flatten = func(frontiter *gust.Option[DeIterator[T]], fold func(any, DeIterator[T]) gust.Result[any]) func(any, D) gust.Result[any] {
		return func(acc any, iter D) gust.Result[any] {
			return fold(acc, *frontiter.Insert(FromDeIterable[T](iter)))
		}
	}
	if f.frontiter.IsSome() {
		x := fold(acc, f.frontiter.Unwrap())
		if x.IsErr() {
			return x
		}
		acc = x
		f.frontiter = gust.None[DeIterator[T]]()
	}

	acc = f.iter.TryFold(acc, flatten(&f.frontiter, fold))
	f.frontiter = gust.None[DeIterator[T]]()

	if f.backiter.IsSome() {
		x := fold(acc, f.backiter.Unwrap())
		if x.IsErr() {
			return x
		}
		acc = x
		f.backiter = gust.None[DeIterator[T]]()
	}
	return gust.Ok(acc)
}

func (f FlattenDeIterator[T, D]) iterFold(acc any, fold func(any, DeIterator[T]) any) any {
	f.frontiter.Inspect(func(i DeIterator[T]) {
		acc = fold(acc, i)
	})
	acc = f.iter.Fold(acc, func(acc any, i D) any {
		return fold(acc, FromDeIterable[T](i))
	})
	f.backiter.Inspect(func(i DeIterator[T]) {
		acc = fold(acc, i)
	})
	return acc
}

func (f FlattenDeIterator[T, D]) iterTryRfold(acc any, fold func(any, DeIterator[T]) gust.Result[any]) gust.Result[any] {
	var flatten = func(backiter *gust.Option[DeIterator[T]], fold func(any, DeIterator[T]) gust.Result[any]) func(any, D) gust.Result[any] {
		return func(acc any, iter D) gust.Result[any] {
			return fold(acc, *backiter.Insert(FromDeIterable[T](iter)))
		}
	}
	if f.backiter.IsSome() {
		x := fold(acc, f.backiter.Unwrap())
		if x.IsErr() {
			return x
		}
		acc = x
		f.backiter = gust.None[DeIterator[T]]()
	}

	acc = f.iter.TryRfold(acc, flatten(&f.backiter, fold))
	f.backiter = gust.None[DeIterator[T]]()

	if f.frontiter.IsSome() {
		x := fold(acc, f.frontiter.Unwrap())
		if x.IsErr() {
			return x
		}
		acc = x
		f.frontiter = gust.None[DeIterator[T]]()
	}
	return gust.Ok(acc)
}

func (f FlattenDeIterator[T, D]) iterRfold(acc any, fold func(any, DeIterator[T]) any) any {
	f.backiter.Inspect(func(i DeIterator[T]) {
		acc = fold(acc, i)
	})
	acc = f.iter.Rfold(acc, func(acc any, i D) any {
		return fold(acc, FromDeIterable[T](i))
	})
	f.frontiter.Inspect(func(i DeIterator[T]) {
		acc = fold(acc, i)
	})
	return acc
}
