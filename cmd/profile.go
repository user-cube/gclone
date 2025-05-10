package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user-cube/gclone/pkg/config"
	"github.com/user-cube/gclone/pkg/ui"
)

// profileCmd represents the profile command
var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage profiles for gclone",
	Long:  `Manage profiles for gclone. You can list, add, edit, or remove profiles.`,
}

// profileListCmd represents the profile list command
var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	Long:  `List all profiles in the gclone configuration.`,
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

		ui.Section("Available profiles")

		for name, profile := range cfg.Profiles {
			ui.Info("Profile: %s", ui.Highlight(name))
			ui.PrintKeyValue("Name", profile.Name)
			ui.PrintKeyValue("SSH Host", profile.SSHHost)

			if len(profile.URLPatterns) > 0 {
				ui.Normal("  URL Patterns:\n")
				for _, pattern := range profile.URLPatterns {
					ui.Normal("    %s\n", pattern)
				}
			}

			if len(profile.GitConfigs) > 0 {
				ui.Normal("  Git Configs:\n")
				for key, value := range profile.GitConfigs {
					ui.Normal("    %s = %s\n", key, value)
				}
			}
			ui.Normal("\n")
		}
	},
}

// profileAddCmd represents the profile add command
var profileAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new profile",
	Long: `Add a new profile to the gclone configuration. 
If no name is provided, you'll be guided through an interactive profile creation process.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config")
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			ui.Error("Error loading configuration: %v", err)
			return
		}

		var profileName string

		// Get profile name interactively if not provided
		if len(args) == 0 {
			var err error
			profileName, err = ui.PromptInput("Profile name", "", func(input string) error {
				if input == "" {
					return fmt.Errorf("profile name cannot be empty")
				}
				return nil
			})
			if err != nil {
				ui.Error("Error getting profile name: %v", err)
				return
			}
		} else {
			profileName = args[0]
		}

		// Check if profile already exists
		if _, exists := cfg.Profiles[profileName]; exists {
			ui.Warning("Profile '%s' already exists", profileName)

			// Confirm overwrite
			confirmed, err := ui.Confirm("Do you want to overwrite it")
			if err != nil || !confirmed {
				ui.Warning("Operation cancelled")
				return
			}
		}

		// Get profile details interactively or from flags
		var name string
		nameFlag, _ := cmd.Flags().GetString("name")

		if nameFlag != "" {
			name = nameFlag
		} else if len(args) == 0 {
			// If in interactive mode and no name flag, prompt for display name
			var err error
			name, err = ui.PromptInput("Display name", profileName, nil)
			if err != nil {
				ui.Error("Error getting display name: %v", err)
				return
			}
		} else {
			name = profileName
		}

		// Get SSH host
		var sshHost string
		sshHostFlag, _ := cmd.Flags().GetString("ssh-host")

		if sshHostFlag != "" {
			sshHost = sshHostFlag
		} else {
			// Prompt for SSH host
			var err error
			sshHost, err = ui.PromptInput("SSH Host (e.g., github.com-personal)", "", func(input string) error {
				if input == "" {
					return fmt.Errorf("SSH host cannot be empty")
				}
				return nil
			})
			if err != nil {
				ui.Error("Error getting SSH host: %v", err)
				return
			}
		}

		// Create profile
		profile := config.Profile{
			Name:        name,
			SSHHost:     sshHost,
			GitConfigs:  make(map[string]string),
			URLPatterns: []string{},
		}

		// Get URL patterns
		urlPatterns, _ := cmd.Flags().GetStringArray("url-pattern")
		if len(urlPatterns) > 0 {
			profile.URLPatterns = urlPatterns
		} else if len(args) == 0 {
			// If in interactive mode, ask if user wants to add URL patterns
			addPatterns, err := ui.Confirm("Do you want to add URL patterns for automatic profile detection")
			if err == nil && addPatterns {
				for {
					pattern, err := ui.PromptInput("URL pattern (leave empty to stop)", "", nil)
					if err != nil || pattern == "" {
						break
					}
					profile.URLPatterns = append(profile.URLPatterns, pattern)
				}
			}
		}

		// Get Git config values
		gitUsername, _ := cmd.Flags().GetString("git-username")
		gitEmail, _ := cmd.Flags().GetString("git-email")

		if gitUsername != "" {
			profile.GitConfigs["user.name"] = gitUsername
		} else if len(args) == 0 {
			// If in interactive mode, prompt for Git username
			username, err := ui.PromptInput("Git username (leave empty to skip)", "", nil)
			if err == nil && username != "" {
				profile.GitConfigs["user.name"] = username
			}
		}

		if gitEmail != "" {
			profile.GitConfigs["user.email"] = gitEmail
		} else if len(args) == 0 {
			// If in interactive mode, prompt for Git email
			email, err := ui.PromptInput("Git email (leave empty to skip)", "", nil)
			if err == nil && email != "" {
				profile.GitConfigs["user.email"] = email
			}
		}

		// If in interactive mode, ask if user wants to add additional Git configurations
		if len(args) == 0 {
			addGitConfigs, err := ui.Confirm("Do you want to add additional Git configurations")
			if err == nil && addGitConfigs {
				for {
					key, err := ui.PromptInput("Git config key (e.g., commit.gpgsign, leave empty to stop)", "", nil)
					if err != nil || key == "" {
						break
					}

					value, err := ui.PromptInput("Value for "+key, "", nil)
					if err != nil {
						ui.Warning("Error getting config value, skipping")
						continue
					}

					profile.GitConfigs[key] = value
				}
			}
		}

		// Save profile
		cfg.Profiles[profileName] = profile
		if err := config.SaveConfig(cfg, configFile); err != nil {
			ui.Error("Error saving configuration: %v", err)
			return
		}

		ui.Success("Profile '%s' added successfully", profileName)
	},
}

// profileRemoveCmd represents the profile remove command
var profileRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a profile",
	Long:  `Remove a profile from the gclone configuration.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profileName := args[0]

		configFile, _ := cmd.Flags().GetString("config")
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			ui.Error("Error loading configuration: %v", err)
			return
		}

		// Check if profile exists
		if _, exists := cfg.Profiles[profileName]; !exists {
			ui.Error("Profile '%s' does not exist", profileName)
			return
		}

		// Confirm removal
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			confirmed, err := ui.Confirm("Are you sure you want to remove profile '" + profileName + "'")
			if err != nil || !confirmed {
				ui.Warning("Operation cancelled")
				return
			}
		}

		// Remove profile
		delete(cfg.Profiles, profileName)
		if err := config.SaveConfig(cfg, configFile); err != nil {
			ui.Error("Error saving configuration: %v", err)
			return
		}

		ui.Success("Profile '%s' removed successfully", profileName)
	},
}

func init() {
	rootCmd.AddCommand(profileCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileAddCmd)
	profileCmd.AddCommand(profileRemoveCmd)

	// Global flags for profile commands
	profileCmd.PersistentFlags().StringP("config", "c", "", "Path to config file (default is $HOME/.gclone/config.yml)")

	// Flags for profile add command
	profileAddCmd.Flags().StringP("name", "n", "", "Display name for the profile")
	profileAddCmd.Flags().StringP("ssh-host", "s", "", "SSH host to use for this profile (e.g., github.com-personal)")
	profileAddCmd.Flags().StringP("git-username", "u", "", "Git username to configure for this profile")
	profileAddCmd.Flags().StringP("git-email", "e", "", "Git email to configure for this profile")
	profileAddCmd.Flags().StringArrayP("url-pattern", "p", []string{}, "URL patterns to automatically match this profile (can be specified multiple times)")

	// Flags for profile remove command
	profileRemoveCmd.Flags().BoolP("force", "f", false, "Force removal without confirmation")
}
