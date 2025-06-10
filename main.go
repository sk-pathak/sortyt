package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

type Video struct {
	ID           string
	Title        string
	PublishedAt  time.Time
	AddedToPlaylist time.Time
}

type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
}

var (
	configFile    = "config.json"
	tokenFile     = "token.json"
	youtubeScopes = []string{youtube.YoutubeScope}
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "yt-playlist-sorter",
		Short: "Sort YouTube playlist by date and create a new private playlist",
		Long: `A CLI tool that takes any YouTube playlist URL, sorts videos by publication date, and creates a new private playlist in chronological order.`,
	}

	var sortCmd = &cobra.Command{
		Use:   "sort [playlist-url]",
		Short: "Sort a YouTube playlist by date",
		Args:  cobra.ExactArgs(1),
		Run:   runSort,
	}

	var setupCmd = &cobra.Command{
		Use:   "setup",
		Short: "Set up YouTube API credentials",
		Run:   runSetup,
	}

	rootCmd.AddCommand(sortCmd, setupCmd)
	
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runSetup(cmd *cobra.Command, args []string) {
	fmt.Println("Setting up YouTube API credentials...")
	fmt.Println()
	fmt.Println("IMPORTANT: To avoid OAuth verification issues, follow these steps:")
	fmt.Println()
	fmt.Println("1. Go to https://console.cloud.google.com/")
	fmt.Println("2. Create a new project or select existing one")
	fmt.Println("3. Enable YouTube Data API v3")
	fmt.Println("4. Go to OAuth consent screen:")
	fmt.Println("   - Choose 'External' user type")
	fmt.Println("   - Fill required fields (App name, User support email, Developer email)")
	fmt.Println("   - Add your Gmail address to 'Test users' section")
	fmt.Println("   - Save and continue through all steps")
	fmt.Println("5. Create OAuth 2.0 credentials (Desktop application)")
	fmt.Println("6. Download the credentials and enter details below:")
	fmt.Println()
	fmt.Println("Note: Your app will be in 'Testing' mode - only test users can access it.")
	fmt.Println()

	var config Config
	fmt.Print("Client ID: ")
	fmt.Scanln(&config.ClientID)
	fmt.Print("Client Secret: ")
	fmt.Scanln(&config.ClientSecret)
	config.RedirectURL = "http://localhost:8080/callback"

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Fatal("Error marshaling config:", err)
	}

	err = os.WriteFile(configFile, data, 0644)
	if err != nil {
		log.Fatal("Error writing config file:", err)
	}

	fmt.Println("Configuration saved successfully!")
	fmt.Println("Now run: yt-playlist-sorter sort <playlist-url>")
}

func runSort(cmd *cobra.Command, args []string) {
	playlistURL := args[0]
	
	fmt.Printf("Processing playlist: %s\n", playlistURL)
	
	playlistID := extractPlaylistID(playlistURL)
	if playlistID == "" {
		log.Fatal("Invalid YouTube playlist URL")
	}
	
	service, err := getYouTubeService()
	if err != nil {
		log.Fatal("Error initializing YouTube service:", err)
	}

	playlistTitle, err := getPlaylistTitle(service, playlistID)
	if err != nil {
		log.Fatal("Error fetching playlist title:", err)
	}
	
	videos, err := getPlaylistVideos(service, playlistID)
	if err != nil {
		log.Fatal("Error fetching playlist videos:", err)
	}
	
	fmt.Printf("📹 Found %d videos\n", len(videos))
	
	err = fetchVideoUploadDates(service, videos)
	if err != nil {
		log.Fatal("Error fetching video upload dates:", err)
	}
	
	sort.Slice(videos, func(i, j int) bool {
		return videos[i].PublishedAt.After(videos[j].PublishedAt)
	})
	
	fmt.Println("Videos sorted by actual upload date (oldest to newest)")
	
	newPlaylistID, err := createSortedPlaylist(service, videos, playlistTitle)
	if err != nil {
		log.Fatal("Error creating new playlist:", err)
	}
	
	fmt.Printf("Created new private playlist: https://www.youtube.com/playlist?list=%s\n", newPlaylistID)
}

func extractPlaylistID(url string) string {
	patterns := []string{
		`[?&]list=([^&]+)`,
		`/playlist\?list=([^&]+)`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

func getYouTubeService() (*youtube.Service, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("config not found. Run 'yt-playlist-sorter setup' first")
	}
	
	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       youtubeScopes,
		Endpoint:     google.Endpoint,
	}
	
	token, err := getToken(oauthConfig)
	if err != nil {
		return nil, err
	}
	
	client := oauthConfig.Client(context.Background(), token)
	
	service, err := youtube.New(client)
	if err != nil {
		return nil, err
	}
	
	return service, nil
}

func loadConfig() (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	
	var config Config
	err = json.Unmarshal(data, &config)
	return &config, err
}

func getToken(config *oauth2.Config) (*oauth2.Token, error) {
	if token, err := loadToken(); err == nil {
		if token.Valid() {
			return token, nil
		}
	}
	
	return getTokenFromWeb(config)
}

