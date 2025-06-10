package utils

import (
	"regexp"
)

func ExtractPlaylistID(url string) string {
	patterns := []string{
		`[?&]list=([^&]+)`,
		`/playlist\?list=([^&]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(url); len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}
