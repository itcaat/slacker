package api

import (
	"context"
	"fmt"
	"log"

	"github.com/itcaat/slacker/models"
	"github.com/slack-go/slack"
)

// SlackClient wraps the Slack API client with our custom functionality
type SlackClient struct {
	client *slack.Client
	token  string
	debug  bool
}

// NewSlackClient creates a new Slack API client
func NewSlackClient(token string, debug bool) *SlackClient {
	var client *slack.Client
	if debug {
		client = slack.New(token, slack.OptionDebug(true))
	} else {
		client = slack.New(token)
	}

	return &SlackClient{
		client: client,
		token:  token,
		debug:  debug,
	}
}

// TestAuth tests the authentication with Slack API
func (sc *SlackClient) TestAuth(ctx context.Context) (*slack.AuthTestResponse, error) {
	if sc.debug {
		log.Println("Testing Slack authentication...")
	}

	response, err := sc.client.AuthTestContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if sc.debug {
		log.Printf("Authentication successful - User: %s, Team: %s", response.User, response.Team)
	}

	return response, nil
}

// GetChannels retrieves all channels the user is a member of
func (sc *SlackClient) GetChannels(ctx context.Context) ([]models.Channel, error) {
	if sc.debug {
		log.Println("Fetching channels...")
	}

	// Get public channels
	channels, _, err := sc.client.GetConversationsContext(ctx, &slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel"},
		Limit: 1000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}

	var result []models.Channel
	for _, ch := range channels {
		// Only include channels the user is a member of
		if ch.IsMember {
			channel := models.Channel{
				ID:         ch.ID,
				Name:       ch.Name,
				IsChannel:  ch.IsChannel,
				IsGroup:    ch.IsGroup,
				IsIM:       ch.IsIM,
				IsMember:   ch.IsMember,
				IsPrivate:  ch.IsPrivate,
				IsArchived: ch.IsArchived,
				NumMembers: ch.NumMembers,
				Created:    int64(ch.Created),
				Creator:    ch.Creator,
				Topic: models.Topic{
					Value:   ch.Topic.Value,
					Creator: ch.Topic.Creator,
					LastSet: int64(ch.Topic.LastSet),
				},
				Purpose: models.Topic{
					Value:   ch.Purpose.Value,
					Creator: ch.Purpose.Creator,
					LastSet: int64(ch.Purpose.LastSet),
				},
			}
			result = append(result, channel)
		}
	}

	if sc.debug {
		log.Printf("Found %d channels", len(result))
	}

	return result, nil
}

// GetChannelHistory retrieves message history for a specific channel
func (sc *SlackClient) GetChannelHistory(ctx context.Context, channelID string, limit int, cursor string) ([]models.Message, string, error) {
	if sc.debug {
		log.Printf("Fetching history for channel %s (limit: %d, cursor: %s)", channelID, limit, cursor)
	}

	params := &slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     limit,
		Cursor:    cursor,
	}

	response, err := sc.client.GetConversationHistoryContext(ctx, params)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get channel history: %w", err)
	}

	var messages []models.Message
	for _, msg := range response.Messages {
		message := sc.convertSlackMessage(msg)
		messages = append(messages, message)
	}

	if sc.debug {
		log.Printf("Retrieved %d messages", len(messages))
	}

	return messages, response.ResponseMetaData.NextCursor, nil
}

// GetThreadReplies retrieves replies for a threaded message
func (sc *SlackClient) GetThreadReplies(ctx context.Context, channelID, threadTS string) ([]models.Message, error) {
	if sc.debug {
		log.Printf("Fetching thread replies for %s in channel %s", threadTS, channelID)
	}

	params := &slack.GetConversationRepliesParameters{
		ChannelID: channelID,
		Timestamp: threadTS,
	}

	messages, _, _, err := sc.client.GetConversationRepliesContext(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread replies: %w", err)
	}

	var replies []models.Message
	// Skip the first message as it's the parent message
	for i := 1; i < len(messages); i++ {
		reply := sc.convertSlackMessage(messages[i])
		replies = append(replies, reply)
	}

	if sc.debug {
		log.Printf("Retrieved %d thread replies", len(replies))
	}

	return replies, nil
}

// GetUsers retrieves user information for the workspace
func (sc *SlackClient) GetUsers(ctx context.Context) ([]models.User, error) {
	if sc.debug {
		log.Println("Fetching users...")
	}

	users, err := sc.client.GetUsersContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	var result []models.User
	for _, user := range users {
		u := models.User{
			ID:       user.ID,
			Name:     user.Name,
			RealName: user.RealName,
			IsBot:    user.IsBot,
			Deleted:  user.Deleted,
			Profile: models.Profile{
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
		result = append(result, u)
	}

	if sc.debug {
		log.Printf("Retrieved %d users", len(result))
	}

	return result, nil
}

// convertSlackMessage converts a slack.Message to our models.Message
func (sc *SlackClient) convertSlackMessage(msg slack.Message) models.Message {
	message := models.Message{
		Type:       msg.Type,
		User:       msg.User,
		Text:       msg.Text,
		Timestamp:  msg.Timestamp,
		ThreadTS:   msg.ThreadTimestamp,
		ReplyCount: msg.ReplyCount,
		BotID:      msg.BotID,
		Username:   msg.Username,
		Subtype:    msg.SubType,
	}

	// Convert attachments
	for _, att := range msg.Attachments {
		attachment := models.Attachment{
			ID:       att.ID,
			Color:    att.Color,
			Fallback: att.Fallback,
			Title:    att.Title,
			Text:     att.Text,
			ImageURL: att.ImageURL,
			ThumbURL: att.ThumbURL,
		}
		message.Attachments = append(message.Attachments, attachment)
	}

	// Convert files
	for _, file := range msg.Files {
		f := models.File{
			ID:       file.ID,
			Name:     file.Name,
			Title:    file.Title,
			Mimetype: file.Mimetype,
			Filetype: file.Filetype,
			Size:     file.Size,
			URL:      file.URLPrivate,
		}
		message.Files = append(message.Files, f)
	}

	// Convert reactions
	for _, reaction := range msg.Reactions {
		r := models.Reaction{
			Name:  reaction.Name,
			Users: reaction.Users,
			Count: reaction.Count,
		}
		message.Reactions = append(message.Reactions, r)
	}

	// Convert edited info
	if msg.Edited != nil {
		message.Edited = &models.Edited{
			User:      msg.Edited.User,
			Timestamp: msg.Edited.Timestamp,
		}
	}

	return message
}

// GetChannelByName finds a channel by name
func (sc *SlackClient) GetChannelByName(ctx context.Context, channelName string) (*models.Channel, error) {
	channels, err := sc.GetChannels(ctx)
	if err != nil {
		return nil, err
	}

	for _, channel := range channels {
		if channel.Name == channelName {
			return &channel, nil
		}
	}

	return nil, fmt.Errorf("channel '%s' not found", channelName)
}
