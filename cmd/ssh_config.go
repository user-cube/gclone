// Package cmd contains commands for gclone
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user-cube/gclone/pkg/config"
	"github.com/user-cube/gclone/pkg/ui"
)

// sshConfigCmd represents the ssh-config command
var sshConfigCmd = &cobra.Command{
	Use:   "ssh-config [profile]",
	Short: "Generate or update SSH config for a profile",
	Long: `Generate or update SSH configuration for a gclone profile.
This command helps you set up the necessary SSH configuration in ~/.gclone/ssh_config
that matches your gclone profile settings, and adds an Include directive to your 
~/.ssh/config file if needed.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config")
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			ui.Error("Error loading configuration: %v", err)
			return
		}

		if len(cfg.Profiles) == 0 {
			ui.Warning("No profiles found. Run 'gclone init' to create default profiles.")
			return
		}

		var profileName string
		if len(args) > 0 {
			profileName = args[0]
		} else if len(cfg.Profiles) == 1 {
			// If only one profile exists, use it
			for name := range cfg.Profiles {
				profileName = name
			}
		} else {
			// If multiple profiles exist, prompt for selection
			profileNames := make([]string, 0, len(cfg.Profiles))
			for name := range cfg.Profiles {
				profileNames = append(profileNames, name)
			}

			selectedProfile, err := ui.SelectFromList("Select profile to configure SSH for", profileNames)
			if err != nil {
				ui.Error("Error selecting profile: %v", err)
				return
			}
			profileName = selectedProfile
		}

		// Check if profile exists
		profile, ok := cfg.Profiles[profileName]
		if !ok {
			ui.Error("Profile '%s' not found", profileName)
			return
		}

		// Generate SSH config
		sshHost := profile.SSHHost
		if sshHost == "" {
			ui.Error("Profile '%s' does not have an SSH host configured", profileName)
			return
		}

		identityFile, _ := cmd.Flags().GetString("identity-file")
		if identityFile == "" {
			// Suggest a default identity file name
			defaultIdentityFile := fmt.Sprintf("~/.ssh/%s", strings.Replace(sshHost, "github.com-", "github_", 1))
			var err error
			identityFile, err = ui.PromptInput("SSH identity file path", defaultIdentityFile, nil)
			if err != nil {
				ui.Error("Error getting identity file: %v", err)
				return
			}
		}

		sshConfig := fmt.Sprintf(`# %s profile (added by gclone)
Host %s
  Hostname github.com
  AddKeysToAgent yes
  UseKeychain yes
  IdentityFile %s
`, profile.Name, sshHost, identityFile)

		// Get SSH config path
		homeDir, err := os.UserHomeDir()
		if err != nil {
			ui.Error("Error getting home directory: %v", err)
			return
		}

		// Create the gclone SSH config file
		gcloneDir := filepath.Join(homeDir, ".gclone")
		gcloneSshConfigPath := filepath.Join(gcloneDir, "ssh_config")
		mainSshConfigPath := filepath.Join(homeDir, ".ssh", "config")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			ui.Section("SSH Configuration (Dry Run)")
			ui.Info("The following configuration would be added to %s:", gcloneSshConfigPath)
			fmt.Println()
			ui.Normal("%s\n", sshConfig)

			includeDirective := "Include ~/.gclone/ssh_config"
			ui.Info("And the following include directive would be added to %s (if not already present):", mainSshConfigPath)
			fmt.Println()
			ui.Normal("%s\n", includeDirective)
			return
		}

		// Ensure gclone directory exists
		if err := os.MkdirAll(gcloneDir, 0755); err != nil {
			ui.Error("Error creating gclone directory: %v", err)
			return
		}

		// Check if gclone SSH config file exists
		gcloneSshConfigExists := false
		var existingContent string

		if _, err := os.Stat(gcloneSshConfigPath); err == nil {
			gcloneSshConfigExists = true

			// Read existing gclone SSH config
			content, err := os.ReadFile(gcloneSshConfigPath)
			if err != nil {
				ui.Error("Error reading gclone SSH config: %v", err)
				return
			}
			existingContent = string(content)

			// Check if host already exists
			hostPattern := fmt.Sprintf("Host %s", sshHost)
			if strings.Contains(existingContent, hostPattern) {
				ui.Warning("SSH configuration for '%s' already exists in %s", sshHost, gcloneSshConfigPath)

				confirmed, err := ui.Confirm("Do you want to update it")
				if err != nil || !confirmed {
					ui.Warning("Operation cancelled")
					return
				}

				// Remove existing configuration for this host
				lines := strings.Split(existingContent, "\n")
				newLines := []string{}
				skip := false

				for _, line := range lines {
					trimmedLine := strings.TrimSpace(line)
					if strings.HasPrefix(trimmedLine, hostPattern) {
						skip = true
						continue
					} else if skip && strings.HasPrefix(trimmedLine, "Host ") {
						skip = false
					}

					if !skip {
						newLines = append(newLines, line)
					}
				}

				// Update the content without the host
				existingContent = strings.Join(newLines, "\n")
			}
		}

		// Write or update the gclone SSH config file
		var newContent string
		if gcloneSshConfigExists {
			// Add a newline if the file doesn't end with one
			if existingContent != "" && !strings.HasSuffix(existingContent, "\n") {
				newContent = existingContent + "\n" + sshConfig
			} else {
				newContent = existingContent + sshConfig
			}
		} else {
			newContent = sshConfig
		}

		// Write the gclone SSH config file
		if err := os.WriteFile(gcloneSshConfigPath, []byte(newContent), 0644); err != nil {
			ui.Error("Error writing gclone SSH config: %v", err)
			return
		}

		// Now ensure the main SSH config includes our gclone SSH config
		ensureIncludeDirective(mainSshConfigPath, "~/.gclone/ssh_config")

		action := "created"
		if gcloneSshConfigExists {
			action = "updated"
		}

		ui.Success("SSH configuration %s at %s", action, gcloneSshConfigPath)
		ui.Info("Configuration for '%s' added:", sshHost)
		ui.Normal("%s\n", sshConfig)

		// Remind about creating the SSH key if it doesn't exist
		identityFileExpanded := strings.Replace(identityFile, "~/", homeDir+"/", 1)
		if _, err := os.Stat(identityFileExpanded); os.IsNotExist(err) {
			ui.Warning("SSH key %s does not exist yet", identityFile)
			ui.Info("You can create it with:")
			ui.Normal("  ssh-keygen -t ed25519 -f %s -C \"your_email@example.com\"\n", identityFile)
		}
	},
}

// ensureIncludeDirective ensures that the specified SSH config file includes the given path
func ensureIncludeDirective(sshConfigPath, includePath string) {
	// Ensure SSH directory exists
	sshDir := filepath.Dir(sshConfigPath)
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		ui.Error("Error creating SSH directory: %v", err)
		return
	}

	includeDirective := fmt.Sprintf("Include %s", includePath)

	// Check if SSH config file exists
	if _, err := os.Stat(sshConfigPath); err == nil {
		// Read existing SSH config
		content, err := os.ReadFile(sshConfigPath)
		if err != nil {
			ui.Error("Error reading SSH config: %v", err)
			return
		}

		// Check if include directive already exists
		if strings.Contains(string(content), includeDirective) {
			// Include directive already exists, nothing to do
			return
		}

		// Add include directive at the top of the file with other includes
		lines := strings.Split(string(content), "\n")
		newLines := []string{}
		includeAdded := false

		for i, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			// Add include directive after any existing include lines
			// but before the first non-include, non-comment line
			if strings.HasPrefix(trimmedLine, "Include ") {
				newLines = append(newLines, line)
				// If this is the last include line, add our include directive
				if i+1 < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i+1]), "Include ") {
					newLines = append(newLines, includeDirective)
					includeAdded = true
				}
			} else if !includeAdded && i > 0 && !strings.HasPrefix(trimmedLine, "#") && trimmedLine != "" {
				// First non-include, non-comment line and we haven't added our include yet
				newLines = append(newLines, includeDirective)
				newLines = append(newLines, line)
				includeAdded = true
			} else {
				newLines = append(newLines, line)
			}
		}

		// If we still haven't added the include directive (e.g., file had no includes),
		// add it at the beginning
		if !includeAdded {
			newLines = append([]string{includeDirective}, newLines...)
		}

		// Write back the updated content
		if err := os.WriteFile(sshConfigPath, []byte(strings.Join(newLines, "\n")), 0644); err != nil {
			ui.Error("Error updating SSH config: %v", err)
			return
		}

		ui.Success("Added include directive to %s", sshConfigPath)
	} else {
		// Create new SSH config file with include directive
		if err := os.WriteFile(sshConfigPath, []byte(includeDirective+"\n"), 0644); err != nil {
			ui.Error("Error creating SSH config: %v", err)
			return
		}

		ui.Success("Created SSH config file at %s with include directive", sshConfigPath)
	}
}

func init() {
	rootCmd.AddCommand(sshConfigCmd)
	sshConfigCmd.Flags().StringP("config", "c", "", "Path to config file (default is $HOME/.gclone/config.yml)")
	sshConfigCmd.Flags().StringP("identity-file", "i", "", "Path to SSH identity file (default is ~/.ssh/github_<profile>)")
	sshConfigCmd.Flags().BoolP("dry-run", "d", false, "Print the configuration without writing to file")
}
