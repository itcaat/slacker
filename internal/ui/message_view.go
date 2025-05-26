package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/itcaat/slacker/models"
)

// MessageViewModel represents the message view component
type MessageViewModel struct {
	messages  []models.Message
	users     map[string]models.User
	cursor    int
	width     int
	height    int
	viewport  int
	scrollTop int
	styles    MessageViewStyles
}

// MessageViewStyles contains styling for the message view
type MessageViewStyles struct {
	Message    lipgloss.Style
	Thread     lipgloss.Style
	Username   lipgloss.Style
	Timestamp  lipgloss.Style
	Selected   lipgloss.Style
	Unselected lipgloss.Style
	Attachment lipgloss.Style
	Reaction   lipgloss.Style
}

// NewMessageViewModel creates a new message view model
func NewMessageViewModel() *MessageViewModel {
	return &MessageViewModel{
		messages: []models.Message{},
		users:    make(map[string]models.User),
		cursor:   0,
		styles:   createMessageViewStyles(),
	}
}

// createMessageViewStyles initializes the message view styles
func createMessageViewStyles() MessageViewStyles {
	return MessageViewStyles{
		Message: lipgloss.NewStyle().
			Padding(0, 1).
			MarginBottom(1),

		Thread: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A49FA5")).
			MarginLeft(2).
			Padding(0, 1),

		Username: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true),

		Timestamp: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")),

		Selected: lipgloss.NewStyle().
			Background(lipgloss.Color("#3C3C3C")).
			Padding(0, 1),

		Unselected: lipgloss.NewStyle().
			Padding(0, 1),

		Attachment: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF8C00")).
			Italic(true),

		Reaction: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")),
	}
}

// SetMessages sets the messages to display
func (m *MessageViewModel) SetMessages(messages []models.Message, users map[string]models.User) {
	m.messages = messages
	m.users = users
	m.cursor = 0
	m.scrollTop = 0
}

// SetSize sets the size of the message view
func (m *MessageViewModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport = height - 2 // Account for borders
}

// Init implements tea.Model
func (m *MessageViewModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *MessageViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.adjustScroll()
			}
		case "down", "j":
			if m.cursor < len(m.messages)-1 {
				m.cursor++
				m.adjustScroll()
			}
		case "home":
			m.cursor = 0
			m.scrollTop = 0
		case "end":
			m.cursor = len(m.messages) - 1
			m.adjustScroll()
		case "pageup":
			m.cursor -= m.viewport
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.adjustScroll()
		case "pagedown":
			m.cursor += m.viewport
			if m.cursor >= len(m.messages) {
				m.cursor = len(m.messages) - 1
			}
			m.adjustScroll()
		}
	}
	return m, nil
}

