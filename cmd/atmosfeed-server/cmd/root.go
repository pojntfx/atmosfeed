package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	redisURLFlag    = "redis-url"
	s3URLFlag       = "s3-url"
	verboseFlag     = "verbose"
	postgresURLFlag = "postgres-url"
)

var rootCmd = &cobra.Command{
	Use:   "atmosfeed-server",
	Short: "Start Atmosfeed managers and workers",
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
	rootCmd.PersistentFlags().String(postgresURLFlag, "postgresql://postgres@localhost:5432/atmosfeed?sslmode=disable", "PostgreSQL URL")
	rootCmd.PersistentFlags().String(redisURLFlag, "redis://localhost:6379/0", "Redis URL")
	rootCmd.PersistentFlags().String(s3URLFlag, "http://minioadmin:minioadmin@localhost:9000?bucket=atmosfeed", "S3 URL")
	rootCmd.PersistentFlags().Bool(verboseFlag, false, "Whether to enable verbose logging")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		return err
	}

	viper.AutomaticEnv()

	return rootCmd.Execute()
}
