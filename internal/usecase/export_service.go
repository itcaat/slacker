package usecase

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/itcaat/slacker/models"
)

// SlackClientInterface defines the interface for Slack API operations
type SlackClientInterface interface {
	GetChannels(ctx context.Context) ([]models.Channel, error)
	GetChannelHistory(ctx context.Context, channelID string, limit int, cursor string) ([]models.Message, string, error)
	GetThreadReplies(ctx context.Context, channelID, threadTS string) ([]models.Message, error)
	GetUsers(ctx context.Context) ([]models.User, error)
}

// ExportService handles the export of Slack channel data
type ExportService struct {
	slackClient SlackClientInterface
	version     string
}

// NewExportService creates a new export service
func NewExportService(slackClient SlackClientInterface, version string) *ExportService {
	return &ExportService{
		slackClient: slackClient,
		version:     version,
	}
}

// ExportChannel exports a complete Slack channel with all messages and threads
func (s *ExportService) ExportChannel(options models.ExportOptions, progressCallback func(models.ExportProgress)) (*models.ExportResult, error) {
	startTime := time.Now()

	// Initialize progress tracking
	progress := models.ExportProgress{
		Stage:       "initializing",
		CurrentStep: "Starting export",
		Progress:    0.0,
		ElapsedTime: 0,
	}

	if progressCallback != nil {
		progressCallback(progress)
	}

	// Step 1: Fetch channel information
	progress.Stage = "channel_fetch"
	progress.CurrentStep = "Fetching channel information"
	progress.Progress = 0.1
	progress.ElapsedTime = time.Since(startTime)
	if progressCallback != nil {
		progressCallback(progress)
	}

	channelFetchStart := time.Now()
	channel, err := s.fetchChannelInfo(options.ChannelID)
	if err != nil {
		return &models.ExportResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to fetch channel info: %v", err),
		}, err
	}
	channelFetchDuration := time.Since(channelFetchStart)

	// Step 2: Fetch all messages
	progress.Stage = "message_fetch"
	progress.CurrentStep = "Fetching channel messages"
	progress.Progress = 0.2
	progress.ElapsedTime = time.Since(startTime)
	if progressCallback != nil {
		progressCallback(progress)
	}

	messageFetchStart := time.Now()
	messages, err := s.fetchAllMessages(options, &progress, progressCallback, startTime)
	if err != nil {
		return &models.ExportResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to fetch messages: %v", err),
		}, err
	}
	messageFetchDuration := time.Since(messageFetchStart)

	// Step 3: Fetch thread replies if enabled
	var threadFetchDuration time.Duration
	if options.IncludeThreads {
		progress.Stage = "thread_fetch"
		progress.CurrentStep = "Fetching thread replies"
		progress.Progress = 0.6
		progress.ElapsedTime = time.Since(startTime)
		if progressCallback != nil {
			progressCallback(progress)
		}

		threadFetchStart := time.Now()
		err = s.fetchThreadReplies(messages, options.ChannelID, &progress, progressCallback, startTime)
		if err != nil {
			return &models.ExportResult{
				Success: false,
				Error:   fmt.Sprintf("Failed to fetch thread replies: %v", err),
			}, err
		}
		threadFetchDuration = time.Since(threadFetchStart)
	}

	// Step 4: Fetch user information
	progress.Stage = "user_fetch"
	progress.CurrentStep = "Fetching user information"
	progress.Progress = 0.8
	progress.ElapsedTime = time.Since(startTime)
	if progressCallback != nil {
		progressCallback(progress)
	}

	userFetchStart := time.Now()
	users, err := s.fetchUserInfo(messages)
	if err != nil {
		return &models.ExportResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to fetch user info: %v", err),
		}, err
	}
	userFetchDuration := time.Since(userFetchStart)

	// Step 5: Process and structure data
	progress.Stage = "data_processing"
	progress.CurrentStep = "Processing and structuring data"
	progress.Progress = 0.9
	progress.ElapsedTime = time.Since(startTime)
	if progressCallback != nil {
		progressCallback(progress)
	}

	dataProcessingStart := time.Now()
	exportData, statistics := s.processExportData(channel, messages, users, options, startTime)
	dataProcessingDuration := time.Since(dataProcessingStart)

	// Step 6: Generate output file
	progress.Stage = "file_generation"
	progress.CurrentStep = "Generating output file"
	progress.Progress = 0.95
	progress.ElapsedTime = time.Since(startTime)
	if progressCallback != nil {
		progressCallback(progress)
	}

	fileGenerationStart := time.Now()
	outputFile, fileSize, err := s.generateOutputFile(exportData, options)
	if err != nil {
		return &models.ExportResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to generate output file: %v", err),
		}, err
	}
	fileGenerationDuration := time.Since(fileGenerationStart)

	// Complete
	totalDuration := time.Since(startTime)
	progress.Stage = "complete"
	progress.CurrentStep = "Export completed successfully"
	progress.Progress = 1.0
	progress.ElapsedTime = totalDuration
	if progressCallback != nil {
		progressCallback(progress)
	}

	// Update processing times in statistics
	statistics.ExportDuration = totalDuration
	statistics.ProcessingTime = models.ProcessingTimeStats{
		ChannelFetch:   channelFetchDuration,
		MessageFetch:   messageFetchDuration,
		ThreadFetch:    threadFetchDuration,
		UserFetch:      userFetchDuration,
		DataProcessing: dataProcessingDuration,
		FileGeneration: fileGenerationDuration,
	}

	return &models.ExportResult{
		Success:    true,
		OutputFile: outputFile,
		FileSize:   fileSize,
		Statistics: statistics,
		Duration:   totalDuration,
	}, nil
}

