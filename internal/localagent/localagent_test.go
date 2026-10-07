package localagent

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// repo makes a git repository (a .git directory is all FindRepoRoot needs) with a workspace folder.
func repo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "envs", "prod"), 0o755); err != nil {
		t.Fatal(err)
	}
	// macOS temp dirs sit behind a symlink; compare resolved paths.
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

func launcher(env map[string]string) *Launcher {
	gitConfig := map[string]string{"user.name": "Dev Eloper", "user.email": "dev@example.com"}
	return &Launcher{Getenv: func(k string) string { return env[k] },
		GitConfig: func(_, key string) string { return gitConfig[key] }, UID: 501, GID: 20}
}

func contains(args []string, seq ...string) bool {
	return strings.Contains(" "+strings.Join(args, " ")+" ", " "+strings.Join(seq, " ")+" ")
}

func TestReviewArgs_DockerMountsTheRepoReadOnlyAndPassesCredentialsByName(t *testing.T) {
	root := repo(t)
	l := launcher(map[string]string{"ANTHROPIC_API_KEY": "sk-secret", "MODEL_PROVIDER": "bedrock"})
	program, args, err := l.ReviewArgs(ReviewOptions{Path: root})
	if err != nil {
		t.Fatal(err)
	}
	if program != "docker" || args[0] != "run" || args[1] != "--rm" {
		t.Fatalf("got %s %v", program, args)
	}
	for _, want := range [][]string{
		{"-v", root + ":/repo:ro"},
		{"-e", "ANTHROPIC_API_KEY"},
		{"-e", "MODEL_PROVIDER"},
		{"-e", "HOME=/tmp"},
		{DefaultImage, "review", "--offline", "/repo"},
	} {
		if !contains(args, want...) {
			t.Errorf("missing %v in %v", want, args)
		}
	}
	if runtime.GOOS != "windows" && !contains(args, "--user", "501:20") {
		t.Errorf("no --user in %v", args)
	}
	joined := strings.Join(args, " ")
	if strings.Contains(joined, "sk-secret") {
		t.Fatal("a secret value is on the command line")
	}
	if strings.Contains(joined, "AWS_REGION") {
		t.Error("an unset variable was passed through")
	}
}

func TestReviewArgs_AWorkspaceFolderAndAPlanFile(t *testing.T) {
	root := repo(t)
	plan := filepath.Join(t.TempDir(), "plan.json")
	if err := os.WriteFile(plan, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, args, err := launcher(nil).ReviewArgs(ReviewOptions{
		Path: filepath.Join(root, "envs", "prod"), Plan: plan, Base: "develop", Format: "json"})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range [][]string{
		{"-v", root + ":/repo:ro"},
		{"-v", plan + ":/input/plan.json:ro"},
		{"review", "--offline", "/repo/envs/prod", "--base", "develop", "--plan", "/input/plan.json", "--format", "json"},
	} {
		if !contains(args, want...) {
			t.Errorf("missing %v in %v", want, args)
		}
	}
}

func TestReviewArgs_TheImageCanBeOverridden(t *testing.T) {
	root := repo(t)
	_, args, _ := launcher(map[string]string{ImageEnv: "mirror.example/idp-agent:0.6.1"}).ReviewArgs(ReviewOptions{Path: root})
	if !contains(args, "mirror.example/idp-agent:0.6.1", "review") {
		t.Errorf("env image not used: %v", args)
	}
	l := launcher(map[string]string{ImageEnv: "ignored"})
	l.Image = "flag/image:1"
	_, args, _ = l.ReviewArgs(ReviewOptions{Path: root})
	if !contains(args, "flag/image:1", "review") {
		t.Errorf("--image not used: %v", args)
	}
}

func TestReviewArgs_NativeRunsTheInstalledAgentOnHostPaths(t *testing.T) {
	root := repo(t)
	l := launcher(nil)
	l.Native = true
	program, args, err := l.ReviewArgs(ReviewOptions{Path: root, Model: "anthropic:claude-sonnet-5"})
	if err != nil {
		t.Fatal(err)
	}
	if program != "idp-agent" {
		t.Fatalf("program %s", program)
	}
	if !contains(args, "review", "--offline", root, "--model", "anthropic:claude-sonnet-5") {
		t.Errorf("args %v", args)
	}
}

func TestReviewArgs_OutsideARepositoryAndAMissingPlanAreErrors(t *testing.T) {
	if _, _, err := launcher(nil).ReviewArgs(ReviewOptions{Path: t.TempDir()}); err == nil ||
		!strings.Contains(err.Error(), "not inside a git repository") {
		t.Errorf("got %v", err)
	}
	if _, _, err := launcher(nil).ReviewArgs(ReviewOptions{Path: repo(t), Plan: "/nope/plan.json"}); err == nil {
		t.Error("a missing plan file was accepted")
	}
}

func TestRun_ReturnsTheAgentsExitCode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses sh")
	}
	for want, script := range map[int]string{0: "exit 0", 1: "exit 1", 2: "exit 2"} {
		code, err := Run("sh", []string{"-c", script})
		if err != nil || code != want {
			t.Errorf("%q: code %d err %v", script, code, err)
		}
	}
	if _, err := Run("definitely-not-installed-xyz", nil); err == nil || !strings.Contains(err.Error(), "--native") {
		t.Errorf("got %v", err)
	}
}

