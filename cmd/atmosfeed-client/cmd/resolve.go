package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	handleFlag = "handle"
)

var (
	errMissingHandle = errors.New("missing handle")
)

var resolveCmd = &cobra.Command{
	Use:     "resolve",
	Aliases: []string{"p"},
	Short:   "Resolve a handle to a DID",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}

		client, _, err := authorize(cmd.Context())
		if err != nil {
			return err
		}

		if strings.TrimSpace(viper.GetString(handleFlag)) == "" {
			return errMissingHandle
		}

		h, err := atproto.IdentityResolveHandle(cmd.Context(), client, viper.GetString(handleFlag))
		if err != nil {
			return err
		}

		fmt.Println(h.Did)

		return nil
	},
}

func init() {
	resolveCmd.PersistentFlags().String(handleFlag, "", "Handle/username/domain to resolve")

	viper.AutomaticEnv()

	rootCmd.AddCommand(resolveCmd)
}
