# Claude Code Instructions for lark

A Go CLI tool for interacting with Lark APIs, designed for use by Claude Code.

## Project Structure

```
cmd/lark/         # Main entry point
internal/
  api/            # Lark API client
  auth/           # OAuth authentication
  cmd/            # Cobra command implementations
  config/         # Configuration handling
  conflicts/      # Conflict detection
  output/         # JSON/human-readable output formatting
  time/           # Date/time parsing
```

## Build & Test

```bash
make build        # Build binary to ./lark
make test         # Run tests
make deps         # Tidy and download dependencies
make install      # Install to $GOPATH/bin
```

## Code Conventions

- JSON output by default
- Error responses use structured format: `{"error": true, "code": "...", "message": "..."}`
- Date parsing supports ISO 8601 formats
- Config in `.lark/config.yaml`, secrets via `LARK_APP_SECRET` env var
- Use Cobra for commands, Viper for config

## Commands

See `USAGE.md` for full CLI documentation. Main commands:

### Auth
- `auth login|status|logout` - Authentication

### Calendar (`cal`)
- `cal list` - List events (supports `--week`, `--from`, `--to`)
- `cal show <id>` - Show event details
- `cal create` - Create event
- `cal update <id>` - Update event
- `cal delete <id>` - Delete event
- `cal search <query>` - Search events
- `cal freebusy` - Query availability
- `cal lookup-user` - Get user ID from email
- `cal common-freetime` - Find mutual availability
- `cal rsvp <id>` - Accept/decline invitation

### Contacts (`contact`)
- `contact get <user_id>` - Get user info by ID
- `contact list-dept [dept_id]` - List users in department
- `contact search-dept <query>` - Search departments by name

### Documents (`doc`)
- `doc list [folder_token]` - List items in a Drive folder
- `doc get <document_id>` - Get document content as markdown
- `doc blocks <document_id>` - Get document block structure (JSON)
- `doc wiki <node_token>` - Resolve wiki node to document token
