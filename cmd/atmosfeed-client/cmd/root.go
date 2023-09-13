package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	atmosfeedURLFlag = "atmosfeed-url"

	pdsURLFlag   = "pds-url"
	usernameFlag = "username"
	passwordFlag = "password"
)

var rootCmd = &cobra.Command{
	Use:   "atmosfeed-client",
	Short: "Interact with feeds on Atmosfeed servers and Bluesky PDSes",
	Long: `Create fully custom Bluesky feeds with Wasm modules, powered by Scale Functions.
Find more information at:
https://github.com/pojntfx/atmosfeed`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		viper.SetEnvPrefix("atmosfeed")
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}

		return nil
	},
}

func Execute() error {
	rootCmd.PersistentFlags().String(atmosfeedURLFlag, "http://localhost:1337", "Atmosfeed server URL")

	rootCmd.PersistentFlags().String(pdsURLFlag, "https://bsky.social", "PDS URL")
	rootCmd.PersistentFlags().String(usernameFlag, "example.bsky.social", "Bluesky username (ignored if `--publish` is not provided)")
	rootCmd.PersistentFlags().String(passwordFlag, "", "Bluesky password, preferably an app password (get one from https://bsky.app/settings/app-passwords, ignored if `--publish` is not provided)")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		return err
	}

	viper.AutomaticEnv()

	return rootCmd.Execute()
}