// fetchChannelInfo retrieves detailed channel information
func (s *ExportService) fetchChannelInfo(channelID string) (*models.Channel, error) {
	ctx := context.Background()
	channels, err := s.slackClient.GetChannels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}

	for _, channel := range channels {
		if channel.ID == channelID {
			return &channel, nil
		}
	}

	return nil, fmt.Errorf("channel with ID %s not found", channelID)
}

// fetchAllMessages retrieves all messages from the channel with pagination
func (s *ExportService) fetchAllMessages(options models.ExportOptions, progress *models.ExportProgress, progressCallback func(models.ExportProgress), startTime time.Time) ([]models.Message, error) {
	var allMessages []models.Message
	var cursor string
	pageCount := 0

	for {
		// Fetch a page of messages
		ctx := context.Background()
		messages, nextCursor, err := s.slackClient.GetChannelHistory(ctx, options.ChannelID, 1000, cursor)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch messages (page %d): %w", pageCount+1, err)
		}

		// Filter messages by date range if specified
		filteredMessages := s.filterMessagesByDate(messages, options.DateFrom, options.DateTo)
		allMessages = append(allMessages, filteredMessages...)

		pageCount++

		// Update progress
		if progressCallback != nil {
			progress.CurrentStep = fmt.Sprintf("Fetched %d messages (%d pages)", len(allMessages), pageCount)
			progress.MessagesTotal = len(allMessages)
			progress.MessagesCurrent = len(allMessages)
			progress.ElapsedTime = time.Since(startTime)
			progressCallback(*progress)
		}

		// Check if we have more pages
		if nextCursor == "" {
			break
		}
		cursor = nextCursor

		// Rate limiting - small delay between requests
		time.Sleep(100 * time.Millisecond)
	}

	// Sort messages by timestamp (oldest first)
	sort.Slice(allMessages, func(i, j int) bool {
		return allMessages[i].Timestamp < allMessages[j].Timestamp
	})

	return allMessages, nil
}

