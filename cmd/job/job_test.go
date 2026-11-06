package job

import "testing"

// `idp job output <id> --org o --workspace w` failed with "unknown shorthand
// flag: 'w'" while `idp job get` required the same flag, so the command line
// that worked for one verb broke the next — and the docs showed it with
// --workspace. Every job subcommand takes the flag, whether it needs it or not.
func TestEverySubcommandAcceptsWorkspace(t *testing.T) {
	for _, sub := range NewCmd().Commands() {
		for _, args := range [][]string{{"--workspace", "prod-vpc"}, {"-w", "prod-vpc"}} {
			if err := sub.ParseFlags(args); err != nil {
				t.Errorf("idp job %s %v: %v", sub.Name(), args, err)
			}
		}
	}
}

func TestWorkspaceReachesTheSubcommand(t *testing.T) {
	cmd := NewCmd()
	get, _, err := cmd.Find([]string{"get"})
	if err != nil {
		t.Fatal(err)
	}
	if err := get.ParseFlags([]string{"-w", "prod-vpc"}); err != nil {
		t.Fatal(err)
	}
	if got := get.Flag("workspace").Value.String(); got != "prod-vpc" {
		t.Errorf("workspace = %q, want %q", got, "prod-vpc")
	}
}
