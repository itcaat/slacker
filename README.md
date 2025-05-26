# Slacker - Slack CLI Client

Slacker is a powerful command-line interface for Slack built with Go and Bubble Tea. It provides an intuitive text-based user interface for browsing channels, viewing messages, and exporting channel histories with full thread structure preservation.

## ✨ Features

- 🖥️ **Interactive TUI** - Beautiful text-based interface for browsing channels and messages
- 📤 **Channel Export** - Export complete channel histories to structured JSON files
- 🧵 **Thread Preservation** - Maintains thread structure in exports with parent-child relationships
- 🔐 **Secure Authentication** - OAuth2 token-based authentication with secure storage
- 📊 **Rich Statistics** - Detailed export statistics including message counts, reactions, and user activity
- 🎨 **Multiple Formats** - JSON, pretty JSON, compact JSON with optional gzip compression
- 📅 **Date Filtering** - Export specific date ranges
- ⚡ **Progress Indication** - Real-time progress bars and stage indicators for exports

## 🚀 Quick Start

### 1. Download and Build

```bash
git clone https://github.com/itcaat/slacker.git
cd slacker
go build -o slacker .
```

### 2. Get Your Slack Token

To use Slacker, you need a Slack Bot User OAuth Token. Follow these steps:

#### Step 1: Create a Slack App
1. Go to **https://api.slack.com/apps**
2. Click **"Create New App"**
3. Choose **"From scratch"**
4. Give your app a name (e.g., "Slacker CLI")
5. Select your workspace
6. Click **"Create App"**

#### Step 2: Add Required Permissions
1. In your app settings, go to **"OAuth & Permissions"**
2. Scroll to **"Bot Token Scopes"** and add these scopes:
   - `channels:history` - Read messages in public channels
   - `channels:read` - View basic information about public channels
   - `groups:history` - Read messages in private channels
   - `groups:read` - View basic information about private channels
   - `users:read` - View people in the workspace

#### Step 3: Install the App
1. Scroll up to **"OAuth Tokens for Your Workspace"**
2. Click **"Install to Workspace"**
3. Review permissions and click **"Allow"**
4. Copy the **"Bot User OAuth Token"** (starts with `xoxb-`)

### 3. Authenticate

```bash
# Method 1: Using the auth command (recommended)
./slacker auth xoxb-your-token-here

# Method 2: Using environment variable
export SLACKER_SLACK_TOKEN=xoxb-your-token-here
```

### 4. Test Authentication

```bash
./slacker auth test
```

You should see:
```
✅ Authentication successful!
   User: your-username
   Team: Your Workspace Name
✅ Channel access successful! Found X channels
```

## 📖 Usage

### Interactive TUI Mode

Launch the interactive interface:
```bash
./slacker tui
```

**TUI Controls:**
- `↑/↓` or `k/j` - Navigate channels/messages
- `Enter` - Select channel or view message details
- `e` - Export current channel
- `r` - Refresh data
- `Esc` - Go back
- `q` - Quit

### Command Line Interface

#### List Channels
```bash
# Basic list
./slacker channels list

# With filtering
./slacker channels list --archived --format json

# Private channels only
./slacker channels list --private-only --verbose
```

#### View Messages
```bash
# View recent messages
./slacker messages --channel general

# View with threads
./slacker messages --channel general --threads --limit 50

# JSON output
./slacker messages --channel general --format json
```

#### Export Channel History
```bash
# Basic export
./slacker export --channel general

# Advanced export with options
./slacker export --channel general \
  --format json-pretty \
  --compress gzip \
  --output general-backup.json \
  --verbose

# Date range export
./slacker export --channel general \
  --from 2024-01-01 \
  --to 2024-01-31 \
  --threads \
  --files \
  --reactions

# Minimal export (no files/reactions)
./slacker export --channel general \
  --no-files \
  --no-reactions \
  --format json-compact
```

