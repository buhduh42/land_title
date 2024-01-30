package util

import (
	"os"
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
)

type pType interface {
	constraints.Ordered | ~bool
}

func Ptr[T pType](val T) *T {
	return &val
}

// following my normal pattern, this
// will take a slice of the usual testdata struct
// and test that subset
// Start=0, Stop=-1 is to the end
// Rest would be a vararg for sub tests
// test specific
type TestMetaData struct {
	Start int
	Stop  int
	Rest  []int
}

func GetTestMetaData() (*TestMetaData, error) {
	toRet := &TestMetaData{
		Start: -1,
		Stop:  -1,
	}
	indexIndecesStr := os.Getenv("TEST_INDECES")
	if indexIndecesStr == "" {
		return toRet, nil
	}
	components := strings.SplitN(indexIndecesStr, ":", 3)
	var err error
	if len(components) > 0 {
		if toRet.Start, err = strconv.Atoi(components[0]); err != nil {
			return nil, err
		}
	}
	if len(components) > 1 {
		if toRet.Stop, err = strconv.Atoi(components[1]); err != nil {
			return nil, err
		}
	}
	if len(components) == 3 {
		rest := strings.Split(components[3], ":")
		tmp := make([]int, len(rest))
		for i, r := range rest {
			if tmp[i], err = strconv.Atoi(r); err != nil {
				return nil, err
			}
		}
		toRet.Rest = tmp
	}
	return toRet, nil
}
