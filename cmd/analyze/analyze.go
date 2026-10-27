package analyze

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/cli"
	"github.com/infradots/idp-cli/internal/output"
)

// NewCmd returns the `idp analyze` command group.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze repositories for the infrastructure they need, and request it",
	}
	cmd.AddCommand(newRepoCmd())
	cmd.AddCommand(newShowCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newConfirmCmd())
	cmd.AddCommand(newRejectCmd())
	return cmd
}

var listHeaders = []string{"ID", "STATUS", "ENV", "SOURCE", "REPOSITORIES", "CREATED"}

func listRow(s api.AnalysisSummary) []string {
	return []string{s.ID, s.Status, s.Environment, s.Source, strings.Join(s.Repositories, ", "), s.CreatedAt}
}

// printRaw prints the API's own JSON, indented, for `--output json`.
func printRaw(p *output.Printer, raw json.RawMessage, fallback any) error {
	if p.Format != output.FormatJSON || len(raw) == 0 {
		return p.Print(fallback, nil, nil)
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return p.JSON(fallback)
	}
	return p.JSON(v)
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func yesNo(b bool, label string) string {
	if b {
		return " [" + label + "]"
	}
	return ""
}

// printSummary prints what the server understood of the request: status, the canonical needs, and the
// needs it picked out of the notes.
func printSummary(w io.Writer, a *api.Analysis) {
	_, _ = fmt.Fprintf(w, "Analysis %s\nStatus:      %s\n", a.ID, a.Status)
	if a.Environment != "" {
		_, _ = fmt.Fprintf(w, "Environment: %s\n", a.Environment)
	}
	if len(a.TagHints) > 0 {
		tags := make([]string, 0, len(a.TagHints))
		for k, v := range a.TagHints {
			tags = append(tags, k+"="+v)
		}
		sort.Strings(tags)
		_, _ = fmt.Fprintf(w, "Tags:        %s\n", strings.Join(tags, ", "))
	}
	if len(a.Needs) > 0 {
		_, _ = fmt.Fprintln(w, "Needs understood:")
		for _, n := range a.Needs {
			_, _ = fmt.Fprintf(w, "  %s -> %s%s\n", n.Alias, strings.Join(n.Kinds, ", "), pinned(n.Pin))
		}
	}
	if len(a.Mentions) > 0 {
		_, _ = fmt.Fprintln(w, "Mentions found in notes:")
		for _, m := range a.Mentions {
			_, _ = fmt.Fprintf(w, "  %q (%s) -> %s%s\n", m.Text, m.Alias, strings.Join(m.Kinds, ", "), pinned(m.Pin))
		}
	}
	if a.FailureReason != "" {
		_, _ = fmt.Fprintf(w, "Failure:     %s\n", a.FailureReason)
	}
}

func pinned(pin []string) string {
	if len(pin) == 0 {
		return ""
	}
	return " (pin: " + strings.Join(pin, ", ") + ")"
}

// printDetail prints everything `show` knows: repositories, the proposal's decisions, questions and targets.
func printDetail(w io.Writer, a *api.Analysis) {
	printSummary(w, a)
	if a.RequestGroupID != "" {
		_, _ = fmt.Fprintf(w, "Request group: %s\n", a.RequestGroupID)
	}
	if len(a.Repositories) > 0 {
		_, _ = fmt.Fprintln(w, "\nRepositories:")
		for _, r := range a.Repositories {
			_, _ = fmt.Fprintf(w, "  %s  %s  commit %s\n", r.Label, r.URL, orDash(r.CommitSHA))
			if r.Error != "" {
				_, _ = fmt.Fprintf(w, "      error: %s\n", r.Error)
			}
		}
	}
	p := a.Proposal
	if p == nil {
		return
	}
	if p.Summary != "" {
		_, _ = fmt.Fprintf(w, "\nSummary: %s\n", p.Summary)
	}
	if len(p.Decisions) > 0 {
		_, _ = fmt.Fprintln(w, "\nDecisions:")
		for _, d := range p.Decisions {
			rank := "-"
			if d.Rank != nil {
				rank = fmt.Sprint(*d.Rank)
			}
			_, _ = fmt.Fprintf(w, "  %s: %s -> %s  resource %s  rank %s%s%s\n", d.Kind, d.Decision,
				orDash(d.TargetWorkspace+yesNo(d.PendingCreation, "to be created")), orDash(d.ResourceType), rank,
				yesNo(d.BelowTopRank, "below top rank"), yesNo(d.Unverified, "unverified"))
			if d.Rationale != "" {
				_, _ = fmt.Fprintf(w, "      %s\n", d.Rationale)
			}
		}
	}
	if len(p.Questions) > 0 {
		_, _ = fmt.Fprintln(w, "\nQuestions:")
		for _, q := range p.Questions {
			_, _ = fmt.Fprintf(w, "  - %s\n", q)
		}
	}
	if len(p.Targets) > 0 {
		_, _ = fmt.Fprintln(w, "\nTargets:")
		printTargets(w, p.Targets)
	}
}

func printTargets(w io.Writer, targets []api.AnalysisTarget) {
	for _, t := range targets {
		mark := "[ ]"
		if t.Selected {
			mark = "[x]"
		}
		_, _ = fmt.Fprintf(w, "  %s %s%s  %s\n", mark, t.WorkspaceName, yesNo(t.PendingCreation, "to be created"),
			strings.Join(t.Kinds, ", "))
		if t.Rationale != "" {
			_, _ = fmt.Fprintf(w, "        %s\n", t.Rationale)
		}
	}
}

// showAnalysis renders an analysis the way `show` does, honouring --output.
func showAnalysis(scope *cli.OrgScope, a *api.Analysis) error {
	if scope.Printer.Format != output.FormatTable {
		return printRaw(scope.Printer, a.Raw, a)
	}
	printDetail(scope.Printer.Out, a)
	return nil
}

// rawsOrItems picks the API's own rows for json, and the typed ones for yaml.
func rawsOrItems(scope *cli.OrgScope, raws []json.RawMessage, items []api.AnalysisSummary) any {
	if scope.Printer.Format == output.FormatJSON {
		return raws
	}
	return items
}
