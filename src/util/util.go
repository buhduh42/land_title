package util

import "golang.org/x/exp/constraints"

type pType interface {
	constraints.Ordered | ~bool
}

func Ptr[T pType](val T) *T {
	return &val
}
