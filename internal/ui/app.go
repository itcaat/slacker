package ui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/itcaat/slacker/internal/api"
	"github.com/itcaat/slacker/internal/config"
	"github.com/itcaat/slacker/internal/usecase"
	"github.com/itcaat/slacker/models"
)

// AppState represents the current state of the application
type AppState int

const (
	StateLoading AppState = iota
	StateChannelList
	StateMessageView
	StateExporting
	StateError
	StateQuit
)

// App represents the main TUI application
type App struct {
	state           AppState
	width           int
	height          int
	slackClient     *api.SlackClient
	messageService  *usecase.MessageService
	channels        []models.Channel
	selectedChannel *models.Channel
	messages        []models.Message
	users           map[string]models.User
	error           error
	loading         bool

	// UI components
	channelList *ChannelListModel
	messageView *MessageViewModel

	// Styles
	styles Styles
}

// Styles contains all the styling for the TUI
type Styles struct {
	Header     lipgloss.Style
	Footer     lipgloss.Style
	Error      lipgloss.Style
	Loading    lipgloss.Style
	Border     lipgloss.Style
	Selected   lipgloss.Style
	Unselected lipgloss.Style
	Message    lipgloss.Style
	Thread     lipgloss.Style
	Username   lipgloss.Style
	Timestamp  lipgloss.Style
}

// NewApp creates a new TUI application
func NewApp() (*App, error) {
	// Get configuration
	configManager := config.NewManager()
	token, err := configManager.GetToken()
	if err != nil {
		return nil, fmt.Errorf("authentication required. Run 'slacker auth <token>' first: %w", err)
	}

	// Create Slack client
	slackClient := api.NewSlackClient(token, false)
	messageService := usecase.NewMessageService(slackClient)

	app := &App{
		state:          StateLoading,
		slackClient:    slackClient,
		messageService: messageService,
		loading:        true,
		styles:         createStyles(),
	}

	// Initialize UI components
	app.channelList = NewChannelListModel()
	app.messageView = NewMessageViewModel()

	return app, nil
}

// createStyles initializes the application styles
func createStyles() Styles {
	return Styles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1),

		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A49FA5")).
			Background(lipgloss.Color("#2B2B2B")).
			Padding(0, 1),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true),

		Loading: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true),

		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")),

		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Bold(true),

		Unselected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A49FA5")),

		Message: lipgloss.NewStyle().
			Padding(0, 1),

		Thread: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A49FA5")).
			MarginLeft(2),

		Username: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true),

		Timestamp: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")),
	}
}

// Init implements tea.Model
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.loadChannels(),
		tea.EnterAltScreen,
	)
}

// Update implements tea.Model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

		// Update component sizes
		if a.channelList != nil {
			a.channelList.SetSize(a.width/3, a.height-4)
		}
		if a.messageView != nil {
			a.messageView.SetSize(2*a.width/3, a.height-4)
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			a.state = StateQuit
			return a, tea.Quit
		case "esc":
			if a.state == StateMessageView {
				a.state = StateChannelList
				a.selectedChannel = nil
			}
		case "r":
			// Refresh data
			if a.state == StateChannelList {
				a.loading = true
				return a, a.loadChannels()
			} else if a.state == StateMessageView && a.selectedChannel != nil {
				a.loading = true
				return a, a.loadMessages(a.selectedChannel.ID)
			}
		case "e":
			// Export channel
			if a.selectedChannel != nil && (a.state == StateChannelList || a.state == StateMessageView) {
				a.state = StateExporting
				return a, a.exportChannel(a.selectedChannel)
			}
		}

	case channelsLoadedMsg:
		a.loading = false
		a.channels = msg.channels
		a.users = msg.users
		a.state = StateChannelList
		a.channelList.SetChannels(a.channels)

	case messagesLoadedMsg:
		a.loading = false
		a.messages = msg.messages
		a.state = StateMessageView
		a.messageView.SetMessages(a.messages, a.users)

	case errorMsg:
		a.loading = false
		a.error = msg.error
		a.state = StateError

	case channelSelectedMsg:
		a.selectedChannel = &msg.channel
		a.loading = true
		return a, a.loadMessages(msg.channel.ID)

	case exportCompletedMsg:
		a.loading = false
		a.state = StateChannelList
		// Show success message (in a real implementation, we might want to show this in the UI)
		if msg.result.Success {
			fmt.Printf("\n✅ Export completed: %s (%s)\n", msg.result.OutputFile, formatFileSize(msg.result.FileSize))
		}

	case exportProgressMsg:
		// Update export progress (could be displayed in UI)
	}

	// Update current component
	switch a.state {
	case StateChannelList:
		if a.channelList != nil {
			var newModel tea.Model
			newModel, cmd = a.channelList.Update(msg)
			a.channelList = newModel.(*ChannelListModel)
			cmds = append(cmds, cmd)
		}
	case StateMessageView:
		if a.messageView != nil {
			var newModel tea.Model
			newModel, cmd = a.messageView.Update(msg)
			a.messageView = newModel.(*MessageViewModel)
			cmds = append(cmds, cmd)
		}
	}

	return a, tea.Batch(cmds...)
}

