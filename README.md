# splitwise

A fast, clean command-line interface for [Splitwise](https://www.splitwise.com). Manage groups, expenses, balances, and settlements from your terminal.

## Install

### Homebrew (macOS/Linux)

```bash
brew tap barronlroth/tap
brew install splitwise
```

### Go

```bash
go install github.com/barronlroth/splitwise-cli@latest
```

### From Source

```bash
git clone https://github.com/barronlroth/splitwise-cli.git
cd splitwise-cli
go build -o splitwise .
```

## Setup

### 1. Register a Splitwise App

1. Go to [secure.splitwise.com/apps](https://secure.splitwise.com/apps) and log in
2. Click **"Register your application"**
3. Fill in:
   - **Application name:** anything (e.g. "Splitwise CLI")
   - **Description:** anything
   - **Homepage URL:** your repo URL or any valid URL
   - **Callback URL:** `http://localhost:8484/callback` (must match exactly — see note below)
4. Accept the Terms of Service and click **"Register and get API key"**
5. Copy your **Consumer Key** (Client ID) and **Consumer Secret**

> **⚠️ Callback URL gotcha:** Splitwise ignores the `redirect_uri` in the OAuth request and always uses the registered callback URL. The CLI starts a local server on a random port, but the redirect goes to whatever you registered. If you hit `ERR_CONNECTION_REFUSED` after authorizing, the port doesn't match. Workaround: do the token exchange manually with `curl` using the `code` from the redirect URL, or register the callback as `http://localhost` without a port.

### 2. Authenticate

```bash
splitwise auth
```

You'll be prompted for your Client ID and Client Secret. The CLI opens your browser for OAuth consent, then stores the token locally at `~/.config/splitwise-cli/auth.json`.

**Manual token exchange** (if the callback redirect fails):
```bash
# After authorizing in the browser, grab the `code` parameter from the redirect URL
curl -s -X POST https://secure.splitwise.com/oauth/token \
  -d "grant_type=authorization_code" \
  -d "code=YOUR_CODE_HERE" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "redirect_uri=http://localhost:8484/callback"

# Save the response to ~/.config/splitwise-cli/auth.json:
# {
#   "client_id": "...",
#   "client_secret": "...",
#   "token": { "access_token": "...", "token_type": "bearer", "expiry": "2099-01-01T00:00:00Z" }
# }
```

## Usage

### Current User

```bash
splitwise me
```

### Groups

```bash
# List all groups
splitwise groups

# Show group details + balances
splitwise group "Barron & Nina"
splitwise group 12345
```

### Expenses

```bash
# List recent expenses
splitwise expenses list
splitwise expenses list --group "Barron & Nina" --limit 10

# Filter by date
splitwise expenses list --after 2025-01-01 --before 2025-06-01

# Create an expense (split evenly in group)
splitwise expenses create "Dinner" 85.50 --group "Barron & Nina"

# Create with specific currency
splitwise expenses create "Groceries" 42.00 --group "Home" --currency EUR

# Delete an expense
splitwise expenses delete 123456
```

### Balances

```bash
# Show all friend balances
splitwise balances

# Show balances within a group
splitwise balances --group "Barron & Nina"
```

### Friends

```bash
splitwise friends
```

### Settle Up

```bash
# Record a settlement with a friend
splitwise settle "Nina"
splitwise settle 12345 --group "Barron & Nina"
```

### Configuration

```bash
# Set defaults so you don't have to pass --group every time
splitwise config set default_group "Barron & Nina"
splitwise config set default_currency USD

# View config
splitwise config show
```

## Output Formats

Every command supports:

| Flag | Description |
|------|-------------|
| `--json` | Raw JSON output |
| `--quiet` | Minimal output (IDs/amounts only) for scripting |
| `--no-color` | Disable color output (also respects `NO_COLOR` env var) |

```bash
# Pipe expense IDs to another command
splitwise expenses list --quiet | head -5

# Get JSON for scripting
splitwise balances --json | jq '.[] | select(.amount != "0.00")'
```

## Logout

```bash
splitwise auth logout
```

## Configuration Files

| File | Purpose |
|------|---------|
| `~/.config/splitwise-cli/auth.json` | OAuth token (keep this secret) |
| `~/.config/splitwise-cli/config.json` | User preferences |

## Development

```bash
# Build
go build -o splitwise .

# Run
./splitwise me

# Release (requires goreleaser)
git tag v1.0.0
git push --tags
# GitHub Actions handles the rest
```

## Release Process

1. Tag a version: `git tag v1.0.0 && git push --tags`
2. GitHub Actions runs GoReleaser automatically
3. Cross-platform binaries are built and attached to the GitHub release
4. Homebrew formula is updated in [barronlroth/homebrew-tap](https://github.com/barronlroth/homebrew-tap)

### Required Secrets

Set these in GitHub repo settings → Secrets:

- `HOMEBREW_TAP_GITHUB_TOKEN` — a PAT with `repo` scope for the homebrew-tap repo

## License

MIT — see [LICENSE](LICENSE).
