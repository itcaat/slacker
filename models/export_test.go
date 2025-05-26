package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParseSlackTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
		hasError bool
	}{
		{
			name:     "Valid timestamp",
			input:    "1704067200.123456",
			expected: time.Unix(1704067200, 123456000),
			hasError: false,
		},
		{
			name:     "Empty timestamp",
			input:    "",
			expected: time.Time{},
			hasError: false,
		},
		{
			name:     "Invalid timestamp",
			input:    "invalid",
			expected: time.Time{},
			hasError: true,
		},
		{
			name:     "Integer timestamp",
			input:    "1704067200",
			expected: time.Unix(1704067200, 0),
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSlackTimestamp(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for input %s, but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for input %s: %v", tt.input, err)
				return
			}

			// Allow for small nanosecond differences due to floating point precision
			diff := result.Sub(tt.expected)
			if diff < 0 {
				diff = -diff
			}
			if diff > time.Microsecond {
				t.Errorf("Expected %v, got %v (diff: %v)", tt.expected, result, diff)
			}
		})
	}
}

func TestConvertToExportMessage(t *testing.T) {
	// Create a test message
	msg := Message{
		Type:       "message",
		User:       "U123456",
		Text:       "Hello, world!",
		Timestamp:  "1704067200.123456",
		ThreadTS:   "1704067100.000000",
		ReplyCount: 2,
		Subtype:    "",
		Attachments: []Attachment{
			{
				ID:       1,
				Title:    "Test Attachment",
				Text:     "Attachment text",
				Fallback: "Fallback text",
				Color:    "good",
				ImageURL: "https://example.com/image.png",
				ThumbURL: "https://example.com/thumb.png",
			},
		},
		Files: []File{
			{
				ID:       "F123456",
				Name:     "test.txt",
				Title:    "Test File",
				Mimetype: "text/plain",
				Filetype: "txt",
				Size:     1024,
				URL:      "https://example.com/file.txt",
			},
		},
		Reactions: []Reaction{
			{
				Name:  "thumbsup",
				Count: 3,
				Users: []string{"U123", "U456", "U789"},
			},
		},
		Edited: &Edited{
			User:      "U123456",
			Timestamp: "1704067300.000000",
		},
		Thread: []Message{
			{
				Type:      "message",
				User:      "U789012",
				Text:      "Reply message",
				Timestamp: "1704067250.000000",
			},
		},
	}

	// Convert to export format
	exportMsg := ConvertToExportMessage(msg)

	// Verify basic fields
	if exportMsg.ID != msg.Timestamp {
		t.Errorf("Expected ID %s, got %s", msg.Timestamp, exportMsg.ID)
	}

	if exportMsg.User != msg.User {
		t.Errorf("Expected User %s, got %s", msg.User, exportMsg.User)
	}

	if exportMsg.Text != msg.Text {
		t.Errorf("Expected Text %s, got %s", msg.Text, exportMsg.Text)
	}

	if exportMsg.Type != msg.Type {
		t.Errorf("Expected Type %s, got %s", msg.Type, exportMsg.Type)
	}

	if exportMsg.ThreadTimestamp != msg.ThreadTS {
		t.Errorf("Expected ThreadTimestamp %s, got %s", msg.ThreadTS, exportMsg.ThreadTimestamp)
	}

	if exportMsg.ReplyCount != msg.ReplyCount {
		t.Errorf("Expected ReplyCount %d, got %d", msg.ReplyCount, exportMsg.ReplyCount)
	}

	// Verify timestamp parsing
	expectedTime := time.Unix(1704067200, 123456000)
	diff := exportMsg.Timestamp.Sub(expectedTime)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Microsecond {
		t.Errorf("Expected Timestamp %v, got %v (diff: %v)", expectedTime, exportMsg.Timestamp, diff)
	}

	// Verify edit info
	if exportMsg.Edited == nil {
		t.Error("Expected Edited info to be present")
	} else {
		if exportMsg.Edited.User != msg.Edited.User {
			t.Errorf("Expected Edited.User %s, got %s", msg.Edited.User, exportMsg.Edited.User)
		}
		expectedEditTime := time.Unix(1704067300, 0)
		if !exportMsg.Edited.Timestamp.Equal(expectedEditTime) {
			t.Errorf("Expected Edited.Timestamp %v, got %v", expectedEditTime, exportMsg.Edited.Timestamp)
		}
	}

	// Verify attachments
	if len(exportMsg.Attachments) != 1 {
		t.Errorf("Expected 1 attachment, got %d", len(exportMsg.Attachments))
	} else {
		att := exportMsg.Attachments[0]
		if att.ID != "1" {
			t.Errorf("Expected attachment ID '1', got '%s'", att.ID)
		}
		if att.Title != msg.Attachments[0].Title {
			t.Errorf("Expected attachment Title %s, got %s", msg.Attachments[0].Title, att.Title)
		}
	}

	// Verify files
	if len(exportMsg.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(exportMsg.Files))
	} else {
		file := exportMsg.Files[0]
		if file.ID != msg.Files[0].ID {
			t.Errorf("Expected file ID %s, got %s", msg.Files[0].ID, file.ID)
		}
		if file.Name != msg.Files[0].Name {
			t.Errorf("Expected file Name %s, got %s", msg.Files[0].Name, file.Name)
		}
	}

	// Verify reactions
	if len(exportMsg.Reactions) != 1 {
		t.Errorf("Expected 1 reaction, got %d", len(exportMsg.Reactions))
	} else {
		reaction := exportMsg.Reactions[0]
		if reaction.Name != msg.Reactions[0].Name {
			t.Errorf("Expected reaction Name %s, got %s", msg.Reactions[0].Name, reaction.Name)
		}
		if reaction.Count != msg.Reactions[0].Count {
			t.Errorf("Expected reaction Count %d, got %d", msg.Reactions[0].Count, reaction.Count)
		}
	}

	// Verify thread replies
	if len(exportMsg.Replies) != 1 {
		t.Errorf("Expected 1 reply, got %d", len(exportMsg.Replies))
	} else {
		reply := exportMsg.Replies[0]
		if reply.User != msg.Thread[0].User {
			t.Errorf("Expected reply User %s, got %s", msg.Thread[0].User, reply.User)
		}
		if reply.Text != msg.Thread[0].Text {
			t.Errorf("Expected reply Text %s, got %s", msg.Thread[0].Text, reply.Text)
		}
	}
}