## 📋 Export Options

| Flag | Description | Default |
|------|-------------|---------|
| `--channel` | Channel name to export | Required |
| `--output` | Output file path | `<channel>-export-<timestamp>.json` |
| `--format` | Output format: `json`, `json-pretty`, `json-compact` | `json-pretty` |
| `--compress` | Compression: `gzip` or `none` | `none` |
| `--threads` | Include thread replies | `true` |
| `--files` | Include file attachments | `true` |
| `--reactions` | Include message reactions | `true` |
| `--from` | Start date (YYYY-MM-DD) | All messages |
| `--to` | End date (YYYY-MM-DD) | All messages |
| `--verbose` | Detailed progress output | `false` |

## 📁 Export Format

Exported JSON files contain:

```json
{
  "export_info": {
    "exported_at": "2024-01-15T10:30:00Z",
    "slacker_version": "1.0.0",
    "export_format": "json-pretty",
    "include_threads": true
  },
  "channel": {
    "id": "C1234567890",
    "name": "general",
    "is_private": false,
    "topic": "General discussion",
    "num_members": 42
  },
  "messages": [
    {
      "id": "1234567890.123456",
      "user": "U1234567890",
      "text": "Hello world!",
      "timestamp": "2024-01-15T09:00:00Z",
      "replies": [
        {
          "id": "1234567890.123457",
          "user": "U0987654321",
          "text": "Hi there!",
          "timestamp": "2024-01-15T09:01:00Z"
        }
      ],
      "reactions": [
        {
          "name": "thumbsup",
          "count": 3,
          "users": ["U1111111111", "U2222222222", "U3333333333"]
        }
      ]
    }
  ],
  "users": {
    "U1234567890": {
      "id": "U1234567890",
      "name": "john.doe",
      "real_name": "John Doe",
      "profile": {
        "display_name": "John",
        "email": "john@example.com"
      }
    }
  },
  "statistics": {
    "total_messages": 150,
    "total_threads": 25,
    "total_users": 42,
    "total_reactions": 89
  }
}
```

## 🔧 Configuration

Slacker stores configuration in `~/.slacker.yaml`:

```yaml
slack:
  token: "xoxb-your-token-here"
debug: false
export:
  default_output_dir: "./exports"
  include_threads: true
  include_users: true
```

## 🛠️ Development

### Requirements
- Go 1.22 or later
- Access to a Slack workspace

### Build from Source
```bash
git clone https://github.com/itcaat/slacker.git
cd slacker
go mod download
go build -o slacker .
```

### Run Tests
```bash
go test ./...
```

### Project Structure
```
slacker/
├── cmd/                 # CLI commands
├── internal/
│   ├── api/            # Slack API client
│   ├── config/         # Configuration management
│   ├── ui/             # TUI components
│   └── usecase/        # Business logic
├── models/             # Data structures
└── main.go
```

## 🔒 Security

- **Token Storage**: Tokens are stored securely in your home directory
- **Permissions**: Only requests the minimum required Slack permissions
- **No Password Storage**: Uses OAuth2 tokens instead of passwords
- **Local Processing**: All data processing happens locally

## 🐛 Troubleshooting

### Authentication Issues
```bash
# Test your token
./slacker auth test

# Check token permissions at https://api.slack.com/apps
# Ensure your bot is added to channels you want to access
```

### Permission Errors
- Make sure your Slack app has the required scopes
- Invite the bot to private channels you want to access
- Check that the bot is installed in your workspace

### Export Issues
```bash
# Use verbose mode for detailed error information
./slacker export --channel general --verbose

# Check available channels
./slacker channels list
```

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📞 Support

- 🐛 **Issues**: Report bugs on GitHub Issues
- 💡 **Feature Requests**: Submit ideas on GitHub Issues
- 📖 **Documentation**: Check this README and command help (`./slacker --help`)

---

**Made with ❤️ using Go and Bubble Tea**
