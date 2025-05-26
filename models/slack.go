package models

import "time"

// Channel represents a Slack channel
type Channel struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	IsChannel  bool   `json:"is_channel"`
	IsGroup    bool   `json:"is_group"`
	IsIM       bool   `json:"is_im"`
	IsMember   bool   `json:"is_member"`
	IsPrivate  bool   `json:"is_private"`
	IsArchived bool   `json:"is_archived"`
	NumMembers int    `json:"num_members"`
	Topic      Topic  `json:"topic"`
	Purpose    Topic  `json:"purpose"`
	Created    int64  `json:"created"`
	Creator    string `json:"creator"`
}

// Topic represents channel topic or purpose
type Topic struct {
	Value   string `json:"value"`
	Creator string `json:"creator"`
	LastSet int64  `json:"last_set"`
}

// Message represents a Slack message
type Message struct {
	Type        string       `json:"type"`
	User        string       `json:"user"`
	Text        string       `json:"text"`
	Timestamp   string       `json:"ts"`
	ThreadTS    string       `json:"thread_ts,omitempty"`
	ReplyCount  int          `json:"reply_count,omitempty"`
	Replies     []Reply      `json:"replies,omitempty"`
	Thread      []Message    `json:"thread,omitempty"` // For our export format
	Attachments []Attachment `json:"attachments,omitempty"`
	Files       []File       `json:"files,omitempty"`
	Reactions   []Reaction   `json:"reactions,omitempty"`
	Edited      *Edited      `json:"edited,omitempty"`
	BotID       string       `json:"bot_id,omitempty"`
	Username    string       `json:"username,omitempty"`
	Subtype     string       `json:"subtype,omitempty"`
}

// Reply represents a thread reply reference
type Reply struct {
	User      string `json:"user"`
	Timestamp string `json:"ts"`
}

// Attachment represents a message attachment
type Attachment struct {
	ID       int    `json:"id"`
	Color    string `json:"color"`
	Fallback string `json:"fallback"`
	Title    string `json:"title"`
	Text     string `json:"text"`
	ImageURL string `json:"image_url"`
	ThumbURL string `json:"thumb_url"`
}

// File represents an uploaded file
type File struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Title    string `json:"title"`
	Mimetype string `json:"mimetype"`
	Filetype string `json:"filetype"`
	Size     int    `json:"size"`
	URL      string `json:"url_private"`
}

// Reaction represents a message reaction
type Reaction struct {
	Name  string   `json:"name"`
	Users []string `json:"users"`
	Count int      `json:"count"`
}

// Edited represents edit information
type Edited struct {
	User      string `json:"user"`
	Timestamp string `json:"ts"`
}

// User represents a Slack user
type User struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	RealName string  `json:"real_name"`
	Profile  Profile `json:"profile"`
	IsBot    bool    `json:"is_bot"`
	Deleted  bool    `json:"deleted"`
}

// Profile represents user profile information
type Profile struct {
	DisplayName string `json:"display_name"`
	RealName    string `json:"real_name"`
	Email       string `json:"email"`
	Image24     string `json:"image_24"`
	Image32     string `json:"image_32"`
	Image48     string `json:"image_48"`
	Image72     string `json:"image_72"`
	Image192    string `json:"image_192"`
	Image512    string `json:"image_512"`
}

// ExportData represents the structure for JSON export
type ExportData struct {
	Channel    string    `json:"channel"`
	Messages   []Message `json:"messages"`
	Users      []User    `json:"users,omitempty"`
	ExportedAt time.Time `json:"exported_at"`
	ExportedBy string    `json:"exported_by"`
}

// SlackConfig represents Slack API configuration
type SlackConfig struct {
	Token     string `json:"token"`
	AppToken  string `json:"app_token,omitempty"`
	BotToken  string `json:"bot_token,omitempty"`
	UserToken string `json:"user_token,omitempty"`
}
