package analyze

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/cli"
	"github.com/infradots/idp-cli/internal/output"
)

type confirmOptions struct {
	id      string
	yes     bool
	only    []string
	include []string
}

func newConfirmCmd() *cobra.Command {
	var opts confirmOptions

	cmd := &cobra.Command{
		Use:   "confirm ID",
		Short: "Confirm a proposal and request the infrastructure in its workspaces",
		Long: `Confirm a proposal: each selected target gets an implementation request. With no flags the proposal's
own selection is used. --only requests exactly the named workspaces; --include adds workspaces the proposal
left unticked.`,
		Example: `  idp analyze confirm 6f1c0c1e-...
  idp analyze confirm 6f1c0c1e-... --only payments-prod --yes
  idp analyze confirm 6f1c0c1e-... --include payments-cache`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}
			opts.id = args[0]
			return runConfirm(scope, opts, os.Stdin)
		},
	}
	f := cmd.Flags()
	f.BoolVarP(&opts.yes, "yes", "y", false, "Do not ask for confirmation")
	f.StringSliceVar(&opts.only, "only", nil, "Request only these workspaces (comma-separated)")
	f.StringSliceVar(&opts.include, "include", nil, "Also request these workspaces, which the proposal left unselected")
	return cmd
}

// buildTargets applies --only/--include to the proposal's targets. With neither it returns nil, and the
// server dispatches the proposal's own selection.
func buildTargets(p *api.AnalysisProposal, only, include []string) ([]api.ConfirmTarget, error) {
	if len(only) == 0 && len(include) == 0 {
		return nil, nil
	}
	known := map[string]bool{}
	for _, t := range p.Targets {
		known[t.WorkspaceName] = true
	}
	for _, name := range append(append([]string{}, only...), include...) {
		if !known[name] {
			return nil, output.NewError("the proposal has no target workspace %q", name)
		}
	}
	set := func(names []string) map[string]bool {
		m := map[string]bool{}
		for _, n := range names {
			m[n] = true
		}
		return m
	}
	onlySet, includeSet := set(only), set(include)

	targets := make([]api.ConfirmTarget, len(p.Targets))
	for i, t := range p.Targets {
		selected := t.Selected
		if len(only) > 0 {
			selected = onlySet[t.WorkspaceName]
		}
		if includeSet[t.WorkspaceName] {
			selected = true
		}
		index := i
		targets[i] = api.ConfirmTarget{Index: &index, WorkspaceName: t.WorkspaceName, Selected: &selected}
	}
	return targets, nil
}

func runConfirm(scope *cli.OrgScope, opts confirmOptions, in io.Reader) error {
	a, err := scope.Client.GetAnalysis(scope.Org, opts.id)
	if err != nil {
		return err
	}
	if a.Proposal == nil {
		return output.NewError("analysis %s has no proposal to confirm (status: %s)", a.ID, a.Status)
	}
	targets, err := buildTargets(a.Proposal, opts.only, opts.include)
	if err != nil {
		return err
	}

	// What will be requested: the explicit list when flags were given, else the proposal's selection.
	var chosen []string
	if targets != nil {
		for _, t := range targets {
			if *t.Selected {
				chosen = append(chosen, t.WorkspaceName)
			}
		}
	} else {
		for _, t := range a.Proposal.Targets {
			if t.Selected {
				chosen = append(chosen, t.WorkspaceName)
			}
		}
	}
	if len(chosen) == 0 {
		return output.NewError("no target is selected — pass --include <workspace> to pick one")
	}

	if !opts.yes {
		out := scope.Printer.Out
		_, _ = fmt.Fprintf(out, "Request infrastructure in %d workspace(s):\n", len(chosen))
		for _, name := range chosen {
			_, _ = fmt.Fprintf(out, "  - %s\n", name)
		}
		_, _ = fmt.Fprint(out, "Confirm? [y/N] ")
		answer, _ := bufio.NewReader(in).ReadString('\n')
		if a := strings.ToLower(strings.TrimSpace(answer)); a != "y" && a != "yes" {
			return output.NewError("aborted")
		}
	}

	res, err := scope.Client.ConfirmAnalysis(scope.Org, a.ID, api.ConfirmInput{Targets: targets})
	if err != nil {
		return err
	}
	if scope.Printer.Format != output.FormatTable {
		return printRaw(scope.Printer, res.Raw, res)
	}
	if scope.Printer.Quiet {
		scope.Printer.PrintID(res.RequestGroupID)
		return nil
	}
	headers := []string{"WORKSPACE", "STATUS", "REASON"}
	rows := make([][]string, len(res.Targets))
	for i, t := range res.Targets {
		rows[i] = []string{t.WorkspaceName, t.Status, t.Reason}
	}
	scope.Printer.Table(headers, rows)
	_, _ = fmt.Fprintf(scope.Printer.Out, "\nRequest group: %s\n", res.RequestGroupID)
	return nil
}