func TestConvertToExportUser(t *testing.T) {
	// Create a test user
	user := User{
		ID:       "U123456",
		Name:     "testuser",
		RealName: "Test User",
		IsBot:    false,
		Deleted:  false,
		Profile: Profile{
			DisplayName: "Test Display Name",
			RealName:    "Test Real Name",
			Email:       "test@example.com",
			Image24:     "https://example.com/image24.png",
			Image32:     "https://example.com/image32.png",
			Image48:     "https://example.com/image48.png",
			Image72:     "https://example.com/image72.png",
			Image192:    "https://example.com/image192.png",
			Image512:    "https://example.com/image512.png",
		},
	}

	// Convert to export format
	exportUser := ConvertToExportUser(user)

	// Verify basic fields
	if exportUser.ID != user.ID {
		t.Errorf("Expected ID %s, got %s", user.ID, exportUser.ID)
	}

	if exportUser.Name != user.Name {
		t.Errorf("Expected Name %s, got %s", user.Name, exportUser.Name)
	}

	if exportUser.RealName != user.RealName {
		t.Errorf("Expected RealName %s, got %s", user.RealName, exportUser.RealName)
	}

	if exportUser.IsBot != user.IsBot {
		t.Errorf("Expected IsBot %v, got %v", user.IsBot, exportUser.IsBot)
	}

	if exportUser.Deleted != user.Deleted {
		t.Errorf("Expected Deleted %v, got %v", user.Deleted, exportUser.Deleted)
	}

	// Verify profile
	if exportUser.Profile.DisplayName != user.Profile.DisplayName {
		t.Errorf("Expected Profile.DisplayName %s, got %s", user.Profile.DisplayName, exportUser.Profile.DisplayName)
	}

	if exportUser.Profile.Email != user.Profile.Email {
		t.Errorf("Expected Profile.Email %s, got %s", user.Profile.Email, exportUser.Profile.Email)
	}

	if exportUser.Profile.Image24 != user.Profile.Image24 {
		t.Errorf("Expected Profile.Image24 %s, got %s", user.Profile.Image24, exportUser.Profile.Image24)
	}
}

