package cmd

import (
	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var unpublishCmd = &cobra.Command{
	Use:     "unpublish",
	Aliases: []string{"u"},
	Short:   "Unpublish a feed from a Bluesky PDS",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}

		client, auth, err := authorize(cmd.Context())
		if err != nil {
			return err
		}

		feedRkey := viper.GetString(feedRkeyFlag)

		if err := atproto.RepoDeleteRecord(cmd.Context(), client, &atproto.RepoDeleteRecord_Input{
			Collection: lexiconFeedGenerator,
			Repo:       auth.Did,
			Rkey:       feedRkey,
		}); err != nil {
			panic(err)
		}

		return nil
	},
}

func init() {
	unpublishCmd.PersistentFlags().String(feedRkeyFlag, "trending", "Machine-readable key for the feed")

	viper.AutomaticEnv()

	rootCmd.AddCommand(unpublishCmd)
}
