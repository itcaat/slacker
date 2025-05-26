package models

import (
	"strconv"
	"time"
)

// ChannelExport represents the complete export structure for a Slack channel
type ChannelExport struct {
	// Export metadata
	ExportInfo ExportMetadata `json:"export_info"`

	// Channel information
	Channel ChannelInfo `json:"channel"`

	// All messages with thread structure preserved
	Messages []ExportMessage `json:"messages"`

	// User directory for message attribution
	Users map[string]ExportUser `json:"users"`

	// Export statistics
	Statistics ExportStatistics `json:"statistics"`
}

// ExportMetadata contains information about the export itself
type ExportMetadata struct {
	ExportedAt     time.Time `json:"exported_at"`
	ExportedBy     string    `json:"exported_by"`
	SlackerVersion string    `json:"slacker_version"`
	ExportFormat   string    `json:"export_format"`
	IncludeThreads bool      `json:"include_threads"`
	DateRange      DateRange `json:"date_range,omitempty"`
}

// DateRange represents the time range of exported messages
type DateRange struct {
	From *time.Time `json:"from,omitempty"`
	To   *time.Time `json:"to,omitempty"`
}

// ChannelInfo contains detailed information about the exported channel
type ChannelInfo struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	IsPrivate  bool      `json:"is_private"`
	IsArchived bool      `json:"is_archived"`
	Topic      string    `json:"topic,omitempty"`
	Purpose    string    `json:"purpose,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	Creator    string    `json:"creator,omitempty"`
	NumMembers int       `json:"num_members"`
	Members    []string  `json:"members,omitempty"`
}

// ExportMessage represents a message in the export with enhanced structure
type ExportMessage struct {
	// Core message data
	ID        string    `json:"id"`
	User      string    `json:"user"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`

	// Message metadata
	Type    string    `json:"type"`
	Subtype string    `json:"subtype,omitempty"`
	Edited  *EditInfo `json:"edited,omitempty"`

	// Thread information
	ThreadTimestamp string     `json:"thread_ts,omitempty"`
	ParentUserID    string     `json:"parent_user_id,omitempty"`
	ReplyCount      int        `json:"reply_count,omitempty"`
	ReplyUsers      []string   `json:"reply_users,omitempty"`
	ReplyUsersCount int        `json:"reply_users_count,omitempty"`
	LatestReply     *time.Time `json:"latest_reply,omitempty"`

	// Thread replies (nested structure)
	Replies []ExportMessage `json:"replies,omitempty"`

	// Rich content
	Attachments []ExportAttachment `json:"attachments,omitempty"`
	Files       []ExportFile       `json:"files,omitempty"`
	Reactions   []ExportReaction   `json:"reactions,omitempty"`

	// Message context
	Permalink   string `json:"permalink,omitempty"`
	ClientMsgID string `json:"client_msg_id,omitempty"`
}

// EditInfo contains information about message edits
type EditInfo struct {
	User      string    `json:"user"`
	Timestamp time.Time `json:"ts"`
}

// ExportAttachment represents an attachment in the export
type ExportAttachment struct {
	ID         string            `json:"id,omitempty"`
	Title      string            `json:"title,omitempty"`
	TitleLink  string            `json:"title_link,omitempty"`
	Text       string            `json:"text,omitempty"`
	Fallback   string            `json:"fallback,omitempty"`
	Color      string            `json:"color,omitempty"`
	Pretext    string            `json:"pretext,omitempty"`
	AuthorName string            `json:"author_name,omitempty"`
	AuthorLink string            `json:"author_link,omitempty"`
	AuthorIcon string            `json:"author_icon,omitempty"`
	ImageURL   string            `json:"image_url,omitempty"`
	ThumbURL   string            `json:"thumb_url,omitempty"`
	Footer     string            `json:"footer,omitempty"`
	FooterIcon string            `json:"footer_icon,omitempty"`
	Timestamp  *time.Time        `json:"ts,omitempty"`
	Fields     []AttachmentField `json:"fields,omitempty"`
}

// AttachmentField represents a field in an attachment
type AttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// ExportFile represents a file in the export
type ExportFile struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Title              string    `json:"title,omitempty"`
	Mimetype           string    `json:"mimetype"`
	Filetype           string    `json:"filetype"`
	PrettyType         string    `json:"pretty_type,omitempty"`
	User               string    `json:"user"`
	Mode               string    `json:"mode,omitempty"`
	Editable           bool      `json:"editable,omitempty"`
	IsExternal         bool      `json:"is_external,omitempty"`
	ExternalType       string    `json:"external_type,omitempty"`
	Size               int       `json:"size,omitempty"`
	URLPrivate         string    `json:"url_private,omitempty"`
	URLPrivateDownload string    `json:"url_private_download,omitempty"`
	Permalink          string    `json:"permalink,omitempty"`
	PermalinkPublic    string    `json:"permalink_public,omitempty"`
	Timestamp          time.Time `json:"timestamp"`
	IsPublic           bool      `json:"is_public,omitempty"`
	PublicURLShared    bool      `json:"public_url_shared,omitempty"`
	DisplayAsBot       bool      `json:"display_as_bot,omitempty"`
	Username           string    `json:"username,omitempty"`
	ThumbTiny          string    `json:"thumb_tiny,omitempty"`
	Thumb64            string    `json:"thumb_64,omitempty"`
	Thumb80            string    `json:"thumb_80,omitempty"`
	Thumb160           string    `json:"thumb_160,omitempty"`
	Thumb360           string    `json:"thumb_360,omitempty"`
	Thumb480           string    `json:"thumb_480,omitempty"`
	Thumb720           string    `json:"thumb_720,omitempty"`
	Thumb800           string    `json:"thumb_800,omitempty"`
	Thumb960           string    `json:"thumb_960,omitempty"`
	Thumb1024          string    `json:"thumb_1024,omitempty"`
	ImageExifRotation  int       `json:"image_exif_rotation,omitempty"`
	OriginalW          int       `json:"original_w,omitempty"`
	OriginalH          int       `json:"original_h,omitempty"`
	ThumbW             int       `json:"thumb_w,omitempty"`
	ThumbH             int       `json:"thumb_h,omitempty"`
}

// ExportReaction represents a reaction in the export
type ExportReaction struct {
	Name  string   `json:"name"`
	Count int      `json:"count"`
	Users []string `json:"users"`
}

// ExportUser represents a user in the export
type ExportUser struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	RealName string        `json:"real_name,omitempty"`
	Profile  ExportProfile `json:"profile"`
	IsBot    bool          `json:"is_bot,omitempty"`
	IsAdmin  bool          `json:"is_admin,omitempty"`
	IsOwner  bool          `json:"is_owner,omitempty"`
	Deleted  bool          `json:"deleted,omitempty"`
	TZ       string        `json:"tz,omitempty"`
	TZLabel  string        `json:"tz_label,omitempty"`
	TZOffset int           `json:"tz_offset,omitempty"`
}

// ExportProfile represents a user profile in the export
type ExportProfile struct {
	DisplayName        string `json:"display_name,omitempty"`
	RealName           string `json:"real_name,omitempty"`
	RealNameNormalized string `json:"real_name_normalized,omitempty"`
	Email              string `json:"email,omitempty"`
	Image24            string `json:"image_24,omitempty"`
	Image32            string `json:"image_32,omitempty"`
	Image48            string `json:"image_48,omitempty"`
	Image72            string `json:"image_72,omitempty"`
	Image192           string `json:"image_192,omitempty"`
	Image512           string `json:"image_512,omitempty"`
	Title              string `json:"title,omitempty"`
	Phone              string `json:"phone,omitempty"`
	Skype              string `json:"skype,omitempty"`
	StatusText         string `json:"status_text,omitempty"`
	StatusEmoji        string `json:"status_emoji,omitempty"`
	Team               string `json:"team,omitempty"`
}

// ExportStatistics contains statistics about the export
type ExportStatistics struct {
	TotalMessages    int                 `json:"total_messages"`
	TotalThreads     int                 `json:"total_threads"`
	TotalReplies     int                 `json:"total_replies"`
	TotalUsers       int                 `json:"total_users"`
	TotalAttachments int                 `json:"total_attachments"`
	TotalFiles       int                 `json:"total_files"`
	TotalReactions   int                 `json:"total_reactions"`
	MessagesByUser   map[string]int      `json:"messages_by_user"`
	MessagesByDate   map[string]int      `json:"messages_by_date"`
	TopReactions     []ReactionStat      `json:"top_reactions"`
	ExportDuration   time.Duration       `json:"export_duration"`
	ProcessingTime   ProcessingTimeStats `json:"processing_time"`
}

// ReactionStat represents statistics for a reaction
type ReactionStat struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ProcessingTimeStats contains timing information for the export process
type ProcessingTimeStats struct {
	ChannelFetch   time.Duration `json:"channel_fetch"`
	MessageFetch   time.Duration `json:"message_fetch"`
	ThreadFetch    time.Duration `json:"thread_fetch"`
	UserFetch      time.Duration `json:"user_fetch"`
	DataProcessing time.Duration `json:"data_processing"`
	FileGeneration time.Duration `json:"file_generation"`
}

// ExportOptions contains configuration for the export process
type ExportOptions struct {
	ChannelID        string     `json:"channel_id"`
	ChannelName      string     `json:"channel_name,omitempty"`
	IncludeThreads   bool       `json:"include_threads"`
	IncludeFiles     bool       `json:"include_files"`
	IncludeReactions bool       `json:"include_reactions"`
	DateFrom         *time.Time `json:"date_from,omitempty"`
	DateTo           *time.Time `json:"date_to,omitempty"`
	OutputFile       string     `json:"output_file"`
	Format           string     `json:"format"`                // "json", "json-pretty", "json-compact"
	Compression      string     `json:"compression,omitempty"` // "gzip", "zip", "none"
}

// ExportProgress represents the current state of an export operation
type ExportProgress struct {
	Stage           string        `json:"stage"`
	CurrentStep     string        `json:"current_step"`
	Progress        float64       `json:"progress"` // 0.0 to 1.0
	MessagesTotal   int           `json:"messages_total"`
	MessagesCurrent int           `json:"messages_current"`
	ThreadsTotal    int           `json:"threads_total"`
	ThreadsCurrent  int           `json:"threads_current"`
	ElapsedTime     time.Duration `json:"elapsed_time"`
	EstimatedTotal  time.Duration `json:"estimated_total"`
	Error           string        `json:"error,omitempty"`
}

// ExportResult contains the result of an export operation
type ExportResult struct {
	Success    bool             `json:"success"`
	OutputFile string           `json:"output_file"`
	FileSize   int64            `json:"file_size"`
	Statistics ExportStatistics `json:"statistics"`
	Duration   time.Duration    `json:"duration"`
	Error      string           `json:"error,omitempty"`
	Warnings   []string         `json:"warnings,omitempty"`
}

// ParseSlackTimestamp parses a Slack timestamp string to time.Time
func ParseSlackTimestamp(ts string) (time.Time, error) {
	if ts == "" {
		return time.Time{}, nil
	}

	// Slack timestamps are Unix timestamps with microseconds (e.g., "1704067200.123456")
	timestamp, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(timestamp), int64((timestamp-float64(int64(timestamp)))*1e9)), nil
}

// ConvertToExportMessage converts a Slack message to export format
func ConvertToExportMessage(msg Message) ExportMessage {
	exportMsg := ExportMessage{
		ID:              msg.Timestamp, // Use timestamp as ID for now
		User:            msg.User,
		Text:            msg.Text,
		Type:            msg.Type,
		Subtype:         msg.Subtype,
		ThreadTimestamp: msg.ThreadTS,
		ReplyCount:      msg.ReplyCount,
	}

	// Parse timestamp
	if timestamp, err := ParseSlackTimestamp(msg.Timestamp); err == nil {
		exportMsg.Timestamp = timestamp
	}

	// Convert edit info
	if msg.Edited != nil {
		if editTime, err := ParseSlackTimestamp(msg.Edited.Timestamp); err == nil {
			exportMsg.Edited = &EditInfo{
				User:      msg.Edited.User,
				Timestamp: editTime,
			}
		}
	}

	// Convert attachments
	for _, att := range msg.Attachments {
		exportAtt := ExportAttachment{
			ID:       strconv.Itoa(att.ID),
			Title:    att.Title,
			Text:     att.Text,
			Fallback: att.Fallback,
			Color:    att.Color,
			ImageURL: att.ImageURL,
			ThumbURL: att.ThumbURL,
		}

		exportMsg.Attachments = append(exportMsg.Attachments, exportAtt)
	}

	// Convert files
	for _, file := range msg.Files {
		exportFile := ExportFile{
			ID:         file.ID,
			Name:       file.Name,
			Title:      file.Title,
			Mimetype:   file.Mimetype,
			Filetype:   file.Filetype,
			Size:       file.Size,
			URLPrivate: file.URL,
		}

		// Set a default timestamp for files (current time if not available)
		exportFile.Timestamp = time.Now()

		exportMsg.Files = append(exportMsg.Files, exportFile)
	}

	// Convert reactions
	for _, reaction := range msg.Reactions {
		exportMsg.Reactions = append(exportMsg.Reactions, ExportReaction{
			Name:  reaction.Name,
			Count: reaction.Count,
			Users: reaction.Users,
		})
	}

	// Convert thread replies recursively
	for _, reply := range msg.Thread {
		exportMsg.Replies = append(exportMsg.Replies, ConvertToExportMessage(reply))
	}

	return exportMsg
}

// ConvertToExportUser converts a Slack user to export format
func ConvertToExportUser(user User) ExportUser {
	return ExportUser{
		ID:       user.ID,
		Name:     user.Name,
		RealName: user.RealName,
		IsBot:    user.IsBot,
		Deleted:  user.Deleted,
		Profile: ExportProfile{
			DisplayName: user.Profile.DisplayName,
			RealName:    user.Profile.RealName,
			Email:       user.Profile.Email,
			Image24:     user.Profile.Image24,
			Image32:     user.Profile.Image32,
			Image48:     user.Profile.Image48,
			Image72:     user.Profile.Image72,
			Image192:    user.Profile.Image192,
			Image512:    user.Profile.Image512,
		},
	}
}
