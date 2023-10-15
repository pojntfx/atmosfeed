package cmd

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deleteUserdata = &cobra.Command{
	Use:     "delete-userdata",
	Aliases: []string{"du"},
	Short:   "Delete all user data from an Atmosfeed server",
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

		u = u.JoinPath("userdata")

		q := u.Query()
		q.Add("service", viper.GetString(pdsURLFlag))
		u.RawQuery = q.Encode()

		req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
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

		return nil
	},
}

func init() {
	viper.AutomaticEnv()

	rootCmd.AddCommand(deleteUserdata)
}
