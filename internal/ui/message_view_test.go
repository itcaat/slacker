package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itcaat/slacker/models"
)

func TestMessageViewModel_SetMessages(t *testing.T) {
	model := NewMessageViewModel()

	messages := []models.Message{
		{User: "U1", Text: "Hello world", Timestamp: "1704067200.123456"},
		{User: "U2", Text: "How are you?", Timestamp: "1704067260.123456"},
	}

	users := map[string]models.User{
		"U1": {ID: "U1", Name: "alice", RealName: "Alice Smith"},
		"U2": {ID: "U2", Name: "bob", RealName: "Bob Jones"},
	}

	model.SetMessages(messages, users)

	if len(model.messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(model.messages))
	}

	if len(model.users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(model.users))
	}

	if model.cursor != 0 {
		t.Errorf("Expected cursor to be reset to 0, got %d", model.cursor)
	}

	if model.scrollTop != 0 {
		t.Errorf("Expected scrollTop to be reset to 0, got %d", model.scrollTop)
	}
}

func TestMessageViewModel_Navigation(t *testing.T) {
	model := NewMessageViewModel()
	model.SetSize(80, 20)

	messages := []models.Message{
		{User: "U1", Text: "Message 1", Timestamp: "1704067200.123456"},
		{User: "U2", Text: "Message 2", Timestamp: "1704067260.123456"},
		{User: "U3", Text: "Message 3", Timestamp: "1704067320.123456"},
	}

	users := map[string]models.User{
		"U1": {ID: "U1", Name: "user1"},
		"U2": {ID: "U2", Name: "user2"},
		"U3": {ID: "U3", Name: "user3"},
	}

	model.SetMessages(messages, users)

	// Test down navigation
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	model = updatedModel.(*MessageViewModel)
	if model.cursor != 1 {
		t.Errorf("Expected cursor to be 1 after down navigation, got %d", model.cursor)
	}

	// Test up navigation
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	model = updatedModel.(*MessageViewModel)
	if model.cursor != 0 {
		t.Errorf("Expected cursor to be 0 after up navigation, got %d", model.cursor)
	}

	// Test end key
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnd})
	model = updatedModel.(*MessageViewModel)
	if model.cursor != 2 {
		t.Errorf("Expected cursor to be 2 after end key, got %d", model.cursor)
	}

	// Test home key
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyHome})
	model = updatedModel.(*MessageViewModel)
	if model.cursor != 0 {
		t.Errorf("Expected cursor to be 0 after home key, got %d", model.cursor)
	}
}

func TestMessageViewModel_GetSelectedMessage(t *testing.T) {
	model := NewMessageViewModel()

	// Test with no messages
	selected := model.GetSelectedMessage()
	if selected != nil {
		t.Error("Expected nil when no messages are available")
	}

	// Test with messages
	messages := []models.Message{
		{User: "U1", Text: "Message 1", Timestamp: "1704067200.123456"},
		{User: "U2", Text: "Message 2", Timestamp: "1704067260.123456"},
	}

	users := map[string]models.User{
		"U1": {ID: "U1", Name: "user1"},
		"U2": {ID: "U2", Name: "user2"},
	}

	model.SetMessages(messages, users)

	selected = model.GetSelectedMessage()
	if selected == nil {
		t.Error("Expected message to be selected")
	} else if selected.Text != "Message 1" {
		t.Errorf("Expected selected message text to be 'Message 1', got '%s'", selected.Text)
	}

	// Move cursor and test again
	model.cursor = 1
	selected = model.GetSelectedMessage()
	if selected == nil {
		t.Error("Expected message to be selected")
	} else if selected.Text != "Message 2" {
		t.Errorf("Expected selected message text to be 'Message 2', got '%s'", selected.Text)
	}
}

