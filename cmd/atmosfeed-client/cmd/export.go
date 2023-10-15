package cmd

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	outFlag = "out"
)

var exportCmd = &cobra.Command{
	Use:     "export",
	Aliases: []string{"e"},
	Short:   "Export all user data from an Atmosfeed server",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}

		_, auth, err := authorize(cmd.Context())
		if err != nil {
			return err
		}

		{
			if err := os.MkdirAll(viper.GetString(outFlag), os.ModePerm); err != nil {
				return err
			}

			u, err := url.Parse(viper.GetString(atmosfeedURLFlag))
			if err != nil {
				return err
			}

			u = u.JoinPath("userdata", "structured")

			q := u.Query()
			q.Add("rkey", viper.GetString(feedRkeyFlag))
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

			if resp.StatusCode != http.StatusOK {
				return errors.New(resp.Status)
			}

			structuredDataFile, err := os.OpenFile(filepath.Join(viper.GetString(outFlag), "structured.json"), os.O_RDWR|os.O_TRUNC|os.O_CREATE, os.ModePerm)
			if err != nil {
				return err
			}
			defer structuredDataFile.Close()

			if _, err := io.Copy(structuredDataFile, resp.Body); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	exportCmd.PersistentFlags().String(outFlag, "atmosfeed-userdata", "Directory to export user data to")

	viper.AutomaticEnv()

	rootCmd.AddCommand(exportCmd)
}
