package youtube

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"google.golang.org/api/youtube/v3"
)

type Video struct {
	ID              string
	Title           string
	PublishedAt     time.Time
	AddedToPlaylist time.Time
}

func FetchAndSortVideos(service *youtube.Service, playlistID string) ([]Video, string, error) {
	videos, err := getPlaylistVideos(service, playlistID)
	if err != nil {
		return nil, "", err
	}

	err = populateUploadDates(service, videos)
	if err != nil {
		return nil, "", err
	}

	sort.Slice(videos, func(i, j int) bool {
		return videos[i].PublishedAt.After(videos[j].PublishedAt)
	})

	title, err := getPlaylistTitle(service, playlistID)
	return videos, title, err
}

func getPlaylistTitle(service *youtube.Service, id string) (string, error) {
	call := service.Playlists.List([]string{"snippet"}).Id(id)
	resp, err := call.Do()
	if err != nil || len(resp.Items) == 0 {
		return "", fmt.Errorf("playlist not found")
	}
	return resp.Items[0].Snippet.Title, nil
}

func getPlaylistVideos(service *youtube.Service, playlistID string) ([]Video, error) {
	var videos []Video
	pageToken := ""
	for {
		resp, err := service.PlaylistItems.List([]string{"snippet"}).PlaylistId(playlistID).MaxResults(50).PageToken(pageToken).Do()
		if err != nil {
			return nil, err
		}
		for _, item := range resp.Items {
			added, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
			videos = append(videos, Video{
				ID:              item.Snippet.ResourceId.VideoId,
				Title:           item.Snippet.Title,
				AddedToPlaylist: added,
			})
		}
		if pageToken = resp.NextPageToken; pageToken == "" {
			break
		}
	}
	return videos, nil
}

func populateUploadDates(service *youtube.Service, videos []Video) error {
	batchSize := 50
	for i := 0; i < len(videos); i += batchSize {
		end := min(i+batchSize, len(videos))

		var ids []string
		for _, v := range videos[i:end] {
			ids = append(ids, v.ID)
		}

		resp, err := service.Videos.List([]string{"snippet"}).Id(strings.Join(ids, ",")).Do()
		if err != nil {
			return err
		}

		videoMap := make(map[string]*youtube.Video)
		for _, v := range resp.Items {
			videoMap[v.Id] = v
		}

		for j := i; j < end; j++ {
			if v, ok := videoMap[videos[j].ID]; ok {
				t, _ := time.Parse(time.RFC3339, v.Snippet.PublishedAt)
				videos[j].PublishedAt = t
				videos[j].Title = v.Snippet.Title
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}
