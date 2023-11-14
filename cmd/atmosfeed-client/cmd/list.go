package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type feedMetatadata struct {
	Rkey       string `json:"rkey"`
	PinnedDid  string `json:"pinnedDID"`
	PinnedRkey string `json:"pinnedRkey"`
}

func authorize(ctx context.Context) (*xrpc.Client, *xrpc.AuthInfo, error) {
	auth := &xrpc.AuthInfo{}

	client := &xrpc.Client{
		Client: http.DefaultClient,
		Host:   viper.GetString(pdsURLFlag),
		Auth:   auth,
	}

	session, err := atproto.ServerCreateSession(ctx, client, &atproto.ServerCreateSession_Input{
		Identifier: viper.GetString(usernameFlag),
		Password:   viper.GetString(passwordFlag),
	})
	if err != nil {
		return nil, nil, err
	}

	auth.AccessJwt = session.AccessJwt
	auth.RefreshJwt = session.RefreshJwt
	auth.Handle = session.Handle
	auth.Did = session.Did

	return client, auth, nil
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "List published feeds on an Atmosfeed server",
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

		u = u.JoinPath("admin", "feeds")

		q := u.Query()
		q.Add("service", viper.GetString(pdsURLFlag))
		u.RawQuery = q.Encode()

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

		feeds := []feedMetatadata{}
		if err := json.NewDecoder(resp.Body).Decode(&feeds); err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return errors.New(resp.Status)
		}

		return yaml.NewEncoder(os.Stdout).Encode(feeds)
	},
}

func init() {
	viper.AutomaticEnv()

	rootCmd.AddCommand(listCmd)
}
