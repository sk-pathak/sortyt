package cmd

import (
	"fmt"
	"log"

	"github.com/sk-pathak/sortyt/config"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up YouTube API credentials",
	Run: func(cmd *cobra.Command, args []string) {
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

		var cfg config.Config

		fmt.Print("Client ID: ")
		fmt.Scanln(&cfg.ClientID)
		fmt.Print("Client Secret: ")
		fmt.Scanln(&cfg.ClientSecret)

		cfg.RedirectURL = "http://localhost:8080/callback"

		if err := config.Save(cfg); err != nil {
			log.Fatalf("Failed to save config: %v", err)
		}

		fmt.Println("Configuration saved successfully!")
	},
}
