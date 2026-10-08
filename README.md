# idp-cli

A Go CLI for the [InfraDots](https://infradots.com) platform. Manage organizations, workspaces, jobs, variables, VCS connections, and agents from your terminal or CI pipeline.

📚 **Docs:** [CLI reference](https://infradots.com/docs/cli) · [AI in Your IDE](https://infradots.com/docs/ai-skill) — add a skill so Claude Code, Cursor, and other assistants can drive `idp` for you.

## Installation

Download the archive for your platform from [Releases](https://github.com/Infra-Dots/idp-cli/releases):
`idp-cli_darwin_arm64.tar.gz` (Apple silicon), `idp-cli_darwin_x86_64.tar.gz`, `idp-cli_linux_x86_64.tar.gz`,
`idp-cli_linux_arm64.tar.gz` or `idp-cli_windows_x86_64.zip`, and put the `idp` binary on your `PATH`:

```sh
tar -xzf idp-cli_darwin_arm64.tar.gz idp && sudo mv idp /usr/local/bin/
idp version
```

To build from source instead: `git clone https://github.com/Infra-Dots/idp-cli && cd idp-cli && go build -o idp .`

## Quick start

```sh
# Log in — opens your browser, signs you in, and saves a freshly minted token
idp auth login

# List orgs you have access to
idp org list

# Run a plan, wait for it, then read its log
idp job run --org my-org --workspace prod-vpc --type plan
idp job get <job-id> --org my-org --workspace prod-vpc --watch
idp job output <job-id> --org my-org
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
idp agent     list | history <job-id> | review --offline [PATH] | implement --offline REQUEST [PATH]
idp version
```

Run `idp <command> --help` for full flags on any subcommand.

### Notes on a few commands

**`auth token create`** prints the token once and never again — the API returns it
only in the creation response. Use `--quiet` to capture just the secret:

```sh
export INFRADOTS_TOKEN=$(idp auth token create -d ci --quiet)
```

**`workspace create`** requires `--source` (the repository the workspace tracks)
and `--tf-version`; `--branch` defaults to `main`.

**`job run --type`** accepts `plan`, `apply`, `destroy`, or `refresh`.

**`job get --watch`** polls until the job reaches a state it will not leave on its
own, then exits with a code CI can branch on:

| Code | Meaning |
|---|---|
| `0` | Finished successfully (`completed` / `applied`) |
| `1` | Failed, rejected, or cancelled |
| `2` | Waiting for approval — run `idp job approve <job-id>` |
| `3` | `--timeout` elapsed (default 3600s; `0` waits forever) |

**`agent review --offline`** reviews the change on your working tree with the InfraDots review
agent before you push: everything since the branch left `--base` (default: the remote's default
branch), committed or not. It runs on your machine, in the public `idp-agent` Docker image (or an
installed `idp-agent` with `--native`), with **your own model key**, and needs no InfraDots account.
The repository is mounted read-only; nothing leaves your machine except the model calls.

```sh
export ANTHROPIC_API_KEY=...          # or MODEL_PROVIDER=bedrock|vertex|foundry with MODEL_MAP
idp agent review --offline                                  # Markdown report in the terminal
idp agent review --offline envs/prod --plan plan.json        # `terraform show -json tfplan > plan.json`
idp agent review --offline -f json > review.json             # for CI
```

Team guidance in `.infradots/guidance.md` is given to the reviewer automatically. Exit codes:
`0` approved, `1` changes requested, `2` error, so it can gate a CI step or a pre-commit hook.
`--dry-run` prints the `docker run` it would execute.

**`agent implement --offline`** turns a request into a new branch of your repository, the same way:
on your machine, with your own model key, no InfraDots account. The agent works in a git worktree on
a new `idp-agent/<timestamp>-<random>` branch cut from your last commit, so your checkout and any
uncommitted work are never touched. It commits as you (your git `user.name`/`user.email`) and never
pushes; review the branch, then merge, push or delete it.

```sh
idp agent implement --offline "add versioning to the logs bucket"           # in the repo, or a workspace folder
idp agent implement --offline "add a read replica" envs/prod --patch         # include the full diff
```

Exit codes: `0` a branch was made, `1` nothing changed, `2` error.

**`variable set`** updates the variable in place when the key already exists in
the same scope. Use `--env` for environment variables (`AWS_ACCESS_KEY_ID` and
friends) rather than Terraform variables.

**`vcs create`** takes OAuth app credentials (`--client-id` / `--client-secret`),
not a personal access token. The connection is created in `pending` status and
must be authorized in the web app before a workspace can use it.

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
