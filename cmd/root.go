package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "yt-playlist-sorter",
	Short: "Sort YouTube playlists by video upload date",
	Long:  `A CLI tool to sort YouTube playlists chronologically and create a new sorted private playlist.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(sortCmd)
}
