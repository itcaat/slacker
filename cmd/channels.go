package cmd

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/itcaat/slacker/internal/api"
	"github.com/itcaat/slacker/internal/config"
	"github.com/itcaat/slacker/models"
	"github.com/spf13/cobra"
)

// channelsCmd represents the channels command
var channelsCmd = &cobra.Command{
	Use:   "channels",
	Short: "List and interact with Slack channels",
	Long: `The channels command provides functionality to list all available channels
that the user is part of and allows selection for viewing messages or exporting.

Examples:
  slacker channels list    # List all available channels
  slacker channels view    # Interactive channel browser (TUI)`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("channels called - TUI interface will be implemented here")
	},
}

// channelsListCmd represents the channels list command
var channelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available channels",
	Long: `List all Slack channels that the authenticated user is a member of.
This command will display channel names, types, and member counts.

Flags:
  -a, --include-archived  Include archived channels in the list
  -p, --private-only      Show only private channels
  -u, --public-only       Show only public channels  
  -f, --format string     Output format: table, json, csv (default "table")
  -v, --verbose           Show detailed channel information

Examples:
  slacker channels list                    # List all active channels
  slacker channels list -a                 # Include archived channels
  slacker channels list -p                 # Show only private channels
  slacker channels list -f json            # Output as JSON
  slacker channels list -v                 # Show detailed information`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := listChannels(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// channelsViewCmd represents the channels view command
var channelsViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Interactive channel browser",
	Long: `Launch the interactive TUI for browsing channels and viewing messages.
This provides a full-screen interface for navigating channels and reading conversations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("channels view called - will implement Bubble Tea TUI")
	},
}

func init() {
	rootCmd.AddCommand(channelsCmd)
	channelsCmd.AddCommand(channelsListCmd)
	channelsCmd.AddCommand(channelsViewCmd)

	// Add flags for channel listing
	channelsListCmd.Flags().BoolP("include-archived", "a", false, "Include archived channels")
	channelsListCmd.Flags().BoolP("private-only", "p", false, "Show only private channels")
	channelsListCmd.Flags().BoolP("public-only", "u", false, "Show only public channels")
	channelsListCmd.Flags().StringP("format", "f", "table", "Output format: table, json, csv")
	channelsListCmd.Flags().BoolP("verbose", "v", false, "Show detailed channel information")
}

func listChannels(cmd *cobra.Command) error {
	// Get flags
	includeArchived, _ := cmd.Flags().GetBool("include-archived")
	privateOnly, _ := cmd.Flags().GetBool("private-only")
	publicOnly, _ := cmd.Flags().GetBool("public-only")
	format, _ := cmd.Flags().GetString("format")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Get token from config
	configManager := config.NewManager()
	token, err := configManager.GetToken()
	if err != nil {
		return fmt.Errorf("authentication required. Run 'slacker auth <token>' first: %w", err)
	}

	// Create Slack client
	client := api.NewSlackClient(token, false)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("🔄 Fetching channels...")

	// Get channels
	channels, err := client.GetChannels(ctx)
	if err != nil {
		return fmt.Errorf("failed to get channels: %w", err)
	}

	// Filter channels based on flags
	filteredChannels := filterChannels(channels, includeArchived, privateOnly, publicOnly)

	if len(filteredChannels) == 0 {
		fmt.Println("No channels found matching the criteria.")
		return nil
	}

	// Output in requested format
	switch format {
	case "json":
		return outputChannelsJSON(filteredChannels)
	case "csv":
		return outputChannelsCSV(filteredChannels, verbose)
	case "table":
		fallthrough
	default:
		return outputChannelsTable(filteredChannels, verbose)
	}
}

// filterChannels filters channels based on the provided criteria
func filterChannels(channels []models.Channel, includeArchived, privateOnly, publicOnly bool) []models.Channel {
	var filtered []models.Channel

	for _, channel := range channels {
		// Skip archived channels unless explicitly requested
		if channel.IsArchived && !includeArchived {
			continue
		}

		// Filter by privacy settings
		if privateOnly && !channel.IsPrivate {
			continue
		}
		if publicOnly && channel.IsPrivate {
			continue
		}

		filtered = append(filtered, channel)
	}

	return filtered
}

// outputChannelsJSON outputs channels in JSON format
func outputChannelsJSON(channels []models.Channel) error {
	data, err := json.MarshalIndent(channels, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal channels to JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// outputChannelsCSV outputs channels in CSV format
func outputChannelsCSV(channels []models.Channel, verbose bool) error {

	var records [][]string

	// Header
	if verbose {
		records = append(records, []string{"Name", "ID", "Type", "Members", "Archived", "Topic", "Purpose"})
	} else {
		records = append(records, []string{"Name", "Type", "Members", "Archived"})
	}

	// Data rows
	for _, channel := range channels {
		channelType := getChannelType(channel)
		archived := "No"
		if channel.IsArchived {
			archived = "Yes"
		}

		if verbose {
			records = append(records, []string{
				channel.Name,
				channel.ID,
				channelType,
				fmt.Sprintf("%d", channel.NumMembers),
				archived,
				channel.Topic.Value,
				channel.Purpose.Value,
			})
		} else {
			records = append(records, []string{
				channel.Name,
				channelType,
				fmt.Sprintf("%d", channel.NumMembers),
				archived,
			})
		}
	}

	// Output CSV
	var output strings.Builder
	writer := csv.NewWriter(&output)
	writer.WriteAll(records)

	if err := writer.Error(); err != nil {
		return fmt.Errorf("failed to write CSV: %w", err)
	}

	fmt.Print(output.String())
	return nil
}

// outputChannelsTable outputs channels in table format
func outputChannelsTable(channels []models.Channel, verbose bool) error {
	fmt.Printf("Found %d channels:\n\n", len(channels))

	if verbose {
		// Detailed table format
		for _, channel := range channels {
			channelType := getChannelType(channel)
			status := ""
			if channel.IsArchived {
				status = " (archived)"
			}

			fmt.Printf("📢 #%-20s %s%s\n", channel.Name, channelType, status)
			fmt.Printf("   ID: %s\n", channel.ID)
			fmt.Printf("   Members: %d\n", channel.NumMembers)
			if channel.Topic.Value != "" {
				fmt.Printf("   Topic: %s\n", channel.Topic.Value)
			}
			if channel.Purpose.Value != "" {
				fmt.Printf("   Purpose: %s\n", channel.Purpose.Value)
			}
			fmt.Printf("   Created: %s\n", time.Unix(channel.Created, 0).Format("2006-01-02 15:04:05"))
			fmt.Println()
		}
	} else {
		// Simple table format
		for _, channel := range channels {
			channelType := getChannelType(channel)
			status := ""
			if channel.IsArchived {
				status = " (archived)"
			}

			fmt.Printf("📢 #%-20s %-8s %3d members%s\n", channel.Name, channelType, channel.NumMembers, status)
		}
	}

	return nil
}

// getChannelType returns a human-readable channel type
func getChannelType(channel models.Channel) string {
	if channel.IsIM {
		return "DM"
	}
	if channel.IsGroup {
		return "Group"
	}
	if channel.IsPrivate {
		return "Private"
	}
	return "Public"
}