func TestImplementArgs_DockerMountsTheRepoWritableAndCommitsAsTheDeveloper(t *testing.T) {
	root := repo(t)
	l := launcher(map[string]string{"ANTHROPIC_API_KEY": "sk-secret"})
	program, args, err := l.ImplementArgs(ImplementOptions{
		Request: "add versioning to the logs bucket", Path: filepath.Join(root, "envs", "prod"), Patch: true})
	if err != nil {
		t.Fatal(err)
	}
	if program != "docker" {
		t.Fatalf("program %s", program)
	}
	for _, want := range [][]string{
		{"-v", root + ":/repo"},
		{"-e", "ANTHROPIC_API_KEY"},
		{"-e", "GIT_AUTHOR_NAME=Dev Eloper"},
		{"-e", "GIT_COMMITTER_EMAIL=dev@example.com"},
		{DefaultImage, "implement", "--offline", "add versioning to the logs bucket", "/repo/envs/prod", "--patch"},
	} {
		if !contains(args, want...) {
			t.Errorf("missing %v in %v", want, args)
		}
	}
	joined := strings.Join(args, " ")
	if strings.Contains(joined, ":ro") {
		t.Error("the repository is read-only, but the agent commits to it")
	}
	if strings.Contains(joined, "sk-secret") {
		t.Fatal("a secret value is on the command line")
	}
}

func TestImplementArgs_AGitIdentityInTheEnvironmentIsPassedByName(t *testing.T) {
	root := repo(t)
	l := launcher(map[string]string{"GIT_AUTHOR_NAME": "Bot", "GIT_COMMITTER_NAME": "Bot"})
	_, args, err := l.ImplementArgs(ImplementOptions{Request: "x", Path: root})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(args, "-e", "GIT_AUTHOR_NAME") || !contains(args, "-e", "GIT_AUTHOR_EMAIL=dev@example.com") {
		t.Errorf("args %v", args)
	}
}

func TestImplementArgs_NeedsARequestAndAGitIdentity(t *testing.T) {
	root := repo(t)
	if _, _, err := launcher(nil).ImplementArgs(ImplementOptions{Request: " ", Path: root}); err == nil ||
		!strings.Contains(err.Error(), "say what to change") {
		t.Errorf("got %v", err)
	}
	l := launcher(nil)
	l.GitConfig = func(_, _ string) string { return "" }
	if _, _, err := l.ImplementArgs(ImplementOptions{Request: "x", Path: root}); err == nil ||
		!strings.Contains(err.Error(), "git config user.name") {
		t.Errorf("got %v", err)
	}
}

func TestImplementArgs_NativeRunsTheInstalledAgent(t *testing.T) {
	root := repo(t)
	l := launcher(nil)
	l.Native = true
	program, args, err := l.ImplementArgs(ImplementOptions{Request: "add a replica", Path: root, Format: "json"})
	if err != nil {
		t.Fatal(err)
	}
	if program != "idp-agent" || !contains(args, "implement", "--offline", "add a replica", root, "--format", "json") {
		t.Errorf("got %s %v", program, args)
	}
}
