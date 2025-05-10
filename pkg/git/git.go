package git

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/user-cube/gclone/pkg/config"
	"github.com/user-cube/gclone/pkg/ui"
)

// TransformGitURL transforms a git URL to use the specified SSH host
func TransformGitURL(url string, profile *config.Profile) (string, error) {
	if profile == nil || profile.SSHHost == "" {
		return url, nil
	}

	// Handle SSH URL format (git@github.com:user/repo.git)
	sshRegex := regexp.MustCompile(`^git@([^:]+):(.+)$`)
	if matches := sshRegex.FindStringSubmatch(url); len(matches) == 3 {
		// We don't need to use originalHost, just get the path part
		path := matches[2]

		// Replace the host with the profile's SSH host
		return fmt.Sprintf("git@%s:%s", profile.SSHHost, path), nil
	}

	return url, fmt.Errorf("unsupported git URL format: %s (only SSH URLs are supported)", url)
}

// CloneRepository clones a repository using the specified profile
func CloneRepository(url, destination string, profile *config.Profile, extraArgs []string) error {
	if profile != nil {
		var err error
		url, err = TransformGitURL(url, profile)
		if err != nil {
			return err
		}
	}

	// Prepare the git clone command
	args := []string{"clone", url}

	// Add destination if provided
	if destination != "" {
		args = append(args, destination)
	} else {
		// Extract repo name from URL for better error messages
		parts := strings.Split(url, "/")
		if len(parts) > 0 {
			repoName := strings.TrimSuffix(parts[len(parts)-1], ".git")
			destination = repoName
		}
	}

	// Add any extra arguments
	if len(extraArgs) > 0 {
		args = append(args, extraArgs...)
	}

	// Execute the git clone command
	ui.Info("Running git %s", strings.Join(args, " "))
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	// Apply Git configurations if a profile is specified
	if profile != nil && len(profile.GitConfigs) > 0 {
		repoPath := destination
		if repoPath == "" {
			// Extract repository name from URL
			parts := strings.Split(url, "/")
			repoName := strings.TrimSuffix(parts[len(parts)-1], ".git")
			repoPath = repoName
		}

		// Apply Git configurations
		if err := ApplyGitConfigs(repoPath, profile.GitConfigs); err != nil {
			return fmt.Errorf("failed to apply git configs: %w", err)
		}
	}

	return nil
}

// ApplyGitConfigs applies Git configurations to a repository
func ApplyGitConfigs(repoPath string, configs map[string]string) error {
	// Ensure the path exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("repository path does not exist: %s", repoPath)
	}

	// Apply each configuration
	for key, value := range configs {
		ui.Info("Setting git config %s=%s", key, value)
		cmd := exec.Command("git", "config", "--local", key, value)
		cmd.Dir = repoPath

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set git config %s=%s: %w", key, value, err)
		}
	}

	return nil
}

// GetRepositoryName extracts the repository name from a Git URL
func GetRepositoryName(url string) string {
	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	// For SSH URLs like git@github.com:user/repo
	if strings.Contains(url, ":") {
		parts := strings.Split(url, ":")
		if len(parts) > 1 {
			subParts := strings.Split(parts[1], "/")
			if len(subParts) > 0 {
				return subParts[len(subParts)-1]
			}
		}
	}

	return ""
}
