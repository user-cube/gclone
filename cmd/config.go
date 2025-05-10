package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/user-cube/gclone/pkg/config"
	"github.com/user-cube/gclone/pkg/ui"
	"gopkg.in/yaml.v3"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display or edit the gclone configuration",
	Long:  `Display or edit the gclone configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config")
		if configFile == "" {
			configFile = config.DefaultConfigFile()
		}

		// Check if config file exists
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			ui.Warning("Configuration file does not exist at %s", configFile)
			ui.Warning("Run 'gclone init' to create a default configuration")
			return
		}

		// Load config
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			ui.Error("Error loading configuration: %v", err)
			return
		}

		// Get format
		format, _ := cmd.Flags().GetString("format")

		switch format {
		case "yaml":
			// Marshal config to YAML
			data, err := yaml.Marshal(cfg)
			if err != nil {
				ui.Error("Error encoding configuration: %v", err)
				return
			}
			ui.Normal("%s\n", string(data))
		default:
			// Display config in a user-friendly format
			ui.Info("Configuration file: %s", configFile)
			ui.Normal("\n")

			if len(cfg.Profiles) == 0 {
				ui.Warning("No profiles found")
				return
			}

			ui.Info("Profiles:")
			ui.Normal("\n")

			for name, profile := range cfg.Profiles {
				ui.Section("Profile: " + ui.Highlight(name))
				ui.Normal("  Name: %s\n", profile.Name)
				ui.Normal("  SSH Host: %s\n", profile.SSHHost)

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
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.Flags().StringP("config", "c", "", "Path to config file (default is $HOME/.gclone/config.yml)")
	configCmd.Flags().StringP("format", "f", "pretty", "Output format (pretty, yaml)")
}
