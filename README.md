# frefresh

Interactive CLI for refreshing tables in Microsoft Fabric semantic models via the Power BI Enhanced Refresh API.

Select a customer, environment, semantic model, and tables — then trigger and monitor the refresh from your terminal.

## Install

### macOS (Homebrew)

```bash
brew install DanielAndreassen97/tap/frefresh
```

### Windows (Scoop)

```powershell
scoop bucket add frefresh https://github.com/DanielAndreassen97/scoop-bucket
scoop install frefresh
```

### Go install

```bash
go install github.com/DanielAndreassen97/frefresh@latest
```

### Download binary

Download from [GitHub Releases](https://github.com/DanielAndreassen97/frefresh/releases/latest) and add to your PATH.

## Usage

```bash
# Interactive main menu
frefresh

# Direct commands
frefresh add        # Add a new customer
frefresh edit       # Edit an existing customer
frefresh remove     # Remove a customer
frefresh list       # List configured customers
frefresh refresh    # Start a refresh

# Demo mode (mock API, fake data)
frefresh --demo
```

## Configuration

Config is stored at `~/.config/frefresh/config.json` (macOS/Linux) or `%APPDATA%\frefresh\config.json` (Windows).

Each customer entry has:
- **path** — local folder containing `.SemanticModel` directories
- **workspace_pattern** — Power BI workspace name with `{env}` placeholder (e.g. `DP - {env} - SemMod`)
- **environments** — list of environments (e.g. `["DEV", "TEST", "PROD"]`)

## How it works

1. Discovers semantic models by scanning for `.platform` files
2. Discovers refreshable tables from `.tmdl` files (excludes calculated tables and calculation groups)
3. Authenticates via browser-based OAuth2 with Microsoft Entra ID
4. Triggers an enhanced refresh via the Power BI REST API
5. Polls until completion and displays the result
