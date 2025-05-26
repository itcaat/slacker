package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/itcaat/slacker/models"
)

// MockSlackClient implements a mock Slack client for testing
type MockSlackClient struct {
	channels []models.Channel
	messages []models.Message
	users    []models.User
	threads  map[string][]models.Message
}

func NewMockSlackClient() *MockSlackClient {
	return &MockSlackClient{
		channels: []models.Channel{
			{
				ID:         "C123456",
				Name:       "general",
				IsPrivate:  false,
				IsArchived: false,
				NumMembers: 10,
				Created:    1704067200,
				Creator:    "U123456",
				Topic: models.Topic{
					Value:   "General discussion",
					Creator: "U123456",
					LastSet: 1704067200,
				},
				Purpose: models.Topic{
					Value:   "Company-wide announcements and general discussion",
					Creator: "U123456",
					LastSet: 1704067200,
				},
			},
		},
		messages: []models.Message{
			{
				Type:       "message",
				User:       "U123456",
				Text:       "Hello everyone!",
				Timestamp:  "1704067200.123456",
				ThreadTS:   "",
				ReplyCount: 0,
				Attachments: []models.Attachment{
					{
						ID:       1,
						Title:    "Test Attachment",
						Text:     "This is a test attachment",
						Fallback: "Test attachment fallback",
						Color:    "good",
					},
				},
				Reactions: []models.Reaction{
					{
						Name:  "thumbsup",
						Count: 3,
						Users: []string{"U123456", "U789012", "U345678"},
					},
				},
			},
			{
				Type:       "message",
				User:       "U789012",
				Text:       "How is everyone doing?",
				Timestamp:  "1704067260.000000",
				ThreadTS:   "1704067260.000000",
				ReplyCount: 2,
			},
		},
		users: []models.User{
			{
				ID:       "U123456",
				Name:     "alice",
				RealName: "Alice Smith",
				IsBot:    false,
				Deleted:  false,
				Profile: models.Profile{
					DisplayName: "Alice",
					RealName:    "Alice Smith",
					Email:       "alice@example.com",
					Image24:     "https://example.com/alice24.png",
					Image32:     "https://example.com/alice32.png",
					Image48:     "https://example.com/alice48.png",
					Image72:     "https://example.com/alice72.png",
					Image192:    "https://example.com/alice192.png",
					Image512:    "https://example.com/alice512.png",
				},
			},
			{
				ID:       "U789012",
				Name:     "bob",
				RealName: "Bob Jones",
				IsBot:    false,
				Deleted:  false,
				Profile: models.Profile{
					DisplayName: "Bob",
					RealName:    "Bob Jones",
					Email:       "bob@example.com",
				},
			},
		},
		threads: map[string][]models.Message{
			"1704067260.000000": {
				{
					Type:      "message",
					User:      "U123456",
					Text:      "I'm doing great, thanks for asking!",
					Timestamp: "1704067280.000000",
				},
				{
					Type:      "message",
					User:      "U345678",
					Text:      "Same here, having a productive day!",
					Timestamp: "1704067300.000000",
				},
			},
		},
	}
}

func (m *MockSlackClient) GetChannels(ctx context.Context) ([]models.Channel, error) {
	return m.channels, nil
}

func (m *MockSlackClient) GetChannelHistory(ctx context.Context, channelID string, limit int, cursor string) ([]models.Message, string, error) {
	// Simple implementation - return all messages for the first call
	if cursor == "" {
		return m.messages, "", nil
	}
	return []models.Message{}, "", nil
}

func (m *MockSlackClient) GetThreadReplies(ctx context.Context, channelID, threadTS string) ([]models.Message, error) {
	if replies, exists := m.threads[threadTS]; exists {
		return replies, nil
	}
	return []models.Message{}, nil
}

func (m *MockSlackClient) GetUsers(ctx context.Context) ([]models.User, error) {
	return m.users, nil
}

func TestExportService_fetchChannelInfo(t *testing.T) {
	mockClient := NewMockSlackClient()
	service := NewExportService(mockClient, "1.0.0-test")

	channel, err := service.fetchChannelInfo("C123456")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if channel == nil {
		t.Error("Expected channel to be found")
		return
	}

	if channel.ID != "C123456" {
		t.Errorf("Expected channel ID C123456, got %s", channel.ID)
	}

	if channel.Name != "general" {
		t.Errorf("Expected channel name 'general', got %s", channel.Name)
	}

	// Test non-existent channel
	_, err = service.fetchChannelInfo("C999999")
	if err == nil {
		t.Error("Expected error for non-existent channel")
	}
}

