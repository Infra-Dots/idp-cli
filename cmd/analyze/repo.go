package analyze

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/cli"
	"github.com/infradots/idp-cli/internal/output"
)

// pollInterval is how often --wait asks for the analysis; a variable so tests needn't sleep.
var pollInterval = 5 * time.Second

// tagKeys are the only tags the API takes.
var tagKeys = map[string]bool{"team": true, "region": true, "app": true}

type repoOptions struct {
	urls        []string
	environment string
	tags        []string
	needs       []string
	notes       string
	vcs         string
	ref         string
	wait        bool
	timeout     time.Duration
}

func newRepoCmd() *cobra.Command {
	var opts repoOptions

	cmd := &cobra.Command{
		Use:   "repo URL...",
		Short: "Analyze repositories and propose the infrastructure they need",
		Long: `Scan one or more repositories, work out the infrastructure they need (databases, buckets, queues…)
and propose which workspaces to request it in. Nothing is provisioned until you run
` + "`idp analyze confirm`" + `.

The analysis runs in the background: this prints its id, plus the needs the server understood from
--needs and the ones it found in --notes. With --wait it polls until the proposal is ready and prints it.`,
		Example: `  idp analyze repo github.com/acme/payments --env prod --tag team=payments --needs postgresql,s3
  idp analyze repo github.com/acme/api github.com/acme/worker --notes "api needs a redis cache" --wait`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}
			opts.urls = args
			return runRepo(scope, opts)
		},
	}
	f := cmd.Flags()
	f.StringVar(&opts.environment, "env", "", "Environment the infrastructure is for, e.g. prod")
	f.StringArrayVar(&opts.tags, "tag", nil, "Tag hint `key=value` (team, region or app); repeatable")
	f.StringSliceVar(&opts.needs, "needs", nil, "Needs, comma-separated, e.g. postgresql,s3")
	f.StringVar(&opts.notes, "notes", "", "Free-text notes; needs mentioned here are picked up too")
	f.StringVar(&opts.vcs, "vcs", "", "VCS connection to read the repositories with")
	f.StringVar(&opts.ref, "ref", "", "Branch to analyze (default: the repository's default branch)")
	f.BoolVar(&opts.wait, "wait", false, "Wait for the proposal and print it")
	f.DurationVar(&opts.timeout, "timeout", 5*time.Minute, "How long --wait waits")
	return cmd
}

func parseTags(tags []string) (map[string]string, error) {
	if len(tags) == 0 {
		return nil, nil
	}
	out := make(map[string]string, len(tags))
	for _, t := range tags {
		k, v, ok := strings.Cut(t, "=")
		if !ok || k == "" || v == "" {
			return nil, output.NewError("--tag %q must be key=value", t)
		}
		if !tagKeys[k] {
			return nil, output.NewError("--tag key %q is not one of team, region, app", k)
		}
		out[k] = v
	}
	return out, nil
}

func runRepo(scope *cli.OrgScope, opts repoOptions) error {
	tags, err := parseTags(opts.tags)
	if err != nil {
		return err
	}
	in := api.AnalysisInput{
		Environment: opts.environment,
		Tags:        tags,
		Needs:       opts.needs,
		Notes:       opts.notes,
		Source:      "cli",
	}
	for _, u := range opts.urls {
		in.Repositories = append(in.Repositories, api.AnalysisRepoInput{URL: u, Ref: opts.ref, VCS: opts.vcs})
	}

	a, err := scope.Client.CreateAnalysis(scope.Org, in)
	if err != nil {
		return err
	}
	if scope.Printer.Quiet && !opts.wait {
		scope.Printer.PrintID(a.ID)
		return nil
	}
	if !opts.wait {
		if scope.Printer.Format != output.FormatTable {
			return printRaw(scope.Printer, a.Raw, a)
		}
		printSummary(scope.Printer.Out, a)
		_, _ = fmt.Fprintf(scope.Printer.Out, "\nFollow it with: idp analyze show %s\n", a.ID)
		return nil
	}

	if scope.Printer.Format == output.FormatTable && !scope.Printer.Quiet {
		printSummary(scope.Printer.Out, a)
	}
	a, err = waitForAnalysis(scope, a, opts.timeout)
	if err != nil {
		return err
	}
	if scope.Printer.Quiet {
		scope.Printer.PrintID(a.ID)
		return nil
	}
	if scope.Printer.Format == output.FormatTable {
		_, _ = fmt.Fprintln(scope.Printer.Out)
	}
	return showAnalysis(scope, a)
}

// waitForAnalysis polls until the analysis leaves queued/scanning/analyzing, or the timeout passes.
func waitForAnalysis(scope *cli.OrgScope, a *api.Analysis, timeout time.Duration) (*api.Analysis, error) {
	deadline := time.Now().Add(timeout)
	last := a.Status
	for api.AnalysisInProgress(a.Status) {
		if time.Now().After(deadline) {
			return nil, output.NewError("timed out after %s with the analysis still %s — follow it with `idp analyze show %s`",
				timeout, a.Status, a.ID)
		}
		time.Sleep(pollInterval)
		var err error
		if a, err = scope.Client.GetAnalysis(scope.Org, a.ID); err != nil {
			return nil, err
		}
		if a.Status != last {
			fmt.Fprintf(os.Stderr, "status: %s\n", a.Status)
			last = a.Status
		}
	}
	return a, nil
}
