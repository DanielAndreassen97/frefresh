# frefresh

Interactive CLI for refreshing tables in Microsoft Fabric semantic models via the Power BI Enhanced Refresh API.

![demo](demo.gif)

## Features

- **Interactive TUI** — navigate with arrow keys, number keys, or keyboard shortcuts
- **Semantic model discovery** — automatically finds models and tables from your Fabric repo (`.tmdl` files)
- **Per-customer config** — manage multiple customers with separate workspaces and environments
- **OAuth2 browser auth** — authenticates via Microsoft Entra ID with per-customer token caching in your OS keychain
- **Cross-platform** — macOS, Linux, Windows

## Install

### Homebrew (macOS/Linux)

```sh
brew install DanielAndreassen97/tap/frefresh
```

### Scoop (Windows)

```powershell
scoop bucket add frefresh https://github.com/DanielAndreassen97/scoop-bucket
scoop install frefresh
```

### Go

```sh
go install github.com/DanielAndreassen97/frefresh@latest
```

### Download binary

Download from [GitHub Releases](https://github.com/DanielAndreassen97/frefresh/releases/latest) and add to your PATH.

## Usage

```sh
frefresh              # Interactive menu
frefresh add          # Add a customer
frefresh edit         # Edit a customer
frefresh remove       # Remove a customer
frefresh list         # List customers
frefresh refresh      # Refresh tables
frefresh logout       # Clear cached credentials
frefresh version      # Show version
```

## Configuration

Config is stored at `~/.config/frefresh/config.json` (macOS/Linux) or `%APPDATA%\frefresh\config.json` (Windows).

Each customer needs:
- **Path** — local path to the Fabric repo containing `.SemanticModel` folders
- **Workspace pattern** — Power BI workspace name with `{env}` placeholder (e.g. `DP - {env} - SemMod`)
- **Environments** — list of environments (e.g. `DEV, TEST, PROD`)

## How it works

1. Discovers semantic models by scanning for `.platform` files
2. Discovers refreshable tables from `.tmdl` files (excludes calculated tables and calculation groups)
3. Authenticates via browser-based OAuth2 with Microsoft Entra ID
4. Triggers an enhanced refresh via the Power BI REST API
5. Polls until completion and displays the result

## Authentication

Uses OAuth2 Authorization Code Flow with the Azure CLI public client ID. On first use per customer, a browser window opens for Microsoft login. Tokens are cached in your OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service) and silently refreshed — you typically authenticate once and stay logged in for months.

Use `frefresh logout` to clear all cached credentials.

## License

MIT