// adjustScroll adjusts the scroll position to keep the cursor visible
func (m *MessageViewModel) adjustScroll() {
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
func (m *MessageViewModel) View() string {
	if len(m.messages) == 0 {
		return m.styles.Unselected.Render("No messages available")
	}

	var items []string

	// Calculate visible range
	start := m.scrollTop
	end := start + m.viewport
	if end > len(m.messages) {
		end = len(m.messages)
	}

	for i := start; i < end; i++ {
		message := m.messages[i]
		messageText := m.formatMessage(message, 0, i == m.cursor)
		items = append(items, messageText)
	}

	content := strings.Join(items, "\n")

	// Add scroll indicators
	if m.scrollTop > 0 {
		content = "↑ More above\n" + content
	}
	if m.scrollTop+m.viewport < len(m.messages) {
		content = content + "\n↓ More below"
	}

	return content
}

// formatMessage formats a single message for display
func (m *MessageViewModel) formatMessage(message models.Message, indent int, selected bool) string {
	var parts []string

	// Get user info
	userName := message.User
	if user, exists := m.users[message.User]; exists {
		if user.Profile.DisplayName != "" {
			userName = user.Profile.DisplayName
		} else if user.RealName != "" {
			userName = user.RealName
		} else {
			userName = user.Name
		}
	}

	// Parse timestamp
	timestamp, err := strconv.ParseFloat(message.Timestamp, 64)
	if err != nil {
		timestamp = 0
	}
	timeStr := time.Unix(int64(timestamp), 0).Format("15:04")

	// Create indentation for threads
	indentStr := strings.Repeat("  ", indent)

	// Format header
	var header string
	if indent > 0 {
		header = fmt.Sprintf("%s↳ %s %s",
			indentStr,
			m.styles.Username.Render(userName),
			m.styles.Timestamp.Render(timeStr))
	} else {
		header = fmt.Sprintf("%s👤 %s %s",
			indentStr,
			m.styles.Username.Render(userName),
			m.styles.Timestamp.Render(timeStr))
	}

	if message.Edited != nil {
		header += m.styles.Timestamp.Render(" (edited)")
	}

	parts = append(parts, header)

	// Format message text
	text := message.Text
	if text == "" && len(message.Attachments) > 0 {
		text = m.styles.Attachment.Render("[Attachment]")
	}
	if text == "" && len(message.Files) > 0 {
		text = m.styles.Attachment.Render("[File]")
	}
	if text == "" {
		text = m.styles.Attachment.Render("[No text content]")
	}

	// Wrap text to fit width
	wrappedText := m.wrapText(text, m.width-len(indentStr)-4)
	lines := strings.Split(wrappedText, "\n")
	for _, line := range lines {
		parts = append(parts, fmt.Sprintf("%s  %s", indentStr, line))
	}

	// Add attachments
	if len(message.Attachments) > 0 {
		for _, att := range message.Attachments {
			attText := m.styles.Attachment.Render(fmt.Sprintf("📎 %s", att.Title))
			parts = append(parts, fmt.Sprintf("%s  %s", indentStr, attText))
		}
	}

	// Add files
	if len(message.Files) > 0 {
		for _, file := range message.Files {
			fileText := m.styles.Attachment.Render(fmt.Sprintf("📁 %s (%s)", file.Name, file.Filetype))
			parts = append(parts, fmt.Sprintf("%s  %s", indentStr, fileText))
		}
	}

	// Add reactions
	if len(message.Reactions) > 0 {
		var reactions []string
		for _, reaction := range message.Reactions {
			reactions = append(reactions, fmt.Sprintf(":%s: %d", reaction.Name, reaction.Count))
		}
		reactionText := m.styles.Reaction.Render(strings.Join(reactions, " "))
		parts = append(parts, fmt.Sprintf("%s  👍 %s", indentStr, reactionText))
	}

	// Add thread replies
	if len(message.Thread) > 0 {
		threadHeader := m.styles.Thread.Render(fmt.Sprintf("💬 %d replies:", len(message.Thread)))
		parts = append(parts, fmt.Sprintf("%s  %s", indentStr, threadHeader))

		// Show first few thread replies
		maxReplies := 3
		for i, reply := range message.Thread {
			if i >= maxReplies {
				remaining := len(message.Thread) - maxReplies
				moreText := m.styles.Thread.Render(fmt.Sprintf("... and %d more replies", remaining))
				parts = append(parts, fmt.Sprintf("%s    %s", indentStr, moreText))
				break
			}
			replyText := m.formatMessage(reply, indent+1, false)
			parts = append(parts, replyText)
		}
	}

	messageContent := strings.Join(parts, "\n")

	// Apply selection styling
	if selected {
		return m.styles.Selected.Width(m.width - 2).Render(messageContent)
	}

	return m.styles.Unselected.Render(messageContent)
}

// wrapText wraps text to the specified width
func (m *MessageViewModel) wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine []string
	currentLength := 0

	for _, word := range words {
		wordLength := len(word)

		// If adding this word would exceed the width, start a new line
		if currentLength+wordLength+len(currentLine) > width && len(currentLine) > 0 {
			lines = append(lines, strings.Join(currentLine, " "))
			currentLine = []string{word}
			currentLength = wordLength
		} else {
			currentLine = append(currentLine, word)
			currentLength += wordLength
		}
	}

	// Add the last line
	if len(currentLine) > 0 {
		lines = append(lines, strings.Join(currentLine, " "))
	}

	return strings.Join(lines, "\n")
}

// GetSelectedMessage returns the currently selected message
func (m *MessageViewModel) GetSelectedMessage() *models.Message {
	if len(m.messages) == 0 || m.cursor >= len(m.messages) {
		return nil
	}
	return &m.messages[m.cursor]
}

// GetMessageCount returns the total number of messages
func (m *MessageViewModel) GetMessageCount() int {
	return len(m.messages)
}
