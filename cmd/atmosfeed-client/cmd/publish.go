package cmd

import (
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	feedGeneratorDIDFlag = "feed-generator-did"
	feedNameFlag         = "feed-name"
	feedDescriptionFlag  = "feed-description"
)

const (
	lexiconFeedGenerator = "app.bsky.feed.generator"
)

var publishCmd = &cobra.Command{
	Use:     "publish",
	Aliases: []string{"p"},
	Short:   "Publish a feed to a Bluesky PDS",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}

		client, auth, err := authorize(cmd.Context())
		if err != nil {
			return err
		}

		feedRkey := viper.GetString(feedRkeyFlag)
		feedDescription := viper.GetString(feedDescriptionFlag)

		rec := &util.LexiconTypeDecoder{
			Val: &bsky.FeedGenerator{
				CreatedAt:   time.Now().Format(time.RFC3339),
				Description: &feedDescription,
				Did:         viper.GetString(feedGeneratorDIDFlag),
				DisplayName: viper.GetString(feedNameFlag),
			},
		}

		ex, err := atproto.RepoGetRecord(cmd.Context(), client, "", lexiconFeedGenerator, auth.Did, feedRkey)
		if err == nil {
			if _, err := atproto.RepoPutRecord(cmd.Context(), client, &atproto.RepoPutRecord_Input{
				Collection: lexiconFeedGenerator,
				Repo:       auth.Did,
				Rkey:       feedRkey,
				Record:     rec,
				SwapRecord: ex.Cid,
			}); err != nil {
				return err
			}
		} else {
			if _, err := atproto.RepoCreateRecord(cmd.Context(), client, &atproto.RepoCreateRecord_Input{
				Collection: lexiconFeedGenerator,
				Repo:       auth.Did,
				Rkey:       &feedRkey,
				Record:     rec,
			}); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	publishCmd.PersistentFlags().String(feedRkeyFlag, "trending", "Machine-readable key for the feed")
	publishCmd.PersistentFlags().String(feedNameFlag, "Atmosfeed Trending", "Human-readable name for the feed (ignored for `--delete`)")
	publishCmd.PersistentFlags().String(feedDescriptionFlag, "An example trending feed for Atmosfeed", "Description for the feed (ignored for `--delete`)")

	publishCmd.PersistentFlags().String(feedGeneratorDIDFlag, "did:web:atmosfeed-feeds.serveo.net", "DID of the feed generator (typically the hostname of the publicly reachable URL) (ignored for `--delete`)")

	viper.AutomaticEnv()

	rootCmd.AddCommand(publishCmd)
}
