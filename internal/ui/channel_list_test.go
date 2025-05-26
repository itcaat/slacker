package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itcaat/slacker/models"
)

func TestChannelListModel_SetChannels(t *testing.T) {
	model := NewChannelListModel()

	channels := []models.Channel{
		{ID: "C1", Name: "general", IsPrivate: false, IsArchived: false, NumMembers: 10},
		{ID: "C2", Name: "random", IsPrivate: false, IsArchived: false, NumMembers: 5},
		{ID: "C3", Name: "private-channel", IsPrivate: true, IsArchived: false, NumMembers: 3},
		{ID: "C4", Name: "archived-channel", IsPrivate: false, IsArchived: true, NumMembers: 0},
	}

	model.SetChannels(channels)

	if len(model.channels) != 4 {
		t.Errorf("Expected 4 channels, got %d", len(model.channels))
	}

	if model.cursor != 0 {
		t.Errorf("Expected cursor to be reset to 0, got %d", model.cursor)
	}

	if model.scrollTop != 0 {
		t.Errorf("Expected scrollTop to be reset to 0, got %d", model.scrollTop)
	}
}

func TestChannelListModel_Navigation(t *testing.T) {
	model := NewChannelListModel()
	model.SetSize(50, 20)

	channels := []models.Channel{
		{ID: "C1", Name: "channel1", IsPrivate: false, IsArchived: false},
		{ID: "C2", Name: "channel2", IsPrivate: false, IsArchived: false},
		{ID: "C3", Name: "channel3", IsPrivate: false, IsArchived: false},
	}
	model.SetChannels(channels)

	// Test down navigation
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	model = updatedModel.(*ChannelListModel)
	if model.cursor != 1 {
		t.Errorf("Expected cursor to be 1 after down navigation, got %d", model.cursor)
	}

	// Test up navigation
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	model = updatedModel.(*ChannelListModel)
	if model.cursor != 0 {
		t.Errorf("Expected cursor to be 0 after up navigation, got %d", model.cursor)
	}

	// Test boundary - up from first item should stay at 0
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	model = updatedModel.(*ChannelListModel)
	if model.cursor != 0 {
		t.Errorf("Expected cursor to stay at 0 when navigating up from first item, got %d", model.cursor)
	}

	// Test end key
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnd})
	model = updatedModel.(*ChannelListModel)
	if model.cursor != 2 {
		t.Errorf("Expected cursor to be 2 after end key, got %d", model.cursor)
	}

	// Test boundary - down from last item should stay at last
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	model = updatedModel.(*ChannelListModel)
	if model.cursor != 2 {
		t.Errorf("Expected cursor to stay at 2 when navigating down from last item, got %d", model.cursor)
	}
}

func TestChannelListModel_Selection(t *testing.T) {
	model := NewChannelListModel()

	channels := []models.Channel{
		{ID: "C1", Name: "general", IsPrivate: false, IsArchived: false},
		{ID: "C2", Name: "random", IsPrivate: false, IsArchived: false},
	}
	model.SetChannels(channels)

	// Test selection with enter key
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("Expected command to be returned when selecting channel")
	}

	// Execute the command to get the message
	msg := cmd()
	if channelMsg, ok := msg.(channelSelectedMsg); ok {
		if channelMsg.channel.ID != "C1" {
			t.Errorf("Expected selected channel ID to be C1, got %s", channelMsg.channel.ID)
		}
		if channelMsg.channel.Name != "general" {
			t.Errorf("Expected selected channel name to be general, got %s", channelMsg.channel.Name)
		}
	} else {
		t.Error("Expected channelSelectedMsg, got different message type")
	}
}

func TestChannelListModel_GetSelectedChannel(t *testing.T) {
	model := NewChannelListModel()

	// Test with no channels
	selected := model.GetSelectedChannel()
	if selected != nil {
		t.Error("Expected nil when no channels are available")
	}

	// Test with channels
	channels := []models.Channel{
		{ID: "C1", Name: "general", IsPrivate: false, IsArchived: false},
		{ID: "C2", Name: "random", IsPrivate: false, IsArchived: false},
	}
	model.SetChannels(channels)

	selected = model.GetSelectedChannel()
	if selected == nil {
		t.Error("Expected channel to be selected")
	} else if selected.ID != "C1" {
		t.Errorf("Expected selected channel ID to be C1, got %s", selected.ID)
	}

	// Move cursor and test again
	model.cursor = 1
	selected = model.GetSelectedChannel()
	if selected == nil {
		t.Error("Expected channel to be selected")
	} else if selected.ID != "C2" {
		t.Errorf("Expected selected channel ID to be C2, got %s", selected.ID)
	}
}

func TestChannelListModel_View(t *testing.T) {
	model := NewChannelListModel()
	model.SetSize(50, 10)

	// Test empty view
	view := model.View()
	if view == "" {
		t.Error("Expected non-empty view even with no channels")
	}

	// Test with channels
	channels := []models.Channel{
		{ID: "C1", Name: "general", IsPrivate: false, IsArchived: false, NumMembers: 10},
		{ID: "C2", Name: "private-channel", IsPrivate: true, IsArchived: false, NumMembers: 3},
		{ID: "C3", Name: "archived-channel", IsPrivate: false, IsArchived: true, NumMembers: 0},
	}
	model.SetChannels(channels)

	view = model.View()
	if view == "" {
		t.Error("Expected non-empty view with channels")
	}

	// Check that channel names appear in the view
	if !contains(view, "general") {
		t.Error("Expected 'general' to appear in view")
	}
	if !contains(view, "private-channel") {
		t.Error("Expected 'private-channel' to appear in view")
	}
	if !contains(view, "archived-channel") {
		t.Error("Expected 'archived-channel' to appear in view")
	}

	// Check that icons appear
	if !contains(view, "#") {
		t.Error("Expected public channel icon '#' to appear in view")
	}
	if !contains(view, "🔒") {
		t.Error("Expected private channel icon '🔒' to appear in view")
	}
	if !contains(view, "📦") {
		t.Error("Expected archived channel icon '📦' to appear in view")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsAt(s, substr, 1)))
}

func containsAt(s, substr string, start int) bool {
	if start >= len(s) {
		return false
	}
	if start+len(substr) > len(s) {
		return containsAt(s, substr, start+1)
	}
	if s[start:start+len(substr)] == substr {
		return true
	}
	return containsAt(s, substr, start+1)
}
