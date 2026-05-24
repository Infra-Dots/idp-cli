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
# Log in (interactive — prompts for host + token)
idp auth login

# List orgs you have access to
idp org list

# Run a plan and stream its output
idp job run --org my-org --workspace prod-vpc --type plan
idp job output <job-id> --org my-org --workspace prod-vpc --stage plan
```

For non-interactive use (CI):

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
    token: <jwt>
    default_org: my-org
  local:
    host: http://localhost:8000
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
