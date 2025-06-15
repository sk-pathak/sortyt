package cmd

import (
	"fmt"
	"log"

	"sortyt/config"
	"sortyt/internal/utils"
	"sortyt/internal/youtube"

	"github.com/spf13/cobra"
)

var sortCmd = &cobra.Command{
	Use:   "sort [playlist-url]",
	Short: "Sort a YouTube playlist by date",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		fmt.Printf("Processing playlist: %s\n", url)

		playlistID := utils.ExtractPlaylistID(url)
		if playlistID == "" {
			log.Fatal("Invalid playlist URL.")
		}

		cfg, err := config.Load()
		if err != nil {
			log.Fatal("Load config failed:", err)
		}

		service, err := youtube.NewService(cfg)
		if err != nil {
			log.Fatal("Could not create YouTube service:", err)
		}

		videos, playlistTitle, err := youtube.FetchAndSortVideos(service, playlistID)
		if err != nil {
			log.Fatal("Failed fetching/sorting videos:", err)
		}
		fmt.Printf("Found %d videos\n", len(videos))

		newID, err := youtube.CreateSortedPlaylist(service, videos, playlistTitle)
		if err != nil {
			log.Fatal("Failed to create playlist:", err)
		}

		fmt.Printf("New playlist created: https://www.youtube.com/playlist?list=%s\n", newID)
	},
}