func TestExportService_fetchAllMessages(t *testing.T) {
	mockClient := NewMockSlackClient()
	service := NewExportService(mockClient, "1.0.0-test")

	options := models.ExportOptions{
		ChannelID: "C123456",
	}

	progress := models.ExportProgress{}
	messages, err := service.fetchAllMessages(options, &progress, nil, time.Now())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	// Verify messages are sorted by timestamp
	if messages[0].Timestamp > messages[1].Timestamp {
		t.Error("Expected messages to be sorted by timestamp")
	}
}

func TestExportService_fetchThreadReplies(t *testing.T) {
	mockClient := NewMockSlackClient()
	service := NewExportService(mockClient, "1.0.0-test")

	// Create messages with a thread
	messages := []models.Message{
		{
			Type:       "message",
			User:       "U789012",
			Text:       "How is everyone doing?",
			Timestamp:  "1704067260.000000",
			ThreadTS:   "1704067260.000000",
			ReplyCount: 2,
		},
	}

	progress := models.ExportProgress{}
	err := service.fetchThreadReplies(messages, "C123456", &progress, nil, time.Now())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check that thread replies were added
	if len(messages[0].Thread) != 2 {
		t.Errorf("Expected 2 thread replies, got %d", len(messages[0].Thread))
	}

	// Verify thread replies are sorted by timestamp
	if len(messages[0].Thread) > 1 {
		if messages[0].Thread[0].Timestamp > messages[0].Thread[1].Timestamp {
			t.Error("Expected thread replies to be sorted by timestamp")
		}
	}
}

func TestExportService_fetchUserInfo(t *testing.T) {
	mockClient := NewMockSlackClient()
	service := NewExportService(mockClient, "1.0.0-test")

	messages := []models.Message{
		{User: "U123456", Text: "Hello"},
		{User: "U789012", Text: "Hi there"},
		{User: "U999999", Text: "Unknown user"}, // This user doesn't exist in mock
	}

	users, err := service.fetchUserInfo(messages)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}

	// Check known users
	if user, exists := users["U123456"]; exists {
		if user.Name != "alice" {
			t.Errorf("Expected user name 'alice', got %s", user.Name)
		}
	} else {
		t.Error("Expected user U123456 to exist")
	}

	// Check placeholder user for unknown user
	if user, exists := users["U999999"]; exists {
		if !user.Deleted {
			t.Error("Expected unknown user to be marked as deleted")
		}
		if user.RealName != "Unknown User" {
			t.Errorf("Expected unknown user real name 'Unknown User', got %s", user.RealName)
		}
	} else {
		t.Error("Expected placeholder user U999999 to exist")
	}
}

func TestExportService_filterMessagesByDate(t *testing.T) {
	service := NewExportService(nil, "1.0.0-test")

	messages := []models.Message{
		{Timestamp: "1704067200.000000"}, // 2024-01-01 00:00:00
		{Timestamp: "1704153600.000000"}, // 2024-01-02 00:00:00
		{Timestamp: "1704240000.000000"}, // 2024-01-03 00:00:00
	}

	// Test with no date filters
	filtered := service.filterMessagesByDate(messages, nil, nil)
	if len(filtered) != 3 {
		t.Errorf("Expected 3 messages with no filters, got %d", len(filtered))
	}

	// Test with from date
	fromDate := time.Unix(1704153600, 0) // 2024-01-02
	filtered = service.filterMessagesByDate(messages, &fromDate, nil)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 messages with from date, got %d", len(filtered))
	}

	// Test with to date
	toDate := time.Unix(1704153600, 0) // 2024-01-02
	filtered = service.filterMessagesByDate(messages, nil, &toDate)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 messages with to date, got %d", len(filtered))
	}

	// Test with both dates
	filtered = service.filterMessagesByDate(messages, &fromDate, &toDate)
	if len(filtered) != 1 {
		t.Errorf("Expected 1 message with both dates, got %d", len(filtered))
	}
}

