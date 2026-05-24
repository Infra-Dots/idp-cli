# idp-cli — Specification

A Go CLI tool that exposes InfraDots' major API capabilities from the terminal,
targeting platform engineers and automation pipelines.

---

## Tech Stack

| Concern | Library |
|---|---|
| CLI framework | `cobra` |
| Config management | `viper` |
| HTTP client | `net/http` + `encoding/json` (stdlib) |
| Table output | `github.com/olekukonko/tablewriter` |
| Styled output | `github.com/charmbracelet/lipgloss` |
| Release builds | `goreleaser` |

---

## Project Structure

```
idp-cli/
├── main.go
├── go.mod / go.sum
├── goreleaser.yaml
│
├── cmd/
│   ├── root.go                 # global flags: --host, --token, --org, --output, --profile
│   ├── version.go
│   ├── auth/
│   │   ├── auth.go             # `idp auth`
│   │   ├── login.go            # `idp auth login`
│   │   ├── logout.go           # `idp auth logout`
│   │   ├── token_list.go       # `idp auth token list`
│   │   ├── token_create.go     # `idp auth token create`
│   │   └── token_revoke.go     # `idp auth token revoke`
│   ├── org/
│   │   ├── org.go              # `idp org`
│   │   ├── list.go             # `idp org list`
│   │   └── get.go              # `idp org get <name>`
│   ├── workspace/
│   │   ├── workspace.go        # `idp workspace`
│   │   ├── list.go             # `idp workspace list`
│   │   ├── create.go           # `idp workspace create`
│   │   ├── get.go              # `idp workspace get`
│   │   ├── update.go           # `idp workspace update`
│   │   └── delete.go           # `idp workspace delete`
│   ├── job/
│   │   ├── job.go              # `idp job`
│   │   ├── list.go             # `idp job list`
│   │   ├── run.go              # `idp job run`
│   │   ├── get.go              # `idp job get`
│   │   ├── approve.go          # `idp job approve`
│   │   ├── cancel.go           # `idp job cancel`
│   │   ├── discard.go          # `idp job discard`
│   │   └── output.go           # `idp job output`
│   ├── variable/
│   │   ├── variable.go         # `idp variable`
│   │   ├── list.go             # `idp variable list`
│   │   ├── set.go              # `idp variable set`
│   │   └── delete.go           # `idp variable delete`
│   ├── vcs/
│   │   ├── vcs.go              # `idp vcs`
│   │   ├── list.go             # `idp vcs list`
│   │   ├── create.go           # `idp vcs create`
│   │   └── delete.go           # `idp vcs delete`
│   └── agent/
│       ├── agent.go            # `idp agent`
│       ├── list.go             # `idp agent list`
│       └── history.go          # `idp agent history`
│
└── internal/
    ├── api/
    │   ├── client.go           # base HTTP client, auth header, typed error handling
    │   ├── organizations.go
    │   ├── workspaces.go
    │   ├── jobs.go
    │   ├── variables.go
    │   ├── vcs.go
    │   └── agents.go
    ├── config/
    │   └── config.go           # profile loading, ~/.idp/config.yaml
    └── output/
        ├── table.go            # table renderer
        └── json.go             # --output json/yaml
```

---

## Command Surface (v1)

```
idp auth login                              # prompts for host + token, saves to profile
idp auth logout
idp auth token list
idp auth token create --description "CI"
idp auth token revoke <token-id>

idp org list
idp org get <org-name>

idp workspace list [--org <org>]
idp workspace create --org <org> --name <name> --vcs <vcs-id> --repo <repo> [--tf-version 1.9.0]
idp workspace get    --org <org> <workspace>
idp workspace update --org <org> <workspace> [--agents-enabled] [--auto-apply]
idp workspace delete --org <org> <workspace>

idp job list    --org <org> --workspace <ws>
idp job run     --org <org> --workspace <ws> [--type plan|apply]
idp job get     --org <org> --workspace <ws> <job-id>
idp job approve <job-id> --org <org>
idp job cancel  <job-id> --org <org>
idp job discard <job-id> --org <org>
idp job output  <job-id> --org <org> --workspace <ws> [--stage plan|apply|init]

idp variable list   --org <org> [--workspace <ws>]
idp variable set    --org <org> [--workspace <ws>] <key> <value> [--sensitive] [--hcl]
idp variable delete --org <org> [--workspace <ws>] <var-id>

idp vcs list   --org <org>
idp vcs create --org <org> --type github --token <pat> --name <name>
idp vcs delete --org <org> <vcs-id>

idp agent list    --org <org>
idp agent history --org <org> [--job <job-id>]

idp version
```

