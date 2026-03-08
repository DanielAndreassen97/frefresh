# frefresh

Interactive CLI for refreshing tables in Microsoft Fabric semantic models via the Power BI Enhanced Refresh API.

![demo](demo.gif)

## Features

- **Interactive TUI** — navigate with arrow keys, number keys, or keyboard shortcuts
- **Live table discovery** — queries the deployed model via the Fabric API to find refreshable tables
- **Smart filtering** — automatically excludes calculated tables, calculation groups, and measure-only tables
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

## Prerequisites

- A workspace on **Fabric (F SKU)**, **Premium (P SKU)**, **Premium Per User (PPU)**, or **Embedded (A/EM SKU)** capacity. Table-level refresh uses the [Enhanced Refresh API](https://learn.microsoft.com/en-us/power-bi/connect-data/asynchronous-refresh) which is not available on Power BI Pro.
- An Entra ID account with Contributor (or higher) permissions on the workspace.

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
frefresh help         # Show available commands
```

## Configuration

Config is stored at `~/.config/frefresh/config.json` (macOS/Linux) or `%APPDATA%\frefresh\config.json` (Windows).

Each customer needs:
- **Workspace pattern** — Power BI workspace name with `{env}` placeholder (e.g. `DP - {env} - SemMod`)
- **Environments** — list of environments (e.g. `DEV, TEST, PROD`)

## How it works

1. Authenticates via browser-based OAuth2 with Microsoft Entra ID
2. Resolves the workspace by name using the [Power BI REST API](https://learn.microsoft.com/en-us/rest/api/power-bi/groups/get-groups)
3. Lists semantic models (datasets) in the workspace via the [Power BI Datasets API](https://learn.microsoft.com/en-us/rest/api/power-bi/datasets/get-datasets-in-group)
4. Fetches the model's TMDL definition using the [Fabric getDefinition API](https://learn.microsoft.com/en-us/rest/api/fabric/semanticmodel/items/get-semantic-model-definition) to discover all tables
5. Filters tables by partition type — only tables with `partition = m` (Power Query / import) are refreshable
6. Triggers an enhanced refresh via the [Power BI Enhanced Refresh API](https://learn.microsoft.com/en-us/power-bi/connect-data/asynchronous-refresh)
7. Polls until completion and displays the result

### APIs used

| API | Purpose |
|-----|---------|
| **Power BI REST API** (`api.powerbi.com`) | Workspace resolution, dataset listing, refresh trigger/polling |
| **Fabric Items API** (`api.fabric.microsoft.com`) | `getDefinition` — fetches TMDL metadata for table discovery |
| **Microsoft Entra ID** | OAuth2 authentication with token caching |

### Table filtering

The Fabric `getDefinition` API returns the full TMDL definition of a semantic model. Each table's `.tmdl` file contains a partition declaration that indicates how the table gets its data:

- `partition 'X' = m` — M (Power Query) partition, **refreshable**
- `partition 'X' = calculated` — DAX calculated table, **excluded**
- No partition — measure-only table, **excluded**

This approach is more reliable and faster than querying the model with DAX, as it reads metadata only without scanning actual data.

## Authentication

Uses OAuth2 Authorization Code Flow with the Azure CLI public client ID. On first use per customer, a browser window opens for Microsoft login. Tokens are cached in your OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service) and silently refreshed — you typically authenticate once and stay logged in for months.

Use `frefresh logout` to clear all cached credentials.

## License

MIT
