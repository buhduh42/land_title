package util

import "golang.org/x/exp/constraints"

func Ptr[T constraints.Ordered](val T) *T {
	return &val
}