---

## Configuration

**File:** `~/.idp/config.yaml`

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

**Resolution order** (highest wins):

1. `--token` / `--host` CLI flags
2. `INFRADOTS_TOKEN` / `INFRADOTS_HOST` environment variables
3. Active profile in `~/.idp/config.yaml`

**Profile selection:** `--profile <name>` flag or `INFRADOTS_PROFILE` env var.

---

## API Client Design

`internal/api/client.go` — single `Client` struct used by all commands:

```go
type Client struct {
    Host string
    Token string
    http *http.Client
}

func (c *Client) do(method, path string, body, out any) error
func (c *Client) get(path string, out any) error
func (c *Client) post(path string, body, out any) error
func (c *Client) patch(path string, body, out any) error
func (c *Client) delete(path string) error
```

All API errors map to a typed `APIError{StatusCode, Message}` so commands can
handle 404 vs 403 distinctly and surface clean messages to the user.

---

## Output

| Flag | Behaviour |
|---|---|
| _(default)_ | Human-readable table via `tablewriter` |
| `--output json` | Raw JSON response body |
| `--output yaml` | Marshalled YAML |
| `--quiet` / `-q` | Print only resource ID or name (pipe-friendly) |

Errors always go to `stderr`; data to `stdout`.

---

## Implementation Phases

### Phase 1 — Foundation
- [ ] `go.mod` init, cobra/viper wiring, `root.go` with global flags
- [ ] `internal/config` — profile load/save, env override
- [ ] `internal/api/client.go` — base HTTP client, typed error handling
- [ ] `idp auth login/logout/token` commands
- [ ] `internal/api/organizations.go` + `idp org list/get`
- [ ] `internal/api/workspaces.go` + `idp workspace` CRUD
- [ ] `internal/output` — table + JSON renderers

### Phase 2 — Core Workflow
- [ ] `internal/api/jobs.go` + `idp job` commands (list, run, approve, cancel, output)
- [ ] `internal/api/variables.go` + `idp variable` commands (org + workspace scoped)

### Phase 3 — VCS + Agents
- [ ] `internal/api/vcs.go` + `idp vcs` commands
- [ ] `internal/api/agents.go` + `idp agent list/history`

### Phase 4 — Polish + Distribution
- [ ] `goreleaser.yaml` — cross-compile linux/darwin/windows amd64+arm64
- [ ] Shell completion (`cobra` built-in: `idp completion bash|zsh|fish`)
- [ ] `--watch` flag on `idp job get` to poll until terminal state
- [ ] GitHub Actions CI (lint + test + release on tag)

---

## Key Design Decisions

1. **No code generation from OpenAPI.** The API surface is stable enough to
   hand-write typed structs. Avoids a heavy generator toolchain dependency.

2. **`--org` as a persistent flag on root.** Set once via config or `--org`,
   flows down to all subcommands automatically via viper binding.

3. **`idp job output` reads from job stage API.** The API returns stage output
   as text; the CLI streams/pages it. No WebSocket required for v1.

4. **`idp auth login` is interactive-first.** Prompts for host + token if not
   supplied as flags, then writes to `~/.idp/config.yaml`. Non-interactive:
   `idp auth login --host x --token y --no-prompt` for CI use.

5. **One binary, no sub-processes.** All functionality ships in a single
   statically linked Go binary with no runtime dependencies.
