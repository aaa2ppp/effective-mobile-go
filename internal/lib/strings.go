package lib

import "unsafe"

func StringAsBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func UnsafeString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
