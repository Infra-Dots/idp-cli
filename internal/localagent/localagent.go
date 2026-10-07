// Package localagent runs the InfraDots agent on the developer's machine: the public idp-agent image under
// Docker (or an idp-agent already installed, with --native), on their working tree, with their own model key.
// Nothing here talks to InfraDots.
package localagent

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// AgentVersion is the idp-agent release this CLI drives: its commands and flags are what Args builds.
const AgentVersion = "0.7.0"

// DefaultImage is the public, multi-arch (amd64/arm64) agent image.
const DefaultImage = "public.ecr.aws/e5i7i1j1/idp-agent:" + AgentVersion

// ImageEnv overrides the image (a mirror, or a newer release).
const ImageEnv = "INFRADOTS_AGENT_IMAGE"

// PassthroughEnv are the variables the agent reads for its model account, passed by name only (`-e NAME`):
// Docker copies the value from this environment, so no secret appears in the process list.
var PassthroughEnv = []string{
	"ANTHROPIC_API_KEY",
	"MODEL_PROVIDER", "MODEL_MAP", "MODEL_OVERRIDE",
	"AWS_REGION", "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN", "AWS_BEARER_TOKEN_BEDROCK",
	"CLOUD_ML_REGION", "ANTHROPIC_VERTEX_PROJECT_ID",
	"ANTHROPIC_FOUNDRY_API_KEY", "ANTHROPIC_FOUNDRY_RESOURCE",
	"TERRAFORM_RELEASES_URL", "OPENTOFU_RELEASES_URL",
	"LOG_LEVEL",
}

const (
	repoMount  = "/repo"
	inputMount = "/input"
)

// ReviewOptions are `idp agent review --offline`'s options, with host paths.
type ReviewOptions struct {
	Path         string // the repository, or a workspace folder in it
	Base         string
	Plan         string // a plan file
	Guidance     string // a team guidance file
	Format       string // "md" or "json"
	Model        string
	Instructions string
	IaC          string // "terraform" or "tofu"
}

// ImplementOptions are `idp agent implement --offline`'s options, with host paths.
type ImplementOptions struct {
	Request string // what to change
	Path    string // the repository, or a workspace folder in it
	Format  string // "md" or "json"
	Model   string
	IaC     string // "terraform" or "tofu"
	Patch   bool   // include the full diff in the report
}

// Launcher builds and runs the agent's command line.
type Launcher struct {
	Native    bool   // run an installed idp-agent instead of the image
	Image     string // the image (default: DefaultImage, or $INFRADOTS_AGENT_IMAGE)
	Getenv    func(string) string
	GitConfig func(dir, key string) string // `git config <key>` in dir; "" when unset
	UID       int                          // the developer's ids, so files and git ownership line up inside the container
	GID       int
}

// NewLauncher is a Launcher for this process: its environment and user.
func NewLauncher(native bool, image string) *Launcher {
	return &Launcher{Native: native, Image: image, Getenv: os.Getenv, GitConfig: gitConfig,
		UID: os.Getuid(), GID: os.Getgid()}
}

func gitConfig(dir, key string) string {
	out, err := exec.Command("git", "-C", dir, "config", "--get", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (l *Launcher) image() string {
	if l.Image != "" {
		return l.Image
	}
	if image := l.Getenv(ImageEnv); image != "" {
		return image
	}
	return DefaultImage
}

// FindRepoRoot is the git repository containing path (the nearest directory with a .git entry).
func FindRepoRoot(path string) (string, error) {
	dir, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", path)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("%s is not inside a git repository", path)
		}
		dir = parent
	}
}

// ReviewArgs is the program and arguments that run an offline review.
func (l *Launcher) ReviewArgs(o ReviewOptions) (string, []string, error) {
	target, err := filepath.Abs(o.Path)
	if err != nil {
		return "", nil, err
	}
	root, err := FindRepoRoot(target)
	if err != nil {
		return "", nil, err
	}
	agentArgs := func(path, plan, guidance string) []string {
		args := []string{"review", "--offline", path}
		for _, flag := range [][2]string{{"--base", o.Base}, {"--plan", plan}, {"--guidance", guidance},
			{"--format", o.Format}, {"--model", o.Model}, {"--instructions", o.Instructions}, {"--iac", o.IaC}} {
			if flag[1] != "" {
				args = append(args, flag[0], flag[1])
			}
		}
		return args
	}

	if l.Native {
		plan, guidance, err := absFiles(o.Plan, o.Guidance)
		if err != nil {
			return "", nil, err
		}
		return "idp-agent", agentArgs(target, plan, guidance), nil
	}

	// The repository is mounted read-only: a review never changes the working tree.
	docker, path, err := l.dockerRun(root, target, ":ro")
	if err != nil {
		return "", nil, err
	}
	inContainer := func(hostPath, name string) (string, error) {
		if hostPath == "" {
			return "", nil
		}
		abs, err := filepath.Abs(hostPath)
		if err != nil {
			return "", err
		}
		if _, err := os.Stat(abs); err != nil {
			return "", fmt.Errorf("%s: %w", hostPath, err)
		}
		mounted := inputMount + "/" + name + filepath.Ext(abs)
		docker = append(docker, "-v", abs+":"+mounted+":ro")
		return mounted, nil
	}
	plan, err := inContainer(o.Plan, "plan")
	if err != nil {
		return "", nil, err
	}
	guidance, err := inContainer(o.Guidance, "guidance")
	if err != nil {
		return "", nil, err
	}
	docker = append(docker, l.image())
	return "docker", append(docker, agentArgs(path, plan, guidance)...), nil
}

