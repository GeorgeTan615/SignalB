package utils

import "reflect"

func SliceContains[T any](slice []T, target T) bool {
	for _, e := range slice {
		if reflect.DeepEqual(e, target) {
			return true
		}
	}
	return false
}