// View implements tea.Model
func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	// Header
	header := a.styles.Header.Width(a.width).Render("Slacker - Slack CLI Client")

	// Footer
	var footer string
	switch a.state {
	case StateChannelList:
		if a.selectedChannel != nil {
			footer = "↑/↓: navigate • enter: select channel • e: export • r: refresh • q: quit"
		} else {
			footer = "↑/↓: navigate • enter: select channel • r: refresh • q: quit"
		}
	case StateMessageView:
		footer = "↑/↓: scroll • e: export • esc: back to channels • r: refresh • q: quit"
	case StateExporting:
		footer = "Exporting channel... please wait"
	case StateError:
		footer = "r: retry • q: quit"
	default:
		footer = "q: quit"
	}
	footerView := a.styles.Footer.Width(a.width).Render(footer)

	// Content area height
	contentHeight := a.height - 2 // Subtract header and footer

	// Content
	var content string
	switch a.state {
	case StateLoading:
		content = a.renderLoading(contentHeight)
	case StateChannelList:
		content = a.renderChannelList(contentHeight)
	case StateMessageView:
		content = a.renderMessageView(contentHeight)
	case StateExporting:
		content = a.renderExporting(contentHeight)
	case StateError:
		content = a.renderError(contentHeight)
	default:
		content = "Unknown state"
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, content, footerView)
}

// renderLoading renders the loading state
func (a *App) renderLoading(height int) string {
	loading := a.styles.Loading.Render("🔄 Loading...")
	return lipgloss.Place(a.width, height, lipgloss.Center, lipgloss.Center, loading)
}

// renderChannelList renders the channel list view
func (a *App) renderChannelList(height int) string {
	if a.channelList == nil {
		return ""
	}

	title := lipgloss.NewStyle().Bold(true).Render("📢 Channels")
	channelView := a.channelList.View()

	content := lipgloss.JoinVertical(lipgloss.Left, title, channelView)
	return a.styles.Border.Width(a.width - 2).Height(height - 2).Render(content)
}

// renderMessageView renders the message view
func (a *App) renderMessageView(height int) string {
	if a.messageView == nil || a.selectedChannel == nil {
		return ""
	}

	title := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("💬 #%s", a.selectedChannel.Name))
	messageView := a.messageView.View()

	content := lipgloss.JoinVertical(lipgloss.Left, title, messageView)
	return a.styles.Border.Width(a.width - 2).Height(height - 2).Render(content)
}

// renderError renders the error state
func (a *App) renderError(height int) string {
	errorText := a.styles.Error.Render(fmt.Sprintf("❌ Error: %v", a.error))
	return lipgloss.Place(a.width, height, lipgloss.Center, lipgloss.Center, errorText)
}

// renderExporting renders the export state
func (a *App) renderExporting(height int) string {
	if a.selectedChannel == nil {
		return ""
	}

	exportText := a.styles.Loading.Render(fmt.Sprintf("📤 Exporting channel #%s...", a.selectedChannel.Name))
	return lipgloss.Place(a.width, height, lipgloss.Center, lipgloss.Center, exportText)
}

// loadChannels loads the channel list
func (a *App) loadChannels() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		channels, err := a.slackClient.GetChannels(ctx)
		if err != nil {
			return errorMsg{error: err}
		}

		users, err := a.slackClient.GetUsers(ctx)
		if err != nil {
			return errorMsg{error: err}
		}

		userMap := make(map[string]models.User)
		for _, user := range users {
			userMap[user.ID] = user
		}

		return channelsLoadedMsg{
			channels: channels,
			users:    userMap,
		}
	}
}

// loadMessages loads messages for a channel
func (a *App) loadMessages(channelID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		opts := usecase.MessageRetrievalOptions{
			ChannelID:      channelID,
			Limit:          50,
			IncludeThreads: true,
			IncludeUsers:   true,
		}

		result, err := a.messageService.GetChannelMessages(ctx, opts)
		if err != nil {
			return errorMsg{error: err}
		}

		return messagesLoadedMsg{
			messages: result.Messages,
		}
	}
}

// Messages for tea.Cmd communication
type channelsLoadedMsg struct {
	channels []models.Channel
	users    map[string]models.User
}

type messagesLoadedMsg struct {
	messages []models.Message
}

type errorMsg struct {
	error error
}

type channelSelectedMsg struct {
	channel models.Channel
}

type exportProgressMsg struct {
	progress models.ExportProgress
}

type exportCompletedMsg struct {
	result *models.ExportResult
}

// exportChannel starts the export process for a channel
func (a *App) exportChannel(channel *models.Channel) tea.Cmd {
	return func() tea.Msg {
		// Create export service
		exportService := usecase.NewExportService(a.slackClient, "1.0.0")

		// Generate output filename
		timestamp := time.Now().Format("20060102-150405")
		outputFile := fmt.Sprintf("%s-export-%s.json", channel.Name, timestamp)

		// Create export options
		options := models.ExportOptions{
			ChannelID:        channel.ID,
			ChannelName:      channel.Name,
			IncludeThreads:   true,
			IncludeFiles:     true,
			IncludeReactions: true,
			OutputFile:       outputFile,
			Format:           "json-pretty",
			Compression:      "",
		}

		// Progress callback (could be used to update UI)
		progressCallback := func(progress models.ExportProgress) {
			// For now, we'll just ignore progress updates in TUI
			// In a more advanced implementation, we could send progress messages
		}

		// Start export
		result, err := exportService.ExportChannel(options, progressCallback)
		if err != nil {
			return errorMsg{error: err}
		}

		return exportCompletedMsg{result: result}
	}
}

// formatFileSize formats file size in human-readable format
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// RunTUI starts the TUI application
func RunTUI() error {
	app, err := NewApp()
	if err != nil {
		return err
	}

	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
