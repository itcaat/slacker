package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/itcaat/slacker/internal/api"
	"github.com/itcaat/slacker/models"
)

// MessageService handles message-related business logic
type MessageService struct {
	slackClient *api.SlackClient
}

// NewMessageService creates a new message service
func NewMessageService(slackClient *api.SlackClient) *MessageService {
	return &MessageService{
		slackClient: slackClient,
	}
}

// MessageRetrievalOptions defines options for message retrieval
type MessageRetrievalOptions struct {
	ChannelID      string
	ChannelName    string
	Limit          int
	IncludeThreads bool
	Before         string
	After          string
	IncludeUsers   bool
}

// MessageResult contains the result of message retrieval
type MessageResult struct {
	Messages []models.Message
	Users    map[string]models.User
	Channel  *models.Channel
	Count    int
}

// GetChannelMessages retrieves messages from a channel with the specified options
func (ms *MessageService) GetChannelMessages(ctx context.Context, opts MessageRetrievalOptions) (*MessageResult, error) {
	var channel *models.Channel
	var err error

	// Find channel by name or use provided ID
	if opts.ChannelName != "" {
		channel, err = ms.slackClient.GetChannelByName(ctx, opts.ChannelName)
		if err != nil {
			return nil, fmt.Errorf("failed to find channel '%s': %w", opts.ChannelName, err)
		}
	} else if opts.ChannelID != "" {
		// For now, we'll need to get all channels to find the one with this ID
		// In a real implementation, we might want to add a GetChannelByID method
		channels, err := ms.slackClient.GetChannels(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get channels: %w", err)
		}

		for _, ch := range channels {
			if ch.ID == opts.ChannelID {
				channel = &ch
				break
			}
		}

		if channel == nil {
			return nil, fmt.Errorf("channel with ID '%s' not found", opts.ChannelID)
		}
	} else {
		return nil, fmt.Errorf("either channel name or channel ID must be provided")
	}

	// Set default limit if not specified
	if opts.Limit <= 0 {
		opts.Limit = 100
	}

	// Retrieve messages with pagination
	messages, err := ms.retrieveMessagesWithPagination(ctx, channel.ID, opts.Limit, opts.Before, opts.After)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve messages: %w", err)
	}

	// Enrich with thread replies if requested
	if opts.IncludeThreads {
		messages, err = ms.enrichWithThreadReplies(ctx, channel.ID, messages)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve thread replies: %w", err)
		}
	}

	// Get user information if requested
	var userMap map[string]models.User
	if opts.IncludeUsers {
		users, err := ms.slackClient.GetUsers(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get users: %w", err)
		}

		userMap = make(map[string]models.User)
		for _, user := range users {
			userMap[user.ID] = user
		}
	}

	return &MessageResult{
		Messages: messages,
		Users:    userMap,
		Channel:  channel,
		Count:    len(messages),
	}, nil
}

// GetAllChannelHistory retrieves complete history for a channel (for export)
func (ms *MessageService) GetAllChannelHistory(ctx context.Context, channelID string, includeThreads bool) (*MessageResult, error) {
	// Get all messages without limit
	var allMessages []models.Message
	cursor := ""

	for {
		messages, nextCursor, err := ms.slackClient.GetChannelHistory(ctx, channelID, 200, cursor)
		if err != nil {
			return nil, fmt.Errorf("failed to get channel history: %w", err)
		}

		allMessages = append(allMessages, messages...)

		if nextCursor == "" || len(messages) == 0 {
			break
		}

		cursor = nextCursor

		// Add a small delay to respect rate limits
		time.Sleep(100 * time.Millisecond)
	}

	// Enrich with thread replies if requested
	if includeThreads {
		var err error
		allMessages, err = ms.enrichWithThreadReplies(ctx, channelID, allMessages)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve thread replies: %w", err)
		}
	}

	// Get user information
	users, err := ms.slackClient.GetUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	userMap := make(map[string]models.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// Get channel information
	channels, err := ms.slackClient.GetChannels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}

	var channel *models.Channel
	for _, ch := range channels {
		if ch.ID == channelID {
			channel = &ch
			break
		}
	}

	return &MessageResult{
		Messages: allMessages,
		Users:    userMap,
		Channel:  channel,
		Count:    len(allMessages),
	}, nil
}

