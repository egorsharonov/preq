package converters

import "strings"

func ToPtr[T any](val T) *T {
	return &val
}

func StrToPtr(val string) *string {
	if strings.TrimSpace(val) == "" {
		return nil
	}

	return &val
}

func PtrToStr(val *string) string {
	if val == nil {
		return ""
	}

	return *val
}
