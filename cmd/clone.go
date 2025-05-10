package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user-cube/gclone/pkg/config"
	"github.com/user-cube/gclone/pkg/git"
	"github.com/user-cube/gclone/pkg/ui"
)

// cloneCmd represents the clone command
var cloneCmd = &cobra.Command{
	Use:   "clone [url] [destination]",
	Short: "Clone a git repository with a specific profile",
	Long: `Clone a git repository with a specific profile.
This will transform the repository URL to use the specified SSH host,
and apply any Git configurations specified in the profile.
Only SSH URL format (git@github.com:user/repo.git) is supported.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Load config
		configFile, _ := cmd.Flags().GetString("config")
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			ui.Error("Error loading configuration: %v", err)
			return
		}

		if len(cfg.Profiles) == 0 {
			ui.Warning("No profiles found in configuration. Run 'gclone init' to create default profiles.")
			return
		}

		// Get URL and destination
		url := args[0]
		var destination string
		if len(args) > 1 {
			destination = args[1]
		}

		// Get profile
		profileName, _ := cmd.Flags().GetString("profile")

		// If no profile specified, try to detect it from the URL
		if profileName == "" {
			detectedProfile, found := git.DetectProfileForURL(url, cfg.Profiles)
			if found {
				profileName = detectedProfile
				ui.Info("Automatically detected profile: %s", ui.Highlight(profileName))
			} else {
				// If no profile detected, prompt user to select one
				profileNames := make([]string, 0, len(cfg.Profiles))
				for name := range cfg.Profiles {
					profileNames = append(profileNames, name)
				}

				selectedProfile, err := ui.SelectFromList("Select profile", profileNames)
				if err != nil {
					ui.Error("Prompt failed: %v", err)
					return
				}

				profileName = selectedProfile
			}
		}

		// Check if profile exists
		profile, ok := cfg.Profiles[profileName]
		if !ok {
			ui.Error("Profile '%s' not found", profileName)
			return
		}

		// Collect extra git args
		var extraArgs []string

		// Check for depth flag
		depth, _ := cmd.Flags().GetInt("depth")
		if depth > 0 {
			extraArgs = append(extraArgs, fmt.Sprintf("--depth=%d", depth))
		}

		// Check for branch flag
		branch, _ := cmd.Flags().GetString("branch")
		if branch != "" {
			extraArgs = append(extraArgs, fmt.Sprintf("--branch=%s", branch))
		}

		// Pass through any additional flags after --
		afterDoubleHyphen, found := findArgsAfterDoubleHyphen(os.Args)
		if found {
			extraArgs = append(extraArgs, afterDoubleHyphen...)
		}

		// Display information about the clone operation
		transformedURL, _ := git.TransformGitURL(url, &profile)

		details := map[string]string{
			"Original URL":    url,
			"Transformed URL": transformedURL,
		}

		ui.OperationInfo("Cloning", profileName, details)

		if len(profile.GitConfigs) > 0 {
			ui.Info("Git configs to apply:")
			for key, value := range profile.GitConfigs {
				ui.PrintKeyValue(key, value)
			}
			ui.Normal("\n")
		}

		// Clone the repository
		err = git.CloneRepository(url, destination, &profile, extraArgs)
		if err != nil {
			ui.OperationError("cloning repository", err)
			return
		}

		repoName := destination
		if repoName == "" {
			repoName = git.GetRepositoryName(url)
		}

		ui.OperationSuccess("Repository cloned successfully: " + repoName)
		if len(profile.GitConfigs) > 0 {
			ui.Success("Git configurations applied successfully")
		}
	},
}

// findArgsAfterDoubleHyphen finds arguments after a -- separator
func findArgsAfterDoubleHyphen(args []string) ([]string, bool) {
	for i, arg := range args {
		if arg == "--" && i < len(args)-1 {
			return args[i+1:], true
		}
	}
	return nil, false
}

func init() {
	rootCmd.AddCommand(cloneCmd)

	cloneCmd.Flags().StringP("profile", "p", "", "Profile to use for cloning")
	cloneCmd.Flags().StringP("config", "c", "", "Path to config file (default is $HOME/.gclone/config.yml)")
	cloneCmd.Flags().IntP("depth", "d", 0, "Create a shallow clone with the specified depth")
	cloneCmd.Flags().StringP("branch", "b", "", "Clone the specified branch instead of the remote's HEAD")
}
