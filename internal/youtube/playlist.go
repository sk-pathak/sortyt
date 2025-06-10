package youtube

import (
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/youtube/v3"
)

func CreateSortedPlaylist(service *youtube.Service, videos []Video, originalTitle string) (string, error) {
	newPlaylist := &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title:       fmt.Sprintf("%s - Sorted", originalTitle),
			Description: "Sorted by upload date (oldest to newest)",
		},
		Status: &youtube.PlaylistStatus{PrivacyStatus: "private"},
	}

	resp, err := service.Playlists.Insert([]string{"snippet", "status"}, newPlaylist).Do()
	if err != nil {
		return "", err
	}

	playlistID := resp.Id
	total := len(videos)

	fmt.Printf("Created new playlist: \"%s\"\n", newPlaylist.Snippet.Title)
	fmt.Println("\nAdding videos to new playlist in chronological order...")

	for i := len(videos) - 1; i >= 0; i-- {
		video := videos[i]

		_, err := service.PlaylistItems.Insert([]string{"snippet"}, &youtube.PlaylistItem{
			Snippet: &youtube.PlaylistItemSnippet{
				PlaylistId: playlistID,
				ResourceId: &youtube.ResourceId{
					Kind:    "youtube#video",
					VideoId: video.ID,
				},
			},
		}).Do()
		if err != nil {
			fmt.Printf("\nSkipped \"%s\": %v\n", truncateString(video.Title, 50), err)
			continue
		}

		index := total - i
		bar := renderProgressBar(index, total, 40)
		fmt.Printf("\r%s %3d/%d - %s", bar, index, total,
			truncateString(video.Title, 40))

		time.Sleep(100 * time.Millisecond)
	}
	return playlistID, nil
}

func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func renderProgressBar(current, total, width int) string {
	if total == 0 {
		return "[" + strings.Repeat(" ", width) + "]"
	}
	progress := float64(current) / float64(total)
	filled := int(progress * float64(width))
	return "[" + strings.Repeat("█", filled) + strings.Repeat(" ", width-filled) + "]"
}
