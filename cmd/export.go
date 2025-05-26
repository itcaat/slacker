package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/itcaat/slacker/internal/api"
	"github.com/itcaat/slacker/internal/config"
	"github.com/itcaat/slacker/internal/usecase"
	"github.com/itcaat/slacker/models"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export Slack channel history to JSON",
	Long: `Export complete Slack channel history including messages, threads, 
attachments, files, reactions, and user information to a structured JSON file.

Examples:
  # Export a channel by name
  slacker export --channel general --output general-export.json

  # Export with threads and compress with gzip
  slacker export --channel general --threads --compress gzip --output general.json

  # Export date range with pretty formatting
  slacker export --channel general --from 2024-01-01 --to 2024-01-31 --format json-pretty

  # Export without files and reactions for smaller output
  slacker export --channel general --no-files --no-reactions --format json-compact`,
	RunE: runExport,
}

var (
	exportChannel   string
	exportChannelID string
	exportOutput    string
	exportFormat    string
	exportCompress  string
	exportThreads   bool
	exportFiles     bool
	exportReactions bool
	exportFromDate  string
	exportToDate    string
	exportVerbose   bool
)

func init() {
	rootCmd.AddCommand(exportCmd)

	// Channel selection
	exportCmd.Flags().StringVarP(&exportChannel, "channel", "c", "", "Channel name to export (required)")
	exportCmd.Flags().StringVar(&exportChannelID, "channel-id", "", "Channel ID to export (alternative to --channel)")

	// Output options
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file path (default: <channel>-export-<timestamp>.json)")
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "json-pretty", "Output format: json, json-pretty, json-compact")
	exportCmd.Flags().StringVar(&exportCompress, "compress", "", "Compression: none, gzip")

	// Content options
	exportCmd.Flags().BoolVar(&exportThreads, "threads", true, "Include thread replies")
	exportCmd.Flags().BoolVar(&exportFiles, "files", true, "Include file attachments")
	exportCmd.Flags().BoolVar(&exportReactions, "reactions", true, "Include message reactions")
	exportCmd.Flags().BoolVar(&exportThreads, "no-threads", false, "Exclude thread replies")
	exportCmd.Flags().BoolVar(&exportFiles, "no-files", false, "Exclude file attachments")
	exportCmd.Flags().BoolVar(&exportReactions, "no-reactions", false, "Exclude message reactions")

	// Date filtering
	exportCmd.Flags().StringVar(&exportFromDate, "from", "", "Start date (YYYY-MM-DD or YYYY-MM-DD HH:MM:SS)")
	exportCmd.Flags().StringVar(&exportToDate, "to", "", "End date (YYYY-MM-DD or YYYY-MM-DD HH:MM:SS)")

	// Other options
	exportCmd.Flags().BoolVarP(&exportVerbose, "verbose", "v", false, "Verbose output with detailed progress")

	// Mark channel as required (either --channel or --channel-id)
	exportCmd.MarkFlagRequired("channel")
}

