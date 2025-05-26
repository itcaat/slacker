package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/itcaat/slacker/models"
)

// ChannelListModel represents the channel list component
type ChannelListModel struct {
	channels  []models.Channel
	cursor    int
	width     int
	height    int
	viewport  int
	scrollTop int
	styles    ChannelListStyles
}

// ChannelListStyles contains styling for the channel list
type ChannelListStyles struct {
	Selected   lipgloss.Style
	Unselected lipgloss.Style
	Public     lipgloss.Style
	Private    lipgloss.Style
	Archived   lipgloss.Style
}

// NewChannelListModel creates a new channel list model
func NewChannelListModel() *ChannelListModel {
	return &ChannelListModel{
		channels: []models.Channel{},
		cursor:   0,
		styles:   createChannelListStyles(),
	}
}

// createChannelListStyles initializes the channel list styles
func createChannelListStyles() ChannelListStyles {
	return ChannelListStyles{
		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Bold(true).
			Padding(0, 1),

		Unselected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A49FA5")).
			Padding(0, 1),

		Public: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")),

		Private: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF8C00")),

		Archived: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Strikethrough(true),
	}
}

// SetChannels sets the channels to display
func (m *ChannelListModel) SetChannels(channels []models.Channel) {
	m.channels = channels
	m.cursor = 0
	m.scrollTop = 0
}

// SetSize sets the size of the channel list
func (m *ChannelListModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport = height - 2 // Account for borders
}

// Init implements tea.Model
func (m *ChannelListModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *ChannelListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.adjustScroll()
			}
		case "down", "j":
			if m.cursor < len(m.channels)-1 {
				m.cursor++
				m.adjustScroll()
			}
		case "enter", " ":
			if len(m.channels) > 0 && m.cursor < len(m.channels) {
				return m, func() tea.Msg {
					return channelSelectedMsg{channel: m.channels[m.cursor]}
				}
			}
		case "home":
			m.cursor = 0
			m.scrollTop = 0
		case "end":
			m.cursor = len(m.channels) - 1
			m.adjustScroll()
		}
	}
	return m, nil
}

// adjustScroll adjusts the scroll position to keep the cursor visible
func (m *ChannelListModel) adjustScroll() {
	if m.viewport <= 0 {
		return
	}

	// Scroll down if cursor is below viewport
	if m.cursor >= m.scrollTop+m.viewport {
		m.scrollTop = m.cursor - m.viewport + 1
	}

	// Scroll up if cursor is above viewport
	if m.cursor < m.scrollTop {
		m.scrollTop = m.cursor
	}

	// Ensure scroll doesn't go negative
	if m.scrollTop < 0 {
		m.scrollTop = 0
	}
}

// View implements tea.Model
func (m *ChannelListModel) View() string {
	if len(m.channels) == 0 {
		return m.styles.Unselected.Render("No channels available")
	}

	var items []string

	// Calculate visible range
	start := m.scrollTop
	end := start + m.viewport
	if end > len(m.channels) {
		end = len(m.channels)
	}

	for i := start; i < end; i++ {
		channel := m.channels[i]

		// Channel icon and name
		var icon string
		var nameStyle lipgloss.Style

		if channel.IsArchived {
			icon = "📦"
			nameStyle = m.styles.Archived
		} else if channel.IsPrivate {
			icon = "🔒"
			nameStyle = m.styles.Private
		} else {
			icon = "#"
			nameStyle = m.styles.Public
		}

		// Format channel name
		name := nameStyle.Render(channel.Name)

		// Add member count if available
		memberInfo := ""
		if channel.NumMembers > 0 {
			memberInfo = fmt.Sprintf(" (%d)", channel.NumMembers)
		}

		// Create the full item text
		itemText := fmt.Sprintf("%s %s%s", icon, name, memberInfo)

		// Apply selection styling
		if i == m.cursor {
			itemText = m.styles.Selected.Width(m.width - 4).Render(itemText)
		} else {
			itemText = m.styles.Unselected.Width(m.width - 4).Render(itemText)
		}

		items = append(items, itemText)
	}

	// Add scroll indicators if needed
	content := strings.Join(items, "\n")

	// Add scroll indicators
	if m.scrollTop > 0 {
		content = "↑ More above\n" + content
	}
	if m.scrollTop+m.viewport < len(m.channels) {
		content = content + "\n↓ More below"
	}

	return content
}

// GetSelectedChannel returns the currently selected channel
func (m *ChannelListModel) GetSelectedChannel() *models.Channel {
	if len(m.channels) == 0 || m.cursor >= len(m.channels) {
		return nil
	}
	return &m.channels[m.cursor]
}

// GetChannelCount returns the total number of channels
func (m *ChannelListModel) GetChannelCount() int {
	return len(m.channels)
}
