package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/user-cube/gclone/pkg/ui"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gclone",
	Short: "A tool to manage Git repository clones with different profiles",
	Long: `GClone is a command-line tool that helps you manage Git repository 
clones with multiple profiles. It allows you to define different SSH hosts 
and Git configurations for different Git accounts (e.g., personal, work), 
and automatically applies the appropriate settings when cloning repositories.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		ui.Error("Error: %v", err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gclone/config.yml)")
}
