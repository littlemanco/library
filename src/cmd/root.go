package cmd

import (
	"fmt"
	"os"

	"github.com/dedelala/sysexits"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "library",
	Short: "A server to make e-books available in the browser",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(sysexits.Usage)
	},
}

// Execute is the entrypoint for the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default /etc/.library.yaml)")
	rootCmd.PersistentFlags().StringP("book-path", "p", "/book.epub", "The path to the book that should be rendered")

	viper.BindPFlag("book.path", rootCmd.PersistentFlags().Lookup("book-path"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("/etc")
		viper.SetConfigName(".library")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
