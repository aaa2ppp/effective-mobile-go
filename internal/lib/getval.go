package lib

import "errors"

type (
	getValueFunc     func(key string) string
	parseFunc[T any] func(s string) (T, error)
	checkFunc[T any] func(v T) error
)

func GetValue[T any](get getValueFunc, parse parseFunc[T], check checkFunc[T], key string, req bool, value T) (T, error) {
	var zero T

	s := get(key)

	if req && s == "" {
		return zero, errors.New(key + " is required")
	}

	if s == "" {
		return value, nil
	}

	v, err := parse(s)

	if err != nil {
		return zero, err
	}

	return v, check(v)
}
