package cmd

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	feedRkeyFlag       = "feed-rkey"
	feedClassifierFlag = "feed-classifier"
	feedPinnedDIDFlag  = "pinned-feed-did"
	feedPinnedRkeyFlag = "pinned-feed-rkey"
	clearPinnedFlag    = "clear-pinned"
)

var applyCmd = &cobra.Command{
	Use:     "apply",
	Aliases: []string{"a"},
	Short:   "Create or update a feed on an Atmosfeed server",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}

		_, auth, err := authorize(cmd.Context())
		if err != nil {
			return err
		}

		u, err := url.Parse(viper.GetString(atmosfeedURLFlag))
		if err != nil {
			return err
		}

		f, err := os.Open(viper.GetString(feedClassifierFlag))
		if err != nil {
			return err
		}
		defer f.Close()

		{
			u := u.JoinPath("admin", "feeds")

			q := u.Query()
			q.Add("rkey", viper.GetString(feedRkeyFlag))
			q.Add("service", viper.GetString(pdsURLFlag))
			u.RawQuery = q.Encode()

			req, err := http.NewRequest(http.MethodPut, u.String(), f)
			if err != nil {
				return err
			}

			req.Header.Set("Authorization", "Bearer "+auth.AccessJwt)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return errors.New(resp.Status)
			}
		}

		if (strings.TrimSpace(viper.GetString(feedPinnedDIDFlag)) != "" && strings.TrimSpace(viper.GetString(feedPinnedRkeyFlag)) != "") ||
			viper.GetBool(clearPinnedFlag) {
			u := u.JoinPath("admin", "feeds")

			q := u.Query()
			q.Add("rkey", viper.GetString(feedRkeyFlag))
			q.Add("service", viper.GetString(pdsURLFlag))
			q.Add("pinnedDID", viper.GetString(feedPinnedDIDFlag))
			q.Add("pinnedRkey", viper.GetString(feedPinnedRkeyFlag))
			u.RawQuery = q.Encode()

			req, err := http.NewRequest(http.MethodPatch, u.String(), nil)
			if err != nil {
				return err
			}

			req.Header.Set("Authorization", "Bearer "+auth.AccessJwt)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return errors.New(resp.Status)
			}
		}

		return nil
	},
}

func init() {
	applyCmd.PersistentFlags().String(feedRkeyFlag, "trending", "Machine-readable key for the feed")

	applyCmd.PersistentFlags().String(feedClassifierFlag, "local-trending-latest.scale", "Path to the feed classifier to upload")

	applyCmd.PersistentFlags().String(feedPinnedDIDFlag, "", "DID of the pinned post for the feed (if left empty, no post will be pinned; empty values don't overwrite non-empty values, see --clear-pinned)")
	applyCmd.PersistentFlags().String(feedPinnedRkeyFlag, "", "Machine-readable key of the pinned post for the feed (if left empty, no post will be pinned; empty values don't overwrite non-empty values, see --clear-pinned)")

	applyCmd.PersistentFlags().Bool(clearPinnedFlag, false, "Whether to clear the pinned post field")

	viper.AutomaticEnv()

	rootCmd.AddCommand(applyCmd)
}
