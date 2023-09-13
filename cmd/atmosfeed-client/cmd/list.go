package cmd

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	atmosfeedURLFlag = "atmosfeed-url"

	pdsURLFlag   = "pds-url"
	usernameFlag = "username"
	passwordFlag = "password"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "List published feeds",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}

		u, err := url.Parse(viper.GetString(atmosfeedURLFlag))
		if err != nil {
			return err
		}

		auth := &xrpc.AuthInfo{}

		client := &xrpc.Client{
			Client: http.DefaultClient,
			Host:   viper.GetString(pdsURLFlag),
			Auth:   auth,
		}

		session, err := atproto.ServerCreateSession(cmd.Context(), client, &atproto.ServerCreateSession_Input{
			Identifier: viper.GetString(usernameFlag),
			Password:   viper.GetString(passwordFlag),
		})
		if err != nil {
			return err
		}

		auth.AccessJwt = session.AccessJwt
		auth.RefreshJwt = session.RefreshJwt
		auth.Handle = session.Handle
		auth.Did = session.Did

		u = u.JoinPath("admin", "feeds")

		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			return err
		}

		req.Header.Set("Authorization", "Bearer "+auth.AccessJwt)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		feeds := []string{}
		if err := json.NewDecoder(resp.Body).Decode(&feeds); err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			panic(resp.Status)
		}

		return yaml.NewEncoder(os.Stdout).Encode(feeds)
	},
}

func init() {
	listCmd.PersistentFlags().String(atmosfeedURLFlag, "http://localhost:1337", "Atmosfeed server URL")

	listCmd.PersistentFlags().String(pdsURLFlag, "https://bsky.social", "PDS URL")
	listCmd.PersistentFlags().String(usernameFlag, "example.bsky.social", "Bluesky username (ignored if `--publish` is not provided)")
	listCmd.PersistentFlags().String(passwordFlag, "", "Bluesky password, preferably an app password (get one from https://bsky.app/settings/app-passwords, ignored if `--publish` is not provided)")

	viper.AutomaticEnv()

	rootCmd.AddCommand(listCmd)
}
