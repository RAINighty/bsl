package collector

import "strings"

const StreetlightPrefix = "#路灯"

func DetectStreetlight(content string) (bool, string) {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, StreetlightPrefix) {
		return false, ""
	}
	note := strings.TrimSpace(strings.TrimPrefix(trimmed, StreetlightPrefix))
	if note == "" {
		note = trimmed
	}
	return true, note
}
