# idp-cli

A Go CLI for the [InfraDots](https://infradots.com) platform. Manage organizations, workspaces, jobs, variables, VCS connections, and agents from your terminal or CI pipeline.

## Installation

### Homebrew (macOS / Linux)

```sh
brew install infradots/tap/idp
```

### From release

Download the archive for your platform from [Releases](https://github.com/infradots/idp-cli/releases) and place the `idp` binary on your `PATH`.

### From source

```sh
go install github.com/infradots/idp-cli@latest
```

## Quick start

```sh
# Log in — opens your browser, signs you in, and saves a freshly minted token
idp auth login

# List orgs you have access to
idp org list

# Run a plan and stream its output
idp job run --org my-org --workspace prod-vpc --type plan
idp job output <job-id> --org my-org --workspace prod-vpc --stage plan
```

`idp auth login` starts a local callback server on `127.0.0.1`, opens the
InfraDots web app to authenticate you, and stores the issued API token in your
profile. For a self-hosted or local install, point it at the right web app:

```sh
idp auth login --host http://localhost:8000 --app-url http://localhost:3001
```

For non-interactive use (CI), skip the browser and pass a token created in the
web app under Settings → Tokens:

```sh
idp auth login --host https://api.infradots.com --token "$INFRADOTS_TOKEN" --no-prompt
```

## Configuration

Config lives at `~/.idp/config.yaml` and supports multiple profiles:

```yaml
default_profile: prod

profiles:
  prod:
    host: https://api.infradots.com
    web_url: https://app.infradots.com   # used by `idp auth login` browser flow
    token: <jwt>
    default_org: my-org
  local:
    host: http://localhost:8000
    web_url: http://localhost:3001
    token: <jwt>
    default_org: dev-org
```

Resolution order (highest wins):

1. `--token` / `--host` flags
2. `INFRADOTS_TOKEN` / `INFRADOTS_HOST` env vars
3. Active profile in `~/.idp/config.yaml`

Switch profile with `--profile <name>` or `INFRADOTS_PROFILE`.

## Commands

```
idp auth      login | logout | token list|create|revoke
idp org       list | get
idp workspace list | create | get | update | delete
idp job       list | run | get | approve | cancel | discard | output
idp variable  list | set | delete
idp vcs       list | create | delete
idp agent     list | history
idp version
```

Run `idp <command> --help` for full flags on any subcommand.

## Output formats

| Flag | Behavior |
|---|---|
| _(default)_ | Human-readable table |
| `--output json` | Raw JSON |
| `--output yaml` | YAML |
| `-q`, `--quiet` | Only resource ID/name (pipe-friendly) |

Data goes to `stdout`; errors to `stderr`.

## Shell completion

```sh
idp completion bash   # or zsh, fish, powershell
```

## Development

```sh
go build -o idp .
go test ./...
```

Release builds are cut with [goreleaser](https://goreleaser.com/) — see [.goreleaser.yaml](./.goreleaser.yaml).

See [SPEC.md](./SPEC.md) for the full design specification.
