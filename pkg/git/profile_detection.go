package git

import (
	"strings"

	"github.com/user-cube/gclone/pkg/config"
)

// DetectProfileForURL determines which profile to use based on the repository URL
func DetectProfileForURL(url string, profiles map[string]config.Profile) (string, bool) {
	// Normalize the URL to handle SSH format
	normalizedURL := NormalizeURL(url)

	// Check each profile for matching URL patterns
	for name, profile := range profiles {
		for _, pattern := range profile.URLPatterns {
			if strings.Contains(normalizedURL, pattern) {
				return name, true
			}
		}
	}

	return "", false
}

// NormalizeURL converts SSH URLs to a common format for pattern matching
func NormalizeURL(url string) string {
	// Handle SSH URL format (git@github.com:user/repo.git)
	if strings.HasPrefix(url, "git@") {
		// Convert git@github.com:user/repo.git to github.com:user/repo.git
		parts := strings.SplitN(url, "@", 2)
		if len(parts) == 2 {
			return parts[1]
		}
	}

	// Return as is if we can't normalize
	return url
}
