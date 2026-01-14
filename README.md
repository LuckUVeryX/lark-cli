# lark

A CLI tool for interacting with Lark/Feishu APIs, designed for use with Claude Code and other AI assistants.

## Why This Tool?

The official Lark MCP server exists, but its tools are not token-efficient. Each tool call returns verbose responses that consume significant context window space when used with AI assistants.

This CLI addresses that by:

- **Returning compact JSON** - Structured output optimized for programmatic consumption
- **Providing markdown conversion** - Documents are converted to markdown (~2-3x smaller than raw block structures)
- **Supporting selective queries** - Fetch only what you need (e.g., just event IDs, just document titles)

The result: AI assistants can interact with Lark using fewer tokens, leaving more context for actual work.

## Features

- **Calendar** - List, create, update, delete events; check availability; find common free time; RSVP
- **Contacts** - Look up users by ID, search by name, list department members
- **Documents** - Read documents as markdown, list folders, resolve wiki nodes, get comments
- **Messages** - Retrieve chat history, download attachments

## Quick Start

1. Create a Lark app at https://open.larksuite.com with appropriate permissions
2. Copy `config.example.yaml` to `.lark/config.yaml` and add your App ID
3. Set `LARK_APP_SECRET` environment variable
4. Run `./lark auth login` to authenticate
5. Start using: `./lark cal list --week`

See [USAGE.md](USAGE.md) for full documentation.

## Building

```bash
make build    # Build binary to ./lark
make test     # Run tests
make install  # Install to $GOPATH/bin
```

## Usage with Claude Code

This tool is designed to be invoked via Claude Code skills. Example skill configuration:

```yaml
name: calendar
description: Manage Lark calendar
---
Run from your workspace with:
LARK_CONFIG_DIR=/path/to/.lark ./lark cal <command>
```

The JSON output format makes it straightforward for AI assistants to parse responses and take action.

## License

MIT - see [LICENSE](LICENSE)
