package cmd

import (
	"fmt"
	"net/url"
	"os"

	"github.com/dedelala/sysexits"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.pkg.littleman.co/library/internal/server"
	"go.pkg.littleman.co/library/internal/server/middleware"
)

// srveCmd represents the srve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		// Default Options
		options := []server.Option{
			server.WithBook(viper.GetString("book.path")),
			server.WithLogging(),
		}

		// Add auth, if set
		if viper.IsSet("server.authentication.oidc") {
			urlStr := viper.GetString("server.authentication.oidc.callback_url")

			if len(urlStr) == 0 {
				fmt.Printf("unable to start server: oidc configuration invalid: url invalid: url empty")
				os.Exit(sysexits.DataErr)
			}

			url, err := url.Parse(urlStr)
			if err != nil {
				fmt.Printf("unable to start server: oidc configuration invalid: url invalid: %s", err.Error())
				os.Exit(sysexits.DataErr)
			}

			claimSets := []middleware.OIDCClaimSet{}

			// Iterate over the sets
			for _, rS := range viper.Get("server.authentication.oidc.claims").([]interface{}) {
				set := middleware.OIDCClaimSet{}

				for rK, rV := range rS.(map[interface{}]interface{}) {
					set[rK.(string)] = rV.(string)
				}

				claimSets = append(claimSets, set)
			}

			options = append(options, server.WithOIDCAuthentication(&server.OIDCConfig{
				Provider:     viper.GetString("server.authentication.oidc.provider"),
				ClientID:     viper.GetString("server.authentication.oidc.client.id"),
				ClientSecret: viper.GetString("server.authentication.oidc.client.secret"),
				RedirectURL:  url,
				ClaimSets:    claimSets,
			}))
		}

		srv, err := server.New(options...)

		if err != nil {
			fmt.Printf("unable to create server: %s", err.Error())
			os.Exit(sysexits.Software)
		}

		if err = srv.Serve(); err != nil {
			fmt.Printf("unable to start server: %s", err.Error())
			os.Exit(sysexits.Software)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
