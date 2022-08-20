package iter

import (
	"github.com/andeya/gust"
)

type ChanNext[T any] struct {
	c <-chan T
}

func NewChanNext[T any](c <-chan T) ChanNext[T] {
	return ChanNext[T]{c: c}
}

func (c ChanNext[T]) ToIter() *AnyIter[T] {
	return IterAny[T](c)
}

func (c ChanNext[T]) Next() gust.Option[T] {
	var x, ok = <-c.c
	if ok {
		return gust.Some(x)
	}
	return gust.None[T]()
}