// fetchThreadReplies fetches replies for all threaded messages
func (s *ExportService) fetchThreadReplies(messages []models.Message, channelID string, progress *models.ExportProgress, progressCallback func(models.ExportProgress), startTime time.Time) error {
	// Find all messages that have threads
	var threadedMessages []*models.Message
	for i := range messages {
		if messages[i].ReplyCount > 0 && messages[i].ThreadTS != "" {
			threadedMessages = append(threadedMessages, &messages[i])
		}
	}

	progress.ThreadsTotal = len(threadedMessages)

	// Fetch replies for each threaded message
	for i, msg := range threadedMessages {
		ctx := context.Background()
		replies, err := s.slackClient.GetThreadReplies(ctx, channelID, msg.ThreadTS)
		if err != nil {
			// Log warning but continue with export
			fmt.Printf("Warning: Failed to fetch replies for thread %s: %v\n", msg.ThreadTS, err)
			continue
		}

		// Remove the parent message from replies (first message is usually the parent)
		if len(replies) > 0 && replies[0].Timestamp == msg.Timestamp {
			replies = replies[1:]
		}

		// Sort replies by timestamp
		sort.Slice(replies, func(i, j int) bool {
			return replies[i].Timestamp < replies[j].Timestamp
		})

		msg.Thread = replies

		// Update progress
		if progressCallback != nil {
			progress.ThreadsCurrent = i + 1
			progress.CurrentStep = fmt.Sprintf("Fetched replies for %d/%d threads", i+1, len(threadedMessages))
			progress.ElapsedTime = time.Since(startTime)
			progressCallback(*progress)
		}

		// Rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// fetchUserInfo retrieves user information for all users mentioned in messages
func (s *ExportService) fetchUserInfo(messages []models.Message) (map[string]models.User, error) {
	userIDs := make(map[string]bool)

	// Collect all unique user IDs from messages and threads
	var collectUserIDs func([]models.Message)
	collectUserIDs = func(msgs []models.Message) {
		for _, msg := range msgs {
			if msg.User != "" {
				userIDs[msg.User] = true
			}
			// Collect from thread replies
			collectUserIDs(msg.Thread)
		}
	}

	collectUserIDs(messages)

	// Fetch all users from workspace
	ctx := context.Background()
	allUsers, err := s.slackClient.GetUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	// Create a map of users we need
	users := make(map[string]models.User)
	for _, user := range allUsers {
		if userIDs[user.ID] {
			users[user.ID] = user
		}
	}

	// Create placeholder users for any missing user IDs
	for userID := range userIDs {
		if _, exists := users[userID]; !exists {
			users[userID] = models.User{
				ID:       userID,
				Name:     fmt.Sprintf("user_%s", userID),
				RealName: "Unknown User",
				Deleted:  true,
			}
		}
	}

	return users, nil
}

// filterMessagesByDate filters messages based on date range
func (s *ExportService) filterMessagesByDate(messages []models.Message, dateFrom, dateTo *time.Time) []models.Message {
	if dateFrom == nil && dateTo == nil {
		return messages
	}

	var filtered []models.Message
	for _, msg := range messages {
		timestamp, err := models.ParseSlackTimestamp(msg.Timestamp)
		if err != nil {
			continue // Skip messages with invalid timestamps
		}

		// Check date range
		if dateFrom != nil && timestamp.Before(*dateFrom) {
			continue
		}
		if dateTo != nil && timestamp.After(*dateTo) {
			continue
		}

		filtered = append(filtered, msg)
	}

	return filtered
}

// processExportData converts raw data into export format and calculates statistics
func (s *ExportService) processExportData(channel *models.Channel, messages []models.Message, users map[string]models.User, options models.ExportOptions, startTime time.Time) (models.ChannelExport, models.ExportStatistics) {
	// Convert messages to export format
	var exportMessages []models.ExportMessage
	for _, msg := range messages {
		exportMsg := models.ConvertToExportMessage(msg)
		exportMessages = append(exportMessages, exportMsg)
	}

	// Convert users to export format
	exportUsers := make(map[string]models.ExportUser)
	for id, user := range users {
		exportUsers[id] = models.ConvertToExportUser(user)
	}

	// Calculate statistics
	statistics := s.calculateStatistics(messages, users)

	// Create channel info
	channelInfo := models.ChannelInfo{
		ID:         channel.ID,
		Name:       channel.Name,
		IsPrivate:  channel.IsPrivate,
		IsArchived: channel.IsArchived,
		NumMembers: channel.NumMembers,
	}

	if channel.Topic.Value != "" {
		channelInfo.Topic = channel.Topic.Value
	}
	if channel.Purpose.Value != "" {
		channelInfo.Purpose = channel.Purpose.Value
	}
	if channel.Created > 0 {
		channelInfo.CreatedAt = time.Unix(channel.Created, 0)
	}
	if channel.Creator != "" {
		channelInfo.Creator = channel.Creator
	}

	// Create export metadata
	exportInfo := models.ExportMetadata{
		ExportedAt:     time.Now(),
		ExportedBy:     "slacker-cli",
		SlackerVersion: s.version,
		ExportFormat:   options.Format,
		IncludeThreads: options.IncludeThreads,
	}

	if options.DateFrom != nil || options.DateTo != nil {
		exportInfo.DateRange = models.DateRange{
			From: options.DateFrom,
			To:   options.DateTo,
		}
	}

	// Create complete export structure
	exportData := models.ChannelExport{
		ExportInfo: exportInfo,
		Channel:    channelInfo,
		Messages:   exportMessages,
		Users:      exportUsers,
		Statistics: statistics,
	}

	return exportData, statistics
}

// calculateStatistics computes various statistics about the export
func (s *ExportService) calculateStatistics(messages []models.Message, users map[string]models.User) models.ExportStatistics {
	stats := models.ExportStatistics{
		MessagesByUser: make(map[string]int),
		MessagesByDate: make(map[string]int),
	}

	var countMessages func([]models.Message)
	countMessages = func(msgs []models.Message) {
		for _, msg := range msgs {
			stats.TotalMessages++

			// Count by user
			if msg.User != "" {
				stats.MessagesByUser[msg.User]++
			}

			// Count by date
			if timestamp, err := models.ParseSlackTimestamp(msg.Timestamp); err == nil {
				dateKey := timestamp.Format("2006-01-02")
				stats.MessagesByDate[dateKey]++
			}

			// Count attachments
			stats.TotalAttachments += len(msg.Attachments)

			// Count files
			stats.TotalFiles += len(msg.Files)

			// Count reactions
			for _, reaction := range msg.Reactions {
				stats.TotalReactions += reaction.Count
			}

			// Count threads and replies
			if len(msg.Thread) > 0 {
				stats.TotalThreads++
				stats.TotalReplies += len(msg.Thread)
				countMessages(msg.Thread) // Recursively count thread messages
			}
		}
	}

	countMessages(messages)
	stats.TotalUsers = len(users)

	// Calculate top reactions
	reactionCounts := make(map[string]int)
	var collectReactions func([]models.Message)
	collectReactions = func(msgs []models.Message) {
		for _, msg := range msgs {
			for _, reaction := range msg.Reactions {
				reactionCounts[reaction.Name] += reaction.Count
			}
			collectReactions(msg.Thread)
		}
	}
	collectReactions(messages)

	// Sort reactions by count
	type reactionPair struct {
		name  string
		count int
	}
	var reactionPairs []reactionPair
	for name, count := range reactionCounts {
		reactionPairs = append(reactionPairs, reactionPair{name, count})
	}
	sort.Slice(reactionPairs, func(i, j int) bool {
		return reactionPairs[i].count > reactionPairs[j].count
	})

	// Take top 10 reactions
	maxReactions := 10
	if len(reactionPairs) < maxReactions {
		maxReactions = len(reactionPairs)
	}
	for i := 0; i < maxReactions; i++ {
		stats.TopReactions = append(stats.TopReactions, models.ReactionStat{
			Name:  reactionPairs[i].name,
			Count: reactionPairs[i].count,
		})
	}

	return stats
}

// generateOutputFile creates the final export file
func (s *ExportService) generateOutputFile(exportData models.ChannelExport, options models.ExportOptions) (string, int64, error) {
	// Ensure output directory exists
	outputDir := filepath.Dir(options.OutputFile)
	if outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return "", 0, fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Marshal JSON based on format
	var jsonData []byte
	var err error

	switch options.Format {
	case "json-pretty":
		jsonData, err = json.MarshalIndent(exportData, "", "  ")
	case "json-compact":
		jsonData, err = json.Marshal(exportData)
	default: // "json"
		jsonData, err = json.MarshalIndent(exportData, "", "  ")
	}

	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal export data: %w", err)
	}

	// Handle compression
	outputFile := options.OutputFile
	switch options.Compression {
	case "gzip":
		if !strings.HasSuffix(outputFile, ".gz") {
			outputFile += ".gz"
		}
		return s.writeGzipFile(outputFile, jsonData)
	case "zip":
		if !strings.HasSuffix(outputFile, ".zip") {
			outputFile += ".zip"
		}
		return s.writeZipFile(outputFile, jsonData, filepath.Base(options.OutputFile))
	default: // "none" or empty
		return s.writeFile(outputFile, jsonData)
	}
}

// writeFile writes data to a regular file
func (s *ExportService) writeFile(filename string, data []byte) (string, int64, error) {
	err := os.WriteFile(filename, data, 0644)
	if err != nil {
		return "", 0, fmt.Errorf("failed to write file: %w", err)
	}

	fileInfo, err := os.Stat(filename)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get file info: %w", err)
	}

	return filename, fileInfo.Size(), nil
}

// writeGzipFile writes data to a gzip-compressed file
func (s *ExportService) writeGzipFile(filename string, data []byte) (string, int64, error) {
	file, err := os.Create(filename)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create gzip file: %w", err)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	_, err = gzipWriter.Write(data)
	if err != nil {
		return "", 0, fmt.Errorf("failed to write gzip data: %w", err)
	}

	err = gzipWriter.Close()
	if err != nil {
		return "", 0, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	err = file.Close()
	if err != nil {
		return "", 0, fmt.Errorf("failed to close file: %w", err)
	}

	fileInfo, err := os.Stat(filename)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get file info: %w", err)
	}

	return filename, fileInfo.Size(), nil
}

// writeZipFile writes data to a zip-compressed file
func (s *ExportService) writeZipFile(filename string, data []byte, entryName string) (string, int64, error) {
	// For now, we'll implement a simple approach without the archive/zip package
	// This can be enhanced later if zip compression is needed
	return "", 0, fmt.Errorf("zip compression not yet implemented")
}
