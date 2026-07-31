# UPass

A secure, local-first password manager for the command line.

## Why UPass?

- 🔑 **Recovery Key** — Forgot your master password? No problem. A unique feature among open-source CLI managers.
- 🛡️ **Built-in Health Check** — Audit weak, duplicate, and breached (HIBP) passwords with k-anonymity.
- 💾 **Auto-backups** — Every save automatically creates a versioned backup. No Git required.
- ⚡ **Single Binary** — No GPG, no Node.js, no background daemon, no cloud dependency.
- 👁️ **Privacy-first** — Everything stays encrypted on your local machine. Secrets are actively wiped from RAM after use.
- 📋 **Clipboard Support** — Passwords are copied to clipboard, never displayed in plain text by default.
- 🏷️ **Multiple Accounts** — Use tags for the same service: `github:work`, `github:personal`.
- 🐚 **Smart Auto-completion** — Native support for Bash, Zsh, and Fish with automatic PATH and config setup.
- 🎲 **Interactive Generation** — Refine generated passwords on the fly by dynamically excluding unwanted characters.
- 🗄️ **Multi-Vault Support** — Easily manage separate contexts (e.g., `default`, `work`, `personal`) with instant switching.
- 📥 **Import/Export** - Supports exporting and importing your vault in both JSON (native format) and CSV (Bitwarden-compatible format).
- 🧠 **Custom Strength Engine** — Evaluates password strength locally using a custom, zero-allocation engine that operates strictly on byte slices, preventing secret leakage into the Go garbage collector.

## Install

```bash
# 1. Clone and build
git clone https://github.com/SaDMikaSa/UPass.git
cd UPass
go build -o upass

# 2 Auto-configure shell completion and PATH
# (Supports Bash, Zsh, and Fish. Run this from the shell you want to configure)
./upass completion install

# 3. (Optional but recommended) Move to your local bin
cp upass ~/.local/bin/
```

## Quick Start

```bash
# Initialize your vault
upass init
# Enter master password
# Scan the QR code or save the recovery key shown on screen

# Add a password interactively
upass add
# Input service: github
# Input login: user@email.com
# Input password: ********

# Add with auto-generated password (press Enter to keep, or type chars to exclude)
upass add reddit -g
# Input login: user123
# Generated password: xK9#mP2$vL5nQ8@rT4
# Enter forbidden characters (e.g., O0l1, or "O 0" to forbid spaces), or Enter to keep:

# Multiple accounts for same service
upass add github:work
upass add github:personal

# Get a password (copies to clipboard)
upass get github

# Show password in terminal (use carefully)
upass get github -s

# List all services
upass list

# Search services, logins, and notes (supports Unicode/Cyrillic for services)
upass search git
upass search user@email.com --login
upass search "work account" --n

# Edit a record (leave fields empty to keep current values, supports renaming)
upass edit github

# Delete a record
upass delete reddit

# Change master password
upass passwd

# Check password health
upass health
# Skip online breach check (fully offline mode)
upass health --no-hibp

# Manage backups
upass backup list
upass backup create
upass backup restore 2

# Generate a password
upass generate
upass generate -l 32 --no-symbols

# Import/Export
upass export -f backup.json
upass export -f backup.csv
upass import -f backup.json
upass import -f backup.csv

# Switch to a new vault context (creates config entry automatically)
upass --vault work init

# Switch vault
upass --vault default
upass --vault work

# You can also use absolute paths for one-off operations without changing the active context
upass --vault /tmp/emergency.json list


```

## Recovery

When you create a vault, UPass generates a random 256-bit recovery key. This key encrypts your master password and is stored in the vault header (`EncryptedMasterPass`). The recovery key is shown as a base64 string — **this is the only time it will be displayed.**

### Storage Recommendations

- Write the recovery key on paper, store in a safe
- Do NOT store it in:
  - Plain text files on your computer
  - Cloud notes (Google Keep, Notion, etc.)
  - Password managers (circular dependency!)
  - Email or messaging apps

## Health

How it works:

1. Your password is hashed with SHA-1 locally
2. Only the first 5 characters of the hash are sent over the network
3. The server returns all known hash suffixes with that prefix
4. UPass checks if your hash suffix appears in the response

The server cannot determine your password from the prefix alone (1M+ possible prefixes, ~500 suffixes per prefix).
Use `upass health --no-hibp` to skip the online check.

## Import/Export

Exporting your vault creates a plaintext JSON file.

- Use case: Migrating to another password manager or creating an offline, air-gapped backup.
- Security Warning: Anyone with access to this file can read all your passwords.
- Best Practice: Export only when necessary, and securely delete the file immediately afterward using tools like shred (Linux/macOS) or sdelete (Windows).