// retrieveMessagesWithPagination handles paginated message retrieval
func (ms *MessageService) retrieveMessagesWithPagination(ctx context.Context, channelID string, limit int, before, after string) ([]models.Message, error) {
	var allMessages []models.Message
	cursor := ""
	remaining := limit

	for remaining > 0 {
		// Calculate batch size (max 200 per API call)
		batchSize := remaining
		if batchSize > 200 {
			batchSize = 200
		}

		messages, nextCursor, err := ms.slackClient.GetChannelHistory(ctx, channelID, batchSize, cursor)
		if err != nil {
			return nil, err
		}

		// Filter messages by time range if specified
		filteredMessages := ms.filterMessagesByTime(messages, before, after)
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

		// Add a small delay to respect rate limits
		time.Sleep(50 * time.Millisecond)
	}

	// Trim to exact limit if we got more than requested
	if len(allMessages) > limit {
		allMessages = allMessages[:limit]
	}

	return allMessages, nil
}

// enrichWithThreadReplies fetches thread replies for messages that have them
func (ms *MessageService) enrichWithThreadReplies(ctx context.Context, channelID string, messages []models.Message) ([]models.Message, error) {
	for i, msg := range messages {
		if msg.ReplyCount > 0 && msg.ThreadTS != "" {
			replies, err := ms.slackClient.GetThreadReplies(ctx, channelID, msg.ThreadTS)
			if err != nil {
				// Log error but continue with other messages
				fmt.Printf("Warning: Failed to get replies for message %s: %v\n", msg.Timestamp, err)
				continue
			}
			messages[i].Thread = replies
		}

		// Add a small delay between thread requests to respect rate limits
		if msg.ReplyCount > 0 {
			time.Sleep(50 * time.Millisecond)
		}
	}
	return messages, nil
}

// filterMessagesByTime filters messages based on before/after timestamps
func (ms *MessageService) filterMessagesByTime(messages []models.Message, before, after string) []models.Message {
	if before == "" && after == "" {
		return messages
	}

	var filtered []models.Message
	for _, msg := range messages {
		// Parse message timestamp
		msgTime, err := time.Parse("1704067200.123456", msg.Timestamp)
		if err != nil {
			// Try parsing as Unix timestamp
			if ts, err := time.Parse("1704067200", msg.Timestamp[:10]); err == nil {
				msgTime = ts
			} else {
				continue // Skip messages with invalid timestamps
			}
		}

		// Check before constraint
		if before != "" {
			beforeTime, err := time.Parse(time.RFC3339, before)
			if err == nil && msgTime.After(beforeTime) {
				continue
			}
		}

		// Check after constraint
		if after != "" {
			afterTime, err := time.Parse(time.RFC3339, after)
			if err == nil && msgTime.Before(afterTime) {
				continue
			}
		}

		filtered = append(filtered, msg)
	}

	return filtered
}

// GetChannelStats returns statistics about a channel's messages
func (ms *MessageService) GetChannelStats(ctx context.Context, channelID string) (*ChannelStats, error) {
	// Get a sample of recent messages to calculate stats
	messages, _, err := ms.slackClient.GetChannelHistory(ctx, channelID, 100, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get channel history for stats: %w", err)
	}

	stats := &ChannelStats{
		ChannelID:    channelID,
		SampleSize:   len(messages),
		MessageCount: len(messages), // This is just a sample, not total
	}

	// Calculate basic statistics
	userCounts := make(map[string]int)
	threadCount := 0

	for _, msg := range messages {
		userCounts[msg.User]++
		if msg.ReplyCount > 0 {
			threadCount++
		}
	}

	stats.UniqueUsers = len(userCounts)
	stats.ThreadCount = threadCount

	// Find most active user
	maxCount := 0
	for user, count := range userCounts {
		if count > maxCount {
			maxCount = count
			stats.MostActiveUser = user
		}
	}

	return stats, nil
}

// ChannelStats represents statistics about a channel
type ChannelStats struct {
	ChannelID      string `json:"channel_id"`
	MessageCount   int    `json:"message_count"`
	SampleSize     int    `json:"sample_size"`
	UniqueUsers    int    `json:"unique_users"`
	ThreadCount    int    `json:"thread_count"`
	MostActiveUser string `json:"most_active_user"`
}