// ImplementArgs is the program and arguments that run an offline implementation: the agent commits on a new
// branch of the repository as the developer, so the repository is mounted read-write and their git identity
// is passed in (the container has no ~/.gitconfig).
func (l *Launcher) ImplementArgs(o ImplementOptions) (string, []string, error) {
	if strings.TrimSpace(o.Request) == "" {
		return "", nil, errors.New("say what to change, e.g. idp agent implement --offline \"add versioning to the logs bucket\"")
	}
	target, err := filepath.Abs(o.Path)
	if err != nil {
		return "", nil, err
	}
	root, err := FindRepoRoot(target)
	if err != nil {
		return "", nil, err
	}
	agentArgs := func(path string) []string {
		args := []string{"implement", "--offline", o.Request, path}
		for _, flag := range [][2]string{{"--format", o.Format}, {"--model", o.Model}, {"--iac", o.IaC}} {
			if flag[1] != "" {
				args = append(args, flag[0], flag[1])
			}
		}
		if o.Patch {
			args = append(args, "--patch")
		}
		return args
	}
	if l.Native {
		return "idp-agent", agentArgs(target), nil
	}

	identity, err := l.gitIdentity(root)
	if err != nil {
		return "", nil, err
	}
	docker, path, err := l.dockerRun(root, target, "")
	if err != nil {
		return "", nil, err
	}
	docker = append(docker, identity...)
	docker = append(docker, l.image())
	return "docker", append(docker, agentArgs(path)...), nil
}

// gitIdentity is `-e GIT_AUTHOR_NAME=…` and friends: from the environment when set there, else from git config.
func (l *Launcher) gitIdentity(root string) ([]string, error) {
	var args []string
	for _, v := range []struct{ env, key string }{
		{"GIT_AUTHOR_NAME", "user.name"}, {"GIT_AUTHOR_EMAIL", "user.email"},
		{"GIT_COMMITTER_NAME", "user.name"}, {"GIT_COMMITTER_EMAIL", "user.email"},
	} {
		if l.Getenv(v.env) != "" {
			args = append(args, "-e", v.env)
			continue
		}
		value := l.GitConfig(root, v.key)
		if value == "" {
			return nil, fmt.Errorf("git doesn't know who you are, and the agent commits as you: set `git config %s`", v.key)
		}
		args = append(args, "-e", v.env+"="+value)
	}
	return args, nil
}

// dockerRun is `docker run` up to the image: the developer's user, a writable HOME, the model account's variables
// (by name), and the repository at /repo with mode (":ro" or ""). path is target's path in the container.
func (l *Launcher) dockerRun(root, target, mode string) (docker []string, path string, err error) {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return nil, "", err
	}
	docker = []string{"run", "--rm"}
	if runtime.GOOS != "windows" {
		docker = append(docker, "--user", fmt.Sprintf("%d:%d", l.UID, l.GID))
	}
	docker = append(docker, "-e", "HOME=/tmp", "-e", "TMPDIR=/tmp")
	for _, name := range PassthroughEnv {
		if l.Getenv(name) != "" {
			docker = append(docker, "-e", name)
		}
	}
	docker = append(docker, "-v", root+":"+repoMount+mode)
	path = repoMount
	if rel != "." {
		path = repoMount + "/" + filepath.ToSlash(rel)
	}
	return docker, path, nil
}

func absFiles(paths ...string) (string, string, error) {
	out := make([]string, len(paths))
	for i, p := range paths {
		if p == "" {
			continue
		}
		abs, err := filepath.Abs(p)
		if err != nil {
			return "", "", err
		}
		out[i] = abs
	}
	return out[0], out[1], nil
}

// ErrNotInstalled is returned when the program the launcher needs isn't on PATH.
var ErrNotInstalled = errors.New("not installed")

// Run runs the program with the terminal's stdin/stdout/stderr and returns its exit code: the agent's
// own (0 approved, 1 changes requested, 2 error), which a CI step or pre-commit hook acts on.
func Run(program string, args []string) (int, error) {
	if _, err := exec.LookPath(program); err != nil {
		if program == "docker" {
			return 0, fmt.Errorf("docker: %w -- install Docker (or Colima/OrbStack), or use --native with idp-agent installed", ErrNotInstalled)
		}
		return 0, fmt.Errorf("%s: %w -- drop --native to run it with Docker", program, ErrNotInstalled)
	}
	cmd := exec.Command(program, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	err := cmd.Run()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return exit.ExitCode(), nil
	}
	if err != nil {
		return 0, err
	}
	return 0, nil
}

// Redacted is the command line for display (--dry-run): passed-through variables are names, never values.
func Redacted(program string, args []string) string {
	quoted := make([]string, len(args))
	for i, a := range args {
		if strings.ContainsAny(a, " \t\"'") {
			quoted[i] = fmt.Sprintf("%q", a)
		} else {
			quoted[i] = a
		}
	}
	return program + " " + strings.Join(quoted, " ")
}
