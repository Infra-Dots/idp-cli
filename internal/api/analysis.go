package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// Analysis statuses. The first three are in progress; the rest are terminal.
const (
	AnalysisQueued    = "queued"
	AnalysisScanning  = "scanning"
	AnalysisAnalyzing = "analyzing"
	AnalysisProposed  = "proposed"
)

// AnalysisInProgress reports whether the analysis is still being worked on.
func AnalysisInProgress(status string) bool {
	switch status {
	case AnalysisQueued, AnalysisScanning, AnalysisAnalyzing:
		return true
	}
	return false
}

// AnalysisRepository is one repository in a repository analysis.
type AnalysisRepository struct {
	Label     string `json:"label"`
	URL       string `json:"url"`
	VCS       string `json:"vcs"`
	Ref       string `json:"ref"`
	CommitSHA string `json:"commit_sha"`
	Error     string `json:"error"`
}

// AnalysisNeed is a need as the server understood it: an alias resolved to canonical kinds.
type AnalysisNeed struct {
	Alias string   `json:"alias"`
	Kinds []string `json:"kinds"`
	Pin   []string `json:"pin"`
}

// AnalysisMention is a need found in the free-text notes.
type AnalysisMention struct {
	Alias string   `json:"alias"`
	Kinds []string `json:"kinds"`
	Pin   []string `json:"pin"`
	Start int      `json:"start"`
	End   int      `json:"end"`
	Text  string   `json:"text"`
}

// AnalysisSizing is the version/tier a provisioned resource was sized to.
type AnalysisSizing struct {
	Version          string `json:"version"`
	Tier             string `json:"tier"`
	HighAvailability bool   `json:"high_availability"`
}

// AnalysisDecision is what the agent decided for one infrastructure need.
type AnalysisDecision struct {
	Kind            string          `json:"kind"`
	Name            string          `json:"name"`
	Repositories    []string        `json:"repositories"`
	Evidence        json.RawMessage `json:"evidence"`
	Decision        string          `json:"decision"` // provision | reuse | unresolved
	TargetWorkspace string          `json:"target_workspace"`
	PendingCreation bool            `json:"pending_creation"`
	ResourceType    string          `json:"resource_type"`
	Sizing          *AnalysisSizing `json:"sizing"`
	Rationale       string          `json:"rationale"`
	Unverified      bool            `json:"unverified"`
	Rank            *int            `json:"rank"`
	BelowTopRank    bool            `json:"below_top_rank"`
}

// AnalysisTarget is a workspace the proposal would send an implementation request to.
type AnalysisTarget struct {
	WorkspaceName   string   `json:"workspace_name"`
	UserPrompt      string   `json:"user_prompt"`
	Rationale       string   `json:"rationale"`
	Selected        bool     `json:"selected"`
	PendingCreation bool     `json:"pending_creation"`
	Kinds           []string `json:"kinds"`
}

// AnalysisProposal is the agent's proposal, awaiting confirmation.
type AnalysisProposal struct {
	Summary   string             `json:"summary"`
	Targets   []AnalysisTarget   `json:"targets"`
	Decisions []AnalysisDecision `json:"decisions"`
	Questions []string           `json:"questions"`
}

// Analysis is a repository analysis. Raw keeps the response as the API sent it, for `--output json`.
type Analysis struct {
	ID             string               `json:"id"`
	Status         string               `json:"status"`
	Environment    string               `json:"environment"`
	TagHints       map[string]string    `json:"tag_hints"`
	Notes          string               `json:"notes"`
	Needs          []AnalysisNeed       `json:"needs"`
	Mentions       []AnalysisMention    `json:"mentions"`
	Source         string               `json:"source"`
	RequestedBy    string               `json:"requested_by"`
	Repositories   []AnalysisRepository `json:"repositories"`
	Proposal       *AnalysisProposal    `json:"proposal"`
	RequestGroupID string               `json:"request_group_id"`
	AgentRunID     json.RawMessage      `json:"agent_run_id"`
	FailureReason  string               `json:"failure_reason"`
	CreatedAt      string               `json:"created_at"`
	UpdatedAt      string               `json:"updated_at"`
	ConfirmedAt    string               `json:"confirmed_at"`
	Raw            json.RawMessage      `json:"-"`
}

