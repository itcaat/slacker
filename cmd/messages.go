package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/itcaat/slacker/internal/api"
	"github.com/itcaat/slacker/internal/config"
	"github.com/itcaat/slacker/models"
	"github.com/spf13/cobra"
)

// messagesCmd represents the messages command
var messagesCmd = &cobra.Command{
	Use:   "messages",
	Short: "View and retrieve channel message history",
	Long: `View and retrieve message history from Slack channels. This command allows you to
browse recent messages, search through history, and view threaded conversations.

Examples:
  slacker messages --channel general           # View recent messages from #general
  slacker messages --channel general --limit 50  # View last 50 messages
  slacker messages --channel general --threads   # Include thread replies`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := viewMessages(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(messagesCmd)

	// Channel selection
	messagesCmd.Flags().StringP("channel", "c", "", "Channel name to view messages from (required)")
	messagesCmd.MarkFlagRequired("channel")

	// Message retrieval options
	messagesCmd.Flags().IntP("limit", "l", 20, "Number of messages to retrieve (default: 20)")
	messagesCmd.Flags().BoolP("threads", "t", false, "Include thread replies")
	messagesCmd.Flags().StringP("before", "b", "", "Show messages before this timestamp")
	messagesCmd.Flags().StringP("after", "a", "", "Show messages after this timestamp")

	// Output options
	messagesCmd.Flags().StringP("format", "f", "text", "Output format: text, json")
	messagesCmd.Flags().BoolP("verbose", "v", false, "Show detailed message information")
	messagesCmd.Flags().BoolP("no-format", "n", false, "Disable text formatting and colors")
}

func viewMessages(cmd *cobra.Command) error {
	// Get flags
	channelName, _ := cmd.Flags().GetString("channel")
	limit, _ := cmd.Flags().GetInt("limit")
	includeThreads, _ := cmd.Flags().GetBool("threads")
	before, _ := cmd.Flags().GetString("before")
	after, _ := cmd.Flags().GetString("after")
	format, _ := cmd.Flags().GetString("format")
	verbose, _ := cmd.Flags().GetBool("verbose")
	noFormat, _ := cmd.Flags().GetBool("no-format")

	// Validate limit
	if limit <= 0 || limit > 1000 {
		return fmt.Errorf("limit must be between 1 and 1000")
	}

	// Get token from config
	configManager := config.NewManager()
	token, err := configManager.GetToken()
	if err != nil {
		return fmt.Errorf("authentication required. Run 'slacker auth <token>' first: %w", err)
	}

	// Create Slack client
	client := api.NewSlackClient(token, false)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Find channel by name
	fmt.Printf("🔄 Finding channel #%s...\n", channelName)
	channel, err := client.GetChannelByName(ctx, channelName)
	if err != nil {
		return fmt.Errorf("failed to find channel: %w", err)
	}

	fmt.Printf("📢 Found channel: #%s (%s)\n", channel.Name, channel.ID)

	// Get message history
	fmt.Printf("🔄 Fetching message history (limit: %d)...\n", limit)
	messages, err := getChannelMessages(ctx, client, channel.ID, limit, before, after)
	if err != nil {
		return fmt.Errorf("failed to get messages: %w", err)
	}

	if len(messages) == 0 {
		fmt.Println("No messages found in the specified range.")
		return nil
	}

	// Get thread replies if requested
	if includeThreads {
		fmt.Println("🔄 Fetching thread replies...")
		messages, err = enrichWithThreads(ctx, client, channel.ID, messages)
		if err != nil {
			return fmt.Errorf("failed to get thread replies: %w", err)
		}
	}

	// Get user information for better display
	fmt.Println("🔄 Fetching user information...")
	users, err := client.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get users: %w", err)
	}

	userMap := make(map[string]models.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// Output messages
	switch format {
	case "json":
		return outputMessagesJSON(messages, userMap)
	case "text":
		fallthrough
	default:
		return outputMessagesText(messages, userMap, verbose, noFormat)
	}
}

// getChannelMessages retrieves messages from a channel with pagination
func getChannelMessages(ctx context.Context, client *api.SlackClient, channelID string, limit int, before, after string) ([]models.Message, error) {
	var allMessages []models.Message
	cursor := ""
	remaining := limit

	for remaining > 0 {
		// Calculate batch size (max 200 per API call)
		batchSize := remaining
		if batchSize > 200 {
			batchSize = 200
		}

		messages, nextCursor, err := client.GetChannelHistory(ctx, channelID, batchSize, cursor)
		if err != nil {
			return nil, err
		}

		// Filter messages by time range if specified
		filteredMessages := filterMessagesByTime(messages, before, after)
		allMessages = append(allMessages, filteredMessages...)

		remaining -= len(filteredMessages)
		cursor = nextCursor

		// Stop if no more messages or no cursor for next page
		if nextCursor == "" || len(messages) == 0 {
			break
		}

		// Stop if we've collected enough messages
		if len(allMessages) >= limit {
			break
		}
	}

	// Trim to exact limit if we got more than requested
	if len(allMessages) > limit {
		allMessages = allMessages[:limit]
	}

	return allMessages, nil
}

// filterMessagesByTime filters messages based on before/after timestamps
func filterMessagesByTime(messages []models.Message, before, after string) []models.Message {
	if before == "" && after == "" {
		return messages
	}

	var filtered []models.Message
	for _, msg := range messages {
		msgTime, err := strconv.ParseFloat(msg.Timestamp, 64)
		if err != nil {
			continue // Skip messages with invalid timestamps
		}

		// Check before constraint
		if before != "" {
			beforeTime, err := strconv.ParseFloat(before, 64)
			if err == nil && msgTime >= beforeTime {
				continue
			}
		}

		// Check after constraint
		if after != "" {
			afterTime, err := strconv.ParseFloat(after, 64)
			if err == nil && msgTime <= afterTime {
				continue
			}
		}

		filtered = append(filtered, msg)
	}

	return filtered
}

// enrichWithThreads fetches thread replies for messages that have them
func enrichWithThreads(ctx context.Context, client *api.SlackClient, channelID string, messages []models.Message) ([]models.Message, error) {
	for i, msg := range messages {
		if msg.ReplyCount > 0 && msg.ThreadTS != "" {
			replies, err := client.GetThreadReplies(ctx, channelID, msg.ThreadTS)
			if err != nil {
				// Log error but continue with other messages
				fmt.Fprintf(os.Stderr, "Warning: Failed to get replies for message %s: %v\n", msg.Timestamp, err)
				continue
			}
			messages[i].Thread = replies
		}
	}
	return messages, nil
}

// outputMessagesJSON outputs messages in JSON format
func outputMessagesJSON(messages []models.Message, userMap map[string]models.User) error {
	// Create export structure
	export := struct {
		Messages []models.Message       `json:"messages"`
		Users    map[string]models.User `json:"users"`
		Count    int                    `json:"count"`
	}{
		Messages: messages,
		Users:    userMap,
		Count:    len(messages),
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal messages to JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// outputMessagesText outputs messages in human-readable text format
func outputMessagesText(messages []models.Message, userMap map[string]models.User, verbose, noFormat bool) error {
	fmt.Printf("\n📝 Found %d messages:\n\n", len(messages))

	for _, msg := range messages {
		if err := displayMessage(msg, userMap, verbose, noFormat, 0); err != nil {
			return err
		}
	}

	return nil
}

// displayMessage displays a single message with proper formatting
func displayMessage(msg models.Message, userMap map[string]models.User, verbose, noFormat bool, indent int) error {
	// Get user info
	userName := msg.User
	if user, exists := userMap[msg.User]; exists {
		if user.Profile.DisplayName != "" {
			userName = user.Profile.DisplayName
		} else if user.RealName != "" {
			userName = user.RealName
		} else {
			userName = user.Name
		}
	}

	// Parse timestamp
	timestamp, err := strconv.ParseFloat(msg.Timestamp, 64)
	if err != nil {
		timestamp = 0
	}
	timeStr := time.Unix(int64(timestamp), 0).Format("2006-01-02 15:04:05")

	// Create indentation for threads
	indentStr := strings.Repeat("  ", indent)

	// Format message header
	if !noFormat {
		if indent > 0 {
			fmt.Printf("%s↳ ", indentStr)
		}
		fmt.Printf("👤 \033[1m%s\033[0m", userName)
		if verbose {
			fmt.Printf(" (\033[90m%s\033[0m)", msg.User)
		}
		fmt.Printf(" \033[90m%s\033[0m", timeStr)
		if msg.Edited != nil {
			fmt.Printf(" \033[93m(edited)\033[0m")
		}
		fmt.Println()
	} else {
		if indent > 0 {
			fmt.Printf("%s> ", indentStr)
		}
		fmt.Printf("[%s] %s", timeStr, userName)
		if verbose {
			fmt.Printf(" (%s)", msg.User)
		}
		if msg.Edited != nil {
			fmt.Printf(" (edited)")
		}
		fmt.Println()
	}

	// Format message text
	text := msg.Text
	if text == "" && len(msg.Attachments) > 0 {
		text = "[Attachment]"
	}
	if text == "" && len(msg.Files) > 0 {
		text = "[File]"
	}
	if text == "" {
		text = "[No text content]"
	}

	// Display message text with indentation
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		fmt.Printf("%s  %s\n", indentStr, line)
	}

	// Display attachments if verbose
	if verbose && len(msg.Attachments) > 0 {
		for _, att := range msg.Attachments {
			fmt.Printf("%s  📎 %s\n", indentStr, att.Title)
			if att.Text != "" {
				fmt.Printf("%s     %s\n", indentStr, att.Text)
			}
		}
	}

	// Display files if verbose
	if verbose && len(msg.Files) > 0 {
		for _, file := range msg.Files {
			fmt.Printf("%s  📁 %s (%s)\n", indentStr, file.Name, file.Filetype)
		}
	}

	// Display reactions if verbose
	if verbose && len(msg.Reactions) > 0 {
		var reactions []string
		for _, reaction := range msg.Reactions {
			reactions = append(reactions, fmt.Sprintf(":%s: %d", reaction.Name, reaction.Count))
		}
		fmt.Printf("%s  👍 %s\n", indentStr, strings.Join(reactions, " "))
	}

	// Display thread replies
	if len(msg.Thread) > 0 {
		fmt.Printf("%s  💬 %d replies:\n", indentStr, len(msg.Thread))
		for _, reply := range msg.Thread {
			if err := displayMessage(reply, userMap, verbose, noFormat, indent+1); err != nil {
				return err
			}
		}
	}

	fmt.Println()
	return nil
}