func runExport(cmd *cobra.Command, args []string) error {
	// Load configuration
	configManager := config.NewManager()
	token, err := configManager.GetToken()
	if err != nil {
		return fmt.Errorf("Slack token not configured. Run 'slacker auth <token>' first")
	}

	// Validate channel specification
	if exportChannel == "" && exportChannelID == "" {
		return fmt.Errorf("either --channel or --channel-id must be specified")
	}

	// Create Slack client
	slackClient := api.NewSlackClient(token, exportVerbose)

	// Resolve channel ID if channel name was provided
	channelID := exportChannelID
	channelName := exportChannel
	if channelID == "" {
		channel, err := slackClient.GetChannelByName(cmd.Context(), exportChannel)
		if err != nil {
			return fmt.Errorf("failed to find channel '%s': %w", exportChannel, err)
		}
		channelID = channel.ID
		channelName = channel.Name
	}

	// Parse date filters
	var fromDate, toDate *time.Time
	if exportFromDate != "" {
		parsed, err := parseDate(exportFromDate)
		if err != nil {
			return fmt.Errorf("invalid from date '%s': %w", exportFromDate, err)
		}
		fromDate = &parsed
	}
	if exportToDate != "" {
		parsed, err := parseDate(exportToDate)
		if err != nil {
			return fmt.Errorf("invalid to date '%s': %w", exportToDate, err)
		}
		toDate = &parsed
	}

	// Handle negative flags
	if cmd.Flags().Changed("no-threads") {
		exportThreads = false
	}
	if cmd.Flags().Changed("no-files") {
		exportFiles = false
	}
	if cmd.Flags().Changed("no-reactions") {
		exportReactions = false
	}

	// Generate output filename if not specified
	outputFile := exportOutput
	if outputFile == "" {
		timestamp := time.Now().Format("20060102-150405")
		outputFile = fmt.Sprintf("%s-export-%s.json", channelName, timestamp)
	}

	// Validate format
	validFormats := map[string]bool{
		"json":         true,
		"json-pretty":  true,
		"json-compact": true,
	}
	if !validFormats[exportFormat] {
		return fmt.Errorf("invalid format '%s'. Valid formats: json, json-pretty, json-compact", exportFormat)
	}

	// Validate compression
	if exportCompress != "" {
		validCompressions := map[string]bool{
			"none": true,
			"gzip": true,
		}
		if !validCompressions[exportCompress] {
			return fmt.Errorf("invalid compression '%s'. Valid compressions: none, gzip", exportCompress)
		}
	}

	// Create export options
	options := models.ExportOptions{
		ChannelID:        channelID,
		ChannelName:      channelName,
		IncludeThreads:   exportThreads,
		IncludeFiles:     exportFiles,
		IncludeReactions: exportReactions,
		DateFrom:         fromDate,
		DateTo:           toDate,
		OutputFile:       outputFile,
		Format:           exportFormat,
		Compression:      exportCompress,
	}

	// Create export service
	version := viper.GetString("version")
	if version == "" {
		version = "1.0.0"
	}
	exportService := usecase.NewExportService(slackClient, version)

	// Print export information
	fmt.Printf("🚀 Starting export of channel '%s'\n", channelName)
	fmt.Printf("📁 Output file: %s\n", outputFile)
	fmt.Printf("📊 Format: %s", exportFormat)
	if exportCompress != "" && exportCompress != "none" {
		fmt.Printf(" (compressed with %s)", exportCompress)
	}
	fmt.Println()

	if fromDate != nil || toDate != nil {
		fmt.Printf("📅 Date range: ")
		if fromDate != nil {
			fmt.Printf("from %s ", fromDate.Format("2006-01-02"))
		}
		if toDate != nil {
			fmt.Printf("to %s ", toDate.Format("2006-01-02"))
		}
		fmt.Println()
	}

	fmt.Printf("🔧 Options: threads=%v, files=%v, reactions=%v\n",
		exportThreads, exportFiles, exportReactions)
	fmt.Println()

	// Progress tracking
	var lastProgress models.ExportProgress
	progressCallback := func(progress models.ExportProgress) {
		if exportVerbose {
			// Verbose progress with detailed information
			fmt.Printf("\r🔄 [%s] %s (%.1f%%) - %s",
				progress.Stage,
				progress.CurrentStep,
				progress.Progress*100,
				progress.ElapsedTime.Round(time.Second))

			if progress.MessagesTotal > 0 {
				fmt.Printf(" - %d messages", progress.MessagesTotal)
			}
			if progress.ThreadsTotal > 0 {
				fmt.Printf(" - %d/%d threads", progress.ThreadsCurrent, progress.ThreadsTotal)
			}
		} else {
			// Simple progress bar
			if progress.Stage != lastProgress.Stage {
				fmt.Printf("\n%s %s...", getStageEmoji(progress.Stage), getStageDescription(progress.Stage))
			}

			// Update progress bar
			barWidth := 30
			filled := int(progress.Progress * float64(barWidth))
			bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
			fmt.Printf("\r[%s] %.1f%% - %s", bar, progress.Progress*100, progress.ElapsedTime.Round(time.Second))
		}

		lastProgress = progress
	}

	// Start export
	result, err := exportService.ExportChannel(options, progressCallback)

	// Clear progress line
	fmt.Print("\r" + strings.Repeat(" ", 80) + "\r")

	if err != nil {
		fmt.Printf("❌ Export failed: %v\n", err)
		return err
	}

	if !result.Success {
		fmt.Printf("❌ Export failed: %s\n", result.Error)
		return fmt.Errorf("export failed: %s", result.Error)
	}

	// Print success information
	fmt.Printf("✅ Export completed successfully!\n\n")
	fmt.Printf("📁 Output file: %s\n", result.OutputFile)
	fmt.Printf("📏 File size: %s\n", formatFileSize(result.FileSize))
	fmt.Printf("⏱️  Duration: %s\n\n", result.Duration.Round(time.Millisecond))

	// Print statistics
	stats := result.Statistics
	fmt.Printf("📊 Export Statistics:\n")
	fmt.Printf("   Messages: %d (including %d thread replies)\n", stats.TotalMessages, stats.TotalReplies)
	fmt.Printf("   Threads: %d\n", stats.TotalThreads)
	fmt.Printf("   Users: %d\n", stats.TotalUsers)
	fmt.Printf("   Attachments: %d\n", stats.TotalAttachments)
	fmt.Printf("   Files: %d\n", stats.TotalFiles)
	fmt.Printf("   Reactions: %d\n", stats.TotalReactions)

	if len(stats.TopReactions) > 0 {
		fmt.Printf("\n🎭 Top Reactions:\n")
		for i, reaction := range stats.TopReactions {
			if i >= 5 { // Show top 5
				break
			}
			fmt.Printf("   %s: %d\n", reaction.Name, reaction.Count)
		}
	}

	if exportVerbose {
		fmt.Printf("\n⏱️  Processing Times:\n")
		fmt.Printf("   Channel fetch: %s\n", stats.ProcessingTime.ChannelFetch.Round(time.Millisecond))
		fmt.Printf("   Message fetch: %s\n", stats.ProcessingTime.MessageFetch.Round(time.Millisecond))
		fmt.Printf("   Thread fetch: %s\n", stats.ProcessingTime.ThreadFetch.Round(time.Millisecond))
		fmt.Printf("   User fetch: %s\n", stats.ProcessingTime.UserFetch.Round(time.Millisecond))
		fmt.Printf("   Data processing: %s\n", stats.ProcessingTime.DataProcessing.Round(time.Millisecond))
		fmt.Printf("   File generation: %s\n", stats.ProcessingTime.FileGeneration.Round(time.Millisecond))
	}

	return nil
}

// parseDate parses date strings in various formats
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date '%s'. Supported formats: YYYY-MM-DD, YYYY-MM-DD HH:MM:SS", dateStr)
}

// getStageEmoji returns an emoji for each export stage
func getStageEmoji(stage string) string {
	switch stage {
	case "initializing":
		return "🚀"
	case "channel_fetch":
		return "📋"
	case "message_fetch":
		return "💬"
	case "thread_fetch":
		return "🧵"
	case "user_fetch":
		return "👥"
	case "data_processing":
		return "⚙️"
	case "file_generation":
		return "📁"
	case "complete":
		return "✅"
	default:
		return "🔄"
	}
}

// getStageDescription returns a human-readable description for each stage
func getStageDescription(stage string) string {
	switch stage {
	case "initializing":
		return "Initializing export"
	case "channel_fetch":
		return "Fetching channel information"
	case "message_fetch":
		return "Fetching messages"
	case "thread_fetch":
		return "Fetching thread replies"
	case "user_fetch":
		return "Fetching user information"
	case "data_processing":
		return "Processing data"
	case "file_generation":
		return "Generating output file"
	case "complete":
		return "Export complete"
	default:
		return "Processing"
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