// AnalysisSummary is a row of the analysis list.
type AnalysisSummary struct {
	ID            string   `json:"id"`
	Status        string   `json:"status"`
	Environment   string   `json:"environment"`
	Source        string   `json:"source"`
	RequestedBy   string   `json:"requested_by"`
	Repositories  []string `json:"repositories"`
	FailureReason string   `json:"failure_reason"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// AnalysisRepoInput names a repository to analyze.
type AnalysisRepoInput struct {
	URL string `json:"url"`
	Ref string `json:"ref,omitempty"`
	VCS string `json:"vcs,omitempty"`
}

// AnalysisInput starts an analysis.
type AnalysisInput struct {
	Repositories []AnalysisRepoInput `json:"repositories"`
	Environment  string              `json:"environment,omitempty"`
	Tags         map[string]string   `json:"tags,omitempty"`
	Needs        []string            `json:"needs,omitempty"`
	Notes        string              `json:"notes,omitempty"`
	Source       string              `json:"source"`
}

// ConfirmTarget ticks or unticks one of the proposal's targets. Without Selected it keeps the proposal's choice.
type ConfirmTarget struct {
	// Index is the proposal target this one stands for. Two targets can share a workspace (an "other"
	// dependency is its own target), so the server can't tell them apart by name.
	Index         *int   `json:"index,omitempty"`
	WorkspaceName string `json:"workspace_name"`
	UserPrompt    string `json:"user_prompt,omitempty"`
	Selected      *bool  `json:"selected,omitempty"`
}

// ConfirmInput confirms an analysis. No Targets means "dispatch the proposal's selected targets".
type ConfirmInput struct {
	Targets []ConfirmTarget `json:"targets,omitempty"`
}

// ConfirmedTarget is the outcome for one workspace of a confirm.
type ConfirmedTarget struct {
	WorkspaceName  string `json:"workspace_name"`
	Status         string `json:"status"` // dispatched | blocked
	AgentHistoryID *int   `json:"agent_history_id"`
	Reason         string `json:"reason"`
}

// ConfirmResult is the response to a confirm.
type ConfirmResult struct {
	RequestGroupID string            `json:"request_group_id"`
	Targets        []ConfirmedTarget `json:"targets"`
	Raw            json.RawMessage   `json:"-"`
}

func analysesPath(orgName string) string {
	return fmt.Sprintf("/api/agents/%s/repository-analyses/", orgName)
}

// analysisError turns the API's rejection into one that says what to do: a 404 here means the feature is
// not switched on for the user, and the extras (`environments`, `known_kinds`, `connections`) list the
// valid choices.
func analysisError(err error) error {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return err
	}
	if apiErr.StatusCode == http.StatusNotFound {
		return fmt.Errorf("repository analysis is not available: either it is not enabled for your user yet " +
			"(it is staff-only for now), or the analysis does not exist")
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s", apiErr.Message)
	keys := make([]string, 0, len(apiErr.Extras))
	for k := range apiErr.Extras {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		switch v := apiErr.Extras[k].(type) {
		case []any:
			items := make([]string, len(v))
			for i, item := range v {
				items[i] = fmt.Sprint(item)
			}
			fmt.Fprintf(&b, "\n  %s: %s", k, strings.Join(items, ", "))
		default:
			fmt.Fprintf(&b, "\n  %s: %v", k, v)
		}
	}
	if apiErr.StatusCode == http.StatusUnprocessableEntity {
		b.WriteString("\n  (the repository is not visible to the VCS connection; check the URL, or pass --vcs)")
	}
	return fmt.Errorf("%s", b.String())
}

func decodeAnalysis(raw json.RawMessage) (*Analysis, error) {
	var a Analysis
	if err := json.Unmarshal(raw, &a); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	a.Raw = raw
	return &a, nil
}

// CreateAnalysis starts a repository analysis (202: it runs in the background).
func (c *Client) CreateAnalysis(orgName string, in AnalysisInput) (*Analysis, error) {
	var raw json.RawMessage
	if err := c.Post(analysesPath(orgName), in, &raw); err != nil {
		return nil, analysisError(err)
	}
	return decodeAnalysis(raw)
}

func (c *Client) GetAnalysis(orgName, id string) (*Analysis, error) {
	var raw json.RawMessage
	if err := c.Get(analysesPath(orgName)+id+"/", &raw); err != nil {
		return nil, analysisError(err)
	}
	return decodeAnalysis(raw)
}

// ListAnalyses returns the summaries and the raw rows (for `--output json`).
func (c *Client) ListAnalyses(orgName string) ([]AnalysisSummary, []json.RawMessage, error) {
	raws, err := getList[json.RawMessage](c, analysesPath(orgName))
	if err != nil {
		return nil, nil, analysisError(err)
	}
	out := make([]AnalysisSummary, len(raws))
	for i, raw := range raws {
		if err := json.Unmarshal(raw, &out[i]); err != nil {
			return nil, nil, fmt.Errorf("decoding response: %w", err)
		}
	}
	return out, raws, nil
}

func (c *Client) ConfirmAnalysis(orgName, id string, in ConfirmInput) (*ConfirmResult, error) {
	var raw json.RawMessage
	if err := c.Post(analysesPath(orgName)+id+"/confirm/", in, &raw); err != nil {
		return nil, analysisError(err)
	}
	var res ConfirmResult
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	res.Raw = raw
	return &res, nil
}

func (c *Client) RejectAnalysis(orgName, id string) (*Analysis, error) {
	var raw json.RawMessage
	if err := c.Post(analysesPath(orgName)+id+"/reject/", nil, &raw); err != nil {
		return nil, analysisError(err)
	}
	return decodeAnalysis(raw)
}
