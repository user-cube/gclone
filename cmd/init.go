package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user-cube/gclone/pkg/config"
	"github.com/user-cube/gclone/pkg/ui"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the gclone configuration",
	Long: `Initialize the gclone configuration file with default settings.
This will create a configuration file at ~/.gclone/config.yml with sample profiles.`,
	Run: func(cmd *cobra.Command, args []string) {
		configFile := config.DefaultConfigFile()
		configDir := config.DefaultConfigDir()

		// Check if config file already exists
		if _, err := os.Stat(configFile); err == nil {
			overwrite, _ := cmd.Flags().GetBool("force")
			if !overwrite {
				ui.Warning("Configuration file already exists at %s", configFile)
				ui.Warning("Use --force to overwrite the existing configuration")
				return
			}
		}

		// Ensure directory exists
		if err := os.MkdirAll(configDir, 0755); err != nil {
			ui.Error("Error creating config directory: %v", err)
			return
		}

		// Get default config
		cfg := config.GetDefaultConfig()

		// Save config
		if err := config.SaveConfig(cfg, configFile); err != nil {
			ui.Error("Error saving configuration: %v", err)
			return
		}

		ui.Success("Configuration initialized successfully at %s", configFile)
		ui.Info("Default profiles created:")
		for name, profile := range cfg.Profiles {
			ui.Normal("  - %s (SSH Host: %s)\n", ui.Highlight(name), profile.SSHHost)

			if len(profile.URLPatterns) > 0 {
				ui.Normal("    URL Patterns: %s\n", strings.Join(profile.URLPatterns, ", "))
			}

			ui.Normal("    Git Config: user.name=%s, user.email=%s\n",
				profile.GitConfigs["user.name"],
				profile.GitConfigs["user.email"])
		}

		ui.Section("Examples")
		ui.Info("You can now clone repositories with a specific profile:")
		ui.Normal("  gclone clone git@github.com:user/repo.git --profile=personal\n")

		ui.Info("Or let gclone automatically detect the appropriate profile based on URL patterns:")
		ui.Normal("  gclone clone git@github.com:your-personal-username/repo.git\n")
		ui.Info("  (Profile will be automatically detected based on configured URL patterns)")

		ui.Warning("Note: Only SSH URL format (git@github.com:user/repo.git) is supported")

		ui.Section("SSH Configuration")
		ui.Info("Important: Make sure to set up your SSH configuration in ~/.ssh/config")
		ui.Info("  For example, for the 'personal' profile with ssh_host 'github.com-personal':")

		sshConfigExample := `Host github.com-personal
  Hostname github.com
  AddKeysToAgent yes
  UseKeychain yes
  IdentityFile ~/.ssh/github_personal`

		ui.Normal("\n%s\n", sshConfigExample)
		ui.Warning("  The Host value must match the ssh_host in your GClone profile")

		ui.Section("SSH Setup Helper")
		ui.Info("Alternatively, you can use the ssh-config command to set up your SSH configuration:")
		ui.Normal("  gclone ssh-config personal\n")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolP("force", "f", false, "Force overwrite existing configuration")
}
