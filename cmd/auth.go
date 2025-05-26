package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/itcaat/slacker/internal/api"
	"github.com/itcaat/slacker/internal/config"
	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Slack",
	Long: `Authenticate with Slack using a token. You can provide the token as an argument
or set it via the SLACKER_SLACK_TOKEN environment variable.

To get a Slack token:
1. Go to https://api.slack.com/apps
2. Create a new app or select an existing one
3. Go to "OAuth & Permissions"
4. Add the following scopes:
   - channels:history
   - channels:read
   - groups:history
   - groups:read
   - users:read
5. Install the app to your workspace
6. Copy the "Bot User OAuth Token"

Examples:
  slacker auth xoxb-your-token-here
  SLACKER_SLACK_TOKEN=xoxb-your-token-here slacker auth test`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			// Token provided as argument
			token := args[0]
			if err := setAndTestToken(token); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Test existing token
			if err := testExistingToken(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

// authTestCmd represents the auth test command
var authTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test current authentication",
	Long:  `Test the current Slack authentication without setting a new token.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := testExistingToken(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authTestCmd)
}

func setAndTestToken(token string) error {
	fmt.Println("Setting and testing Slack token...")

	// Save token to config
	configManager := config.NewManager()
	if err := configManager.SetToken(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Println("✅ Token saved to configuration")

	// Test the token
	return testToken(token)
}

func testExistingToken() error {
	fmt.Println("Testing existing Slack authentication...")

	// Get token from config or environment
	configManager := config.NewManager()
	token, err := configManager.GetToken()
	if err != nil {
		return err
	}

	return testToken(token)
}

func testToken(token string) error {
	// Create Slack client
	client := api.NewSlackClient(token, true) // Enable debug for auth testing

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test authentication
	fmt.Println("🔄 Testing Slack API connection...")
	authResponse, err := client.TestAuth(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Printf("✅ Authentication successful!\n")
	fmt.Printf("   User: %s\n", authResponse.User)
	fmt.Printf("   Team: %s\n", authResponse.Team)
	fmt.Printf("   URL: %s\n", authResponse.URL)

	// Test getting channels
	fmt.Println("🔄 Testing channel access...")
	channels, err := client.GetChannels(ctx)
	if err != nil {
		return fmt.Errorf("failed to get channels: %w", err)
	}

	fmt.Printf("✅ Channel access successful! Found %d channels\n", len(channels))

	// Show first few channels as examples
	if len(channels) > 0 {
		fmt.Println("   Sample channels:")
		for i, channel := range channels {
			if i >= 3 { // Show max 3 channels
				break
			}
			fmt.Printf("   - #%s (%d members)\n", channel.Name, channel.NumMembers)
		}
		if len(channels) > 3 {
			fmt.Printf("   ... and %d more\n", len(channels)-3)
		}
	}

	return nil
}
