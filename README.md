# ContextKeeper Agent

A lightweight Go agent that automatically captures and syncs AI tool interactions with [ContextKeeper.dev](https://contextkeeper.dev).

## Features

- **Automatic Detection**: Monitors terminal and processes for AI tool usage (Claude, Gemini, ChatGPT, etc.)
- **Anonymous Usage**: Works without signup - get 50 free session saves
- **Cross-Platform**: Supports macOS, Linux, and Windows
- **Low Resource Usage**: Minimal CPU and memory footprint
- **Local Buffering**: Continues working offline, syncs when connected
- **Batch Uploads**: Efficient syncing with ContextKeeper.dev
- **VS Code Integration**: Works with the ContextKeeper VS Code extension

## Quick Start

### Installation

**Option 1: Download Binary**
```bash
# Download and run install script
curl -fsSL https://raw.githubusercontent.com/carsor007/contextkeeper-agent/main/scripts/install.sh | bash
```

**Option 2: Build from Source**
```bash
git clone https://github.com/carsor007/contextkeeper-agent.git
cd contextkeeper-agent
./scripts/build.sh
```

### Usage

```bash
# Start the agent
contextkeeper-agent

# Run as daemon
contextkeeper-agent --daemon

# Check usage
contextkeeper-agent --usage

# Show version
contextkeeper-agent --version
```

## How It Works

1. **Monitor**: The agent monitors your terminal and running processes for AI tool activity
2. **Parse**: AI outputs are parsed and structured (tool type, content, project context)
3. **Buffer**: Sessions are stored locally for batch uploading
4. **Sync**: Automatically syncs with ContextKeeper.dev (respects 50-session limit for anonymous users)

## Architecture

```
┌─────────────────┐    ┌──────────────┐    ┌─────────────────┐
│   VS Code       │    │  Go Agent    │    │ ContextKeeper   │
│   Extension     │───▶│  (Local)     │───▶│ Web App         │
│                 │    │              │    │                 │
└─────────────────┘    └──────────────┘    └─────────────────┘
```

## Supported AI Tools

- **Claude** (Anthropic CLI, Claude Code)
- **Gemini** (Google AI)
- **ChatGPT** (OpenAI CLI tools)
- **Aider** (AI pair programming)
- **Cursor** (AI code editor)
- **GitHub Copilot**
- Generic AI tool detection

## Configuration

The agent stores configuration in `~/.contextkeeper/config.yaml`:

```yaml
server_url: "https://contextkeeper.dev"
api_key: ""  # For authenticated users
log_level: "info"
max_sessions: 100
upload_batch: 5
```

## Anonymous vs Authenticated Usage

### Anonymous (Default)
- ✅ 50 free session saves
- ✅ No signup required
- ✅ Local storage and sync
- ❌ Limited to 50 sessions

### Authenticated (Pro)
- ✅ Unlimited sessions
- ✅ Advanced features
- ✅ Cross-device sync
- ✅ Premium support

**Upgrade**: Get your API key from [ContextKeeper.dev/pricing](https://contextkeeper.dev/pricing)

## Development

### Prerequisites
- Go 1.24+
- Git

### Building
```bash
# Clone repository
git clone https://github.com/carsor007/contextkeeper-agent.git
cd contextkeeper-agent

# Build for current platform
go build -o bin/contextkeeper-agent ./cmd/agent

# Build for all platforms
./scripts/build.sh
```

### Project Structure
```
contextkeeper-agent/
├── cmd/agent/          # Main application
├── internal/
│   ├── agent/          # Core agent logic
│   ├── parser/         # AI output parsers
│   ├── sync/           # ContextKeeper.dev API client
│   └── platform/       # Platform-specific code
├── pkg/types/          # Shared types
├── scripts/            # Build and install scripts
└── configs/            # Configuration files
```

### Running Tests
```bash
go test ./...
```

## Integration

### VS Code Extension
The agent works seamlessly with the [ContextKeeper VS Code extension](https://github.com/AIContextKeeper/vscode):

1. Install the VS Code extension
2. Start the ContextKeeper agent
3. AI interactions are automatically captured from both VS Code and terminal

### API Integration
The agent exposes a local HTTP API for integrations:

```bash
# Health check
curl http://localhost:8080/health

# Get current usage
curl http://localhost:8080/api/usage

# Submit session (from VS Code extension)
curl -X POST http://localhost:8080/api/session \
  -H "Content-Type: application/json" \
  -d '{"title": "My Session", "content": "..."}'
```

## Privacy & Security

- **Local-First**: All processing happens locally
- **Minimal Data**: Only AI interaction summaries are sent to ContextKeeper.dev
- **No Code Upload**: Your source code never leaves your machine
- **Anonymous by Default**: No personal information required

## Troubleshooting

### Agent Not Detecting AI Tools
1. Check the agent is running: `contextkeeper-agent --usage`
2. Verify AI tools are in supported list
3. Check logs for detection patterns

### Sync Issues
1. Test connection: `curl https://contextkeeper.dev/api/summaries`
2. Check usage limit: `contextkeeper-agent --usage`
3. Review agent logs for errors

### Performance Issues
1. Reduce buffer size in config
2. Increase flush interval
3. Lower log level to 'error'

## Contributing

This is proprietary software. For feature requests or bug reports, please contact support at support@contextkeeper.dev

## License

Proprietary software - see [LICENSE](LICENSE) file for details.

## Links

- **Website**: [ContextKeeper.dev](https://contextkeeper.dev)
- **VS Code Extension**: [GitHub](https://github.com/AIContextKeeper/vscode)
- **Web App**: [GitHub](https://github.com/carsor007/ContextKeeper) (Private)
- **Documentation**: [Docs](https://contextkeeper.dev/docs)
- **Support**: [Issues](https://github.com/carsor007/contextkeeper-agent/issues)