package cmd

import (
	"log"
	"net"
	"os"
	"strconv"
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
	Long: `Create custom Bluesky feeds with WebAssembly and Scale.
Find more information at:
https://github.com/pojntfx/atmosfeed`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		viper.SetEnvPrefix("atmosfeed")
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}

		if v := os.Getenv("DATABASE_URL"); v != "" {
			log.Println("Using database address from DATABASE_URL env variable")

			viper.Set(postgresURLFlag, v)
		}

		if v := os.Getenv("REDIS_URL"); v != "" {
			log.Println("Using Redis address from REDIS_URL env variable")

			viper.Set(redisURLFlag, v)
		}

		if v := os.Getenv("S3_URL"); v != "" {
			log.Println("Using S3 address from S3_URL env variable")

			viper.Set(s3URLFlag, v)
		}

		if v := os.Getenv("PORT"); v != "" {
			log.Println("Using port from PORT env variable")

			la, err := net.ResolveTCPAddr("tcp", viper.GetString(laddrFlag))
			if err != nil {
				return err
			}

			p, err := strconv.Atoi(v)
			if err != nil {
				return err
			}

			la.Port = p
			viper.Set(laddrFlag, la.String())
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
