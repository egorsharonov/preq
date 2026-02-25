package converters

import "strings"

const pinPrefix = "pin"

func WithPinPrefix(id string) string {
	id = strings.TrimSpace(id)
	if id == "" || strings.HasPrefix(id, pinPrefix) {
		return id
	}

	return pinPrefix + id
}

func WithoutPinPrefix(id string) string {
	id = strings.TrimSpace(id)
	return strings.TrimPrefix(id, pinPrefix)
}