func loadToken() (*oauth2.Token, error) {
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return nil, err
	}
	
	var token oauth2.Token
	err = json.Unmarshal(data, &token)
	return &token, err
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Please visit this URL to authorize the application:\n%s\n", authURL)
	
	tokenChan := make(chan string)
	server := &http.Server{Addr: ":8080"}
	
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		tokenChan <- code
		fmt.Fprintf(w, "Authorization successful! You can close this window.")
		go func() {
			time.Sleep(2 * time.Second)
			server.Shutdown(context.Background())
		}()
	})
	
	go server.ListenAndServe()
	
	fmt.Println("Waiting for authorization...")
	code := <-tokenChan
	
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}
	
	saveToken(token)
	
	return token, nil
}

func saveToken(token *oauth2.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return os.WriteFile(tokenFile, data, 0644)
}

func getPlaylistVideos(service *youtube.Service, playlistID string) ([]Video, error) {
	var videos []Video
	pageToken := ""
	
	for {
		call := service.PlaylistItems.List([]string{"snippet"}).
			PlaylistId(playlistID).
			MaxResults(50)
		
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		
		response, err := call.Do()
		if err != nil {
			return nil, err
		}
		
		for _, item := range response.Items {
			addedToPlaylist, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
			if err != nil {
				continue
			}
			
			videos = append(videos, Video{
				ID:              item.Snippet.ResourceId.VideoId,
				Title:           item.Snippet.Title,
				PublishedAt:     time.Time{},
				AddedToPlaylist: addedToPlaylist,
			})
		}
		
		pageToken = response.NextPageToken
		if pageToken == "" {
			break
		}
	}
	
	return videos, nil
}

func fetchVideoUploadDates(service *youtube.Service, videos []Video) error {
	batchSize := 50
	
	for i := 0; i < len(videos); i += batchSize {
		end := i + batchSize
		if end > len(videos) {
			end = len(videos)
		}
		
		var videoIds []string
		for j := i; j < end; j++ {
			videoIds = append(videoIds, videos[j].ID)
		}
		
		call := service.Videos.List([]string{"snippet"}).
			Id(strings.Join(videoIds, ","))
		
		response, err := call.Do()
		if err != nil {
			return fmt.Errorf("error fetching video details: %v", err)
		}
		
		videoMap := make(map[string]*youtube.Video)
		for _, video := range response.Items {
			videoMap[video.Id] = video
		}
		
		for j := i; j < end; j++ {
			if videoData, exists := videoMap[videos[j].ID]; exists {
				uploadDate, err := time.Parse(time.RFC3339, videoData.Snippet.PublishedAt)
				if err == nil {
					videos[j].PublishedAt = uploadDate
					videos[j].Title = videoData.Snippet.Title
				}
			}
		}
		
		fmt.Printf("Fetched upload dates for %d/%d videos\n", end, len(videos))
		
		time.Sleep(100 * time.Millisecond)
	}
	
	return nil
}

func getPlaylistTitle(service *youtube.Service, playlistID string) (string, error) {
	call := service.Playlists.List([]string{"snippet"}).Id(playlistID)
	resp, err := call.Do()
	if err != nil {
		return "", err
	}
	if len(resp.Items) == 0 {
		return "", fmt.Errorf("playlist not found")
	}
	return resp.Items[0].Snippet.Title, nil
}

func createSortedPlaylist(service *youtube.Service, videos []Video, originalTitle string) (string, error) {
	playlistTitle := fmt.Sprintf("%s - Sorted", originalTitle)
	
	playlist := &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title:       playlistTitle,
			Description: "Playlist sorted by video publication date (oldest to newest) using yt-playlist-sorter",
		},
		Status: &youtube.PlaylistStatus{
			PrivacyStatus: "private",
		},
	}
	
	insertCall := service.Playlists.Insert([]string{"snippet", "status"}, playlist)
	createdPlaylist, err := insertCall.Do()
	if err != nil {
		return "", err
	}
	
	playlistID := createdPlaylist.Id
	
	fmt.Println("Adding videos to new playlist (in reverse to achieve chronological order)...")
	
	for i := len(videos) - 1; i >= 0; i-- {
		video := videos[i]
		playlistItem := &youtube.PlaylistItem{
			Snippet: &youtube.PlaylistItemSnippet{
				PlaylistId: playlistID,
				ResourceId: &youtube.ResourceId{
					Kind:    "youtube#video",
					VideoId: video.ID,
				},
			},
		}
		
		_, err := service.PlaylistItems.Insert([]string{"snippet"}, playlistItem).Do()
		if err != nil {
			fmt.Printf("Warning: Could not add video '%s': %v\n", video.Title, err)
			continue
		}
		
		chronoIndex := len(videos) - i
		fmt.Printf("✓ Added %d/%d: %s (Uploaded: %s)\n", 
			chronoIndex, len(videos), 
			truncateString(video.Title, 50), 
			video.PublishedAt.Format("2006-01-02"))
		
		time.Sleep(100 * time.Millisecond)
	}
	
	return playlistID, nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