func TestChannelExportJSONSerialization(t *testing.T) {
	// Create a test export structure
	export := ChannelExport{
		ExportInfo: ExportMetadata{
			ExportedAt:     time.Now(),
			ExportedBy:     "test-user",
			SlackerVersion: "1.0.0",
			ExportFormat:   "json",
			IncludeThreads: true,
		},
		Channel: ChannelInfo{
			ID:         "C123456",
			Name:       "general",
			IsPrivate:  false,
			IsArchived: false,
			NumMembers: 10,
		},
		Messages: []ExportMessage{
			{
				ID:        "1704067200.123456",
				User:      "U123456",
				Text:      "Test message",
				Timestamp: time.Unix(1704067200, 123456000),
				Type:      "message",
			},
		},
		Users: map[string]ExportUser{
			"U123456": {
				ID:       "U123456",
				Name:     "testuser",
				RealName: "Test User",
				Profile: ExportProfile{
					DisplayName: "Test Display Name",
					Email:       "test@example.com",
				},
			},
		},
		Statistics: ExportStatistics{
			TotalMessages: 1,
			TotalUsers:    1,
			MessagesByUser: map[string]int{
				"U123456": 1,
			},
		},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(export)
	if err != nil {
		t.Errorf("Failed to marshal export to JSON: %v", err)
	}

	// Test JSON deserialization
	var unmarshaled ChannelExport
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal export from JSON: %v", err)
	}

	// Verify some key fields
	if unmarshaled.Channel.ID != export.Channel.ID {
		t.Errorf("Expected Channel.ID %s, got %s", export.Channel.ID, unmarshaled.Channel.ID)
	}

	if len(unmarshaled.Messages) != len(export.Messages) {
		t.Errorf("Expected %d messages, got %d", len(export.Messages), len(unmarshaled.Messages))
	}

	if len(unmarshaled.Users) != len(export.Users) {
		t.Errorf("Expected %d users, got %d", len(export.Users), len(unmarshaled.Users))
	}
}

func TestExportOptions(t *testing.T) {
	options := ExportOptions{
		ChannelID:        "C123456",
		ChannelName:      "general",
		IncludeThreads:   true,
		IncludeFiles:     true,
		IncludeReactions: true,
		OutputFile:       "export.json",
		Format:           "json-pretty",
		Compression:      "gzip",
	}

	// Test JSON serialization of options
	jsonData, err := json.Marshal(options)
	if err != nil {
		t.Errorf("Failed to marshal options to JSON: %v", err)
	}

	// Test JSON deserialization of options
	var unmarshaled ExportOptions
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal options from JSON: %v", err)
	}

	// Verify fields
	if unmarshaled.ChannelID != options.ChannelID {
		t.Errorf("Expected ChannelID %s, got %s", options.ChannelID, unmarshaled.ChannelID)
	}

	if unmarshaled.IncludeThreads != options.IncludeThreads {
		t.Errorf("Expected IncludeThreads %v, got %v", options.IncludeThreads, unmarshaled.IncludeThreads)
	}

	if unmarshaled.Format != options.Format {
		t.Errorf("Expected Format %s, got %s", options.Format, unmarshaled.Format)
	}
}
