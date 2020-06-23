package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// srveCmd represents the srve command
var srveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("srve called")
	},
}

func init() {
	rootCmd.AddCommand(srveCmd)
}