func TestMessageViewModel_View(t *testing.T) {
	model := NewMessageViewModel()
	model.SetSize(80, 10)

	// Test empty view
	view := model.View()
	if view == "" {
		t.Error("Expected non-empty view even with no messages")
	}

	// Test with messages
	messages := []models.Message{
		{
			User:      "U1",
			Text:      "Hello everyone!",
			Timestamp: "1704067200.123456",
		},
		{
			User:      "U2",
			Text:      "How is everyone doing?",
			Timestamp: "1704067260.123456",
			Thread: []models.Message{
				{User: "U3", Text: "I'm doing great!", Timestamp: "1704067280.123456"},
			},
		},
	}

	users := map[string]models.User{
		"U1": {ID: "U1", Name: "alice", RealName: "Alice Smith"},
		"U2": {ID: "U2", Name: "bob", RealName: "Bob Jones"},
		"U3": {ID: "U3", Name: "charlie", RealName: "Charlie Brown"},
	}

	model.SetMessages(messages, users)

	view = model.View()
	if view == "" {
		t.Error("Expected non-empty view with messages")
	}

	// Check that message text appears in the view
	if !strings.Contains(view, "Hello everyone!") {
		t.Error("Expected 'Hello everyone!' to appear in view")
	}
	if !strings.Contains(view, "How is everyone doing?") {
		t.Error("Expected 'How is everyone doing?' to appear in view")
	}

	// Check that user names appear
	if !strings.Contains(view, "Alice Smith") {
		t.Error("Expected 'Alice Smith' to appear in view")
	}
	if !strings.Contains(view, "Bob Jones") {
		t.Error("Expected 'Bob Jones' to appear in view")
	}

	// Check that thread indicator appears
	if !strings.Contains(view, "💬") {
		t.Error("Expected thread indicator '💬' to appear in view")
	}

	// Check that user icon appears
	if !strings.Contains(view, "👤") {
		t.Error("Expected user icon '👤' to appear in view")
	}
}

func TestMessageViewModel_WrapText(t *testing.T) {
	model := NewMessageViewModel()

	// Test short text (no wrapping needed)
	text := "Short text"
	wrapped := model.wrapText(text, 50)
	if wrapped != text {
		t.Errorf("Expected short text to remain unchanged, got '%s'", wrapped)
	}

	// Test long text (wrapping needed)
	longText := "This is a very long text that should be wrapped when it exceeds the specified width limit"
	wrapped = model.wrapText(longText, 20)
	lines := strings.Split(wrapped, "\n")
	if len(lines) <= 1 {
		t.Error("Expected long text to be wrapped into multiple lines")
	}

	// Check that no line exceeds the width (approximately)
	for _, line := range lines {
		if len(line) > 25 { // Allow some tolerance
			t.Errorf("Line too long: '%s' (length: %d)", line, len(line))
		}
	}

	// Test edge case: zero width
	wrapped = model.wrapText("test", 0)
	if wrapped != "test" {
		t.Errorf("Expected text to remain unchanged with zero width, got '%s'", wrapped)
	}
}

func TestMessageViewModel_FormatMessage(t *testing.T) {
	model := NewMessageViewModel()
	model.SetSize(80, 20)

	users := map[string]models.User{
		"U1": {ID: "U1", Name: "alice", RealName: "Alice Smith"},
	}
	model.users = users

	message := models.Message{
		User:      "U1",
		Text:      "Test message",
		Timestamp: "1704067200.123456",
		Attachments: []models.Attachment{
			{Title: "test.pdf"},
		},
		Reactions: []models.Reaction{
			{Name: "thumbsup", Count: 3},
		},
	}

	formatted := model.formatMessage(message, 0, false)

	// Check that user name appears
	if !strings.Contains(formatted, "Alice Smith") {
		t.Error("Expected user name 'Alice Smith' to appear in formatted message")
	}

	// Check that message text appears
	if !strings.Contains(formatted, "Test message") {
		t.Error("Expected message text 'Test message' to appear in formatted message")
	}

	// Check that attachment appears
	if !strings.Contains(formatted, "📎") {
		t.Error("Expected attachment icon '📎' to appear in formatted message")
	}

	// Check that reaction appears
	if !strings.Contains(formatted, "👍") {
		t.Error("Expected reaction icon '👍' to appear in formatted message")
	}
	if !strings.Contains(formatted, "thumbsup") {
		t.Error("Expected reaction name 'thumbsup' to appear in formatted message")
	}
}
