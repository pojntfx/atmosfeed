package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pojntfx/atmosfeed/pkg/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	outFlag = "out"
)

type structuredUserdata struct {
	Feeds     []models.Feed     `json:"feeds"`
	Posts     []models.Post     `json:"posts"`
	FeedPosts []models.FeedPost `json:"feedPosts"`
}

var exportUserdata = &cobra.Command{
	Use:     "export-userdata",
	Aliases: []string{"eu"},
	Short:   "Export all user data from an Atmosfeed server",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}

		_, auth, err := authorize(cmd.Context())
		if err != nil {
			return err
		}

		classifiersDir := filepath.Join(viper.GetString(outFlag), "blobs", "classifiers")
		if err := os.MkdirAll(classifiersDir, os.ModePerm); err != nil {
			return err
		}

		var structuredData structuredUserdata
		{
			u, err := url.Parse(viper.GetString(atmosfeedURLFlag))
			if err != nil {
				return err
			}

			u = u.JoinPath("userdata", "structured")

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

			if resp.StatusCode != http.StatusOK {
				return errors.New(resp.Status)
			}

			if err := json.NewDecoder(resp.Body).Decode(&structuredData); err != nil {
				return err
			}

			structuredDataFile, err := os.OpenFile(filepath.Join(viper.GetString(outFlag), "structured.json"), os.O_RDWR|os.O_TRUNC|os.O_CREATE, os.ModePerm)
			if err != nil {
				return err
			}
			defer structuredDataFile.Close()

			b, err := json.Marshal(structuredData)
			if err != nil {
				return err
			}

			if _, err := io.Copy(structuredDataFile, bytes.NewBuffer(b)); err != nil {
				return err
			}
		}

		{
			for _, feed := range structuredData.Feeds {
				u, err := url.Parse(viper.GetString(atmosfeedURLFlag))
				if err != nil {
					return err
				}

				u = u.JoinPath("userdata", "blob")

				q := u.Query()
				q.Add("service", viper.GetString(pdsURLFlag))
				q.Add("resource", "classifier")
				q.Add("rkey", feed.Rkey)
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

				classifierFile, err := os.OpenFile(filepath.Join(classifiersDir, feed.Rkey+".scale"), os.O_RDWR|os.O_TRUNC|os.O_CREATE, os.ModePerm)
				if err != nil {
					return err
				}
				defer classifierFile.Close()

				if _, err := io.Copy(classifierFile, resp.Body); err != nil {
					return err
				}
			}
		}

		return nil
	},
}

func init() {
	exportUserdata.PersistentFlags().String(outFlag, "atmosfeed-userdata", "Directory to export user data to")

	viper.AutomaticEnv()

	rootCmd.AddCommand(exportUserdata)
}
