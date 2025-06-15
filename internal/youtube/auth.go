package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
	"sortyt/internal/config"
)

const tokenFile = "token.json"

func NewService(cfg *config.Config) (*youtube.Service, error) {
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       []string{youtube.YoutubeScope},
		Endpoint:     google.Endpoint,
	}

	token, err := getToken(oauthCfg)
	if err != nil {
		return nil, err
	}

	client := oauthCfg.Client(context.Background(), token)
	return youtube.New(client)
}

func getToken(cfg *oauth2.Config) (*oauth2.Token, error) {
	if token, err := loadToken(); err == nil && token.Valid() {
		return token, nil
	}
	return getTokenFromWeb(cfg)
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

func getTokenFromWeb(cfg *oauth2.Config) (*oauth2.Token, error) {
	url := cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Visit this URL to authorize:\n%s\n", url)

	codeChan := make(chan string)
	server := &http.Server{Addr: ":8080"}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		codeChan <- r.URL.Query().Get("code")
		fmt.Fprintln(w, "Authorized! You can close this window.")
		go func() {
			time.Sleep(2 * time.Second)
			server.Shutdown(context.Background())
		}()
	})

	go server.ListenAndServe()
	code := <-codeChan

	token, err := cfg.Exchange(context.Background(), code)
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
