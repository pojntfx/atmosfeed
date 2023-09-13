package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "atmosfeed-client",
	Short: "Interact with the Atmosfeed server",
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
	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		return err
	}

	viper.AutomaticEnv()

	return rootCmd.Execute()
}