func TestExportService_calculateStatistics(t *testing.T) {
	service := NewExportService(nil, "1.0.0-test")

	messages := []models.Message{
		{
			User:      "U123456",
			Timestamp: "1704067200.000000",
			Attachments: []models.Attachment{
				{ID: 1, Title: "Attachment 1"},
			},
			Files: []models.File{
				{ID: "F123", Name: "file1.txt"},
			},
			Reactions: []models.Reaction{
				{Name: "thumbsup", Count: 3},
				{Name: "heart", Count: 2},
			},
			Thread: []models.Message{
				{
					User:      "U789012",
					Timestamp: "1704067260.000000",
					Reactions: []models.Reaction{
						{Name: "thumbsup", Count: 1},
					},
				},
			},
		},
		{
			User:      "U789012",
			Timestamp: "1704067300.000000",
		},
	}

	users := map[string]models.User{
		"U123456": {ID: "U123456", Name: "alice"},
		"U789012": {ID: "U789012", Name: "bob"},
	}

	stats := service.calculateStatistics(messages, users)

	// Check basic counts
	if stats.TotalMessages != 3 { // 2 main messages + 1 thread reply
		t.Errorf("Expected 3 total messages, got %d", stats.TotalMessages)
	}

	if stats.TotalThreads != 1 {
		t.Errorf("Expected 1 thread, got %d", stats.TotalThreads)
	}

	if stats.TotalReplies != 1 {
		t.Errorf("Expected 1 reply, got %d", stats.TotalReplies)
	}

	if stats.TotalUsers != 2 {
		t.Errorf("Expected 2 users, got %d", stats.TotalUsers)
	}

	if stats.TotalAttachments != 1 {
		t.Errorf("Expected 1 attachment, got %d", stats.TotalAttachments)
	}

	if stats.TotalFiles != 1 {
		t.Errorf("Expected 1 file, got %d", stats.TotalFiles)
	}

	if stats.TotalReactions != 6 { // 3 + 2 + 1
		t.Errorf("Expected 6 total reactions, got %d", stats.TotalReactions)
	}

	// Check messages by user
	if stats.MessagesByUser["U123456"] != 1 {
		t.Errorf("Expected 1 message from U123456, got %d", stats.MessagesByUser["U123456"])
	}

	if stats.MessagesByUser["U789012"] != 2 { // 1 main message + 1 thread reply
		t.Errorf("Expected 2 messages from U789012, got %d", stats.MessagesByUser["U789012"])
	}

	// Check top reactions
	if len(stats.TopReactions) == 0 {
		t.Error("Expected top reactions to be calculated")
	} else {
		// Should be sorted by count, so thumbsup (4 total) should be first
		if stats.TopReactions[0].Name != "thumbsup" {
			t.Errorf("Expected top reaction to be 'thumbsup', got %s", stats.TopReactions[0].Name)
		}
		if stats.TopReactions[0].Count != 4 {
			t.Errorf("Expected top reaction count to be 4, got %d", stats.TopReactions[0].Count)
		}
	}
}

func TestExportService_processExportData(t *testing.T) {
	mockClient := NewMockSlackClient()
	service := NewExportService(mockClient, "1.0.0-test")

	channel := &models.Channel{
		ID:         "C123456",
		Name:       "general",
		IsPrivate:  false,
		IsArchived: false,
		NumMembers: 10,
		Created:    1704067200,
		Creator:    "U123456",
		Topic: models.Topic{
			Value: "General discussion",
		},
	}

	messages := []models.Message{
		{
			User:      "U123456",
			Text:      "Hello world",
			Timestamp: "1704067200.000000",
		},
	}

	users := map[string]models.User{
		"U123456": {
			ID:       "U123456",
			Name:     "alice",
			RealName: "Alice Smith",
		},
	}

	options := models.ExportOptions{
		ChannelID:      "C123456",
		IncludeThreads: true,
		Format:         "json-pretty",
	}

	exportData, statistics := service.processExportData(channel, messages, users, options, time.Now())

	// Check export metadata
	if exportData.ExportInfo.SlackerVersion != "1.0.0-test" {
		t.Errorf("Expected version '1.0.0-test', got %s", exportData.ExportInfo.SlackerVersion)
	}

	if exportData.ExportInfo.ExportFormat != "json-pretty" {
		t.Errorf("Expected format 'json-pretty', got %s", exportData.ExportInfo.ExportFormat)
	}

	if !exportData.ExportInfo.IncludeThreads {
		t.Error("Expected IncludeThreads to be true")
	}

	// Check channel info
	if exportData.Channel.ID != "C123456" {
		t.Errorf("Expected channel ID 'C123456', got %s", exportData.Channel.ID)
	}

	if exportData.Channel.Topic != "General discussion" {
		t.Errorf("Expected topic 'General discussion', got %s", exportData.Channel.Topic)
	}

	// Check messages
	if len(exportData.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(exportData.Messages))
	}

	// Check users
	if len(exportData.Users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(exportData.Users))
	}

	// Check statistics
	if statistics.TotalMessages != 1 {
		t.Errorf("Expected 1 total message in statistics, got %d", statistics.TotalMessages)
	}

	if statistics.TotalUsers != 1 {
		t.Errorf("Expected 1 total user in statistics, got %d", statistics.TotalUsers)
	}
}
