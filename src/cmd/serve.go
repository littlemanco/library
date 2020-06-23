package cmd

import (
	"fmt"
	"os"

	"github.com/dedelala/sysexits"
	"github.com/spf13/cobra"

	"go.pkg.littleman.co/library/internal/server"
)

// srveCmd represents the srve command
var srveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		srv, err := server.New()

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
	rootCmd.AddCommand(srveCmd)
}
