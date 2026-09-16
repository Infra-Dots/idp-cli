package api

import (
	"fmt"
	"sort"
)

// JobTypes are the values the API accepts for a job's `type`. These mirror
// idp.workers.models.JobTypes — note the wire value for a plan-only run is
// "plan", not the Python constant name PLAN_ONLY.
var JobTypes = []string{"plan", "apply", "destroy", "refresh"}

// ValidJobType reports whether t is an accepted job type.
func ValidJobType(t string) bool {
	for _, v := range JobTypes {
		if v == t {
			return true
		}
	}
	return false
}

// Job represents a workspace job (plan or apply).
type Job struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	CreatedBy string `json:"created_by"`
	Via       string `json:"via"`
	// The API names these `*_date`, not `*_at`.
	CreatedDate string `json:"created_date"`
	UpdatedDate string `json:"updated_date"`
}

// CreateJobInput holds the fields for job creation.
type CreateJobInput struct {
	Type string `json:"type"`
}

// JobStage is one stage (init/plan/apply/…) of a job, with its log output.
type JobStage struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Output   string `json:"output"`
	Finished bool   `json:"finished"`
	Errored  bool   `json:"errored"`
}

func (c *Client) ListJobs(orgName, wsName string) ([]Job, error) {
	path := fmt.Sprintf("/api/organizations/%s/workspaces/%s/jobs/", orgName, wsName)
	return getList[Job](c, path)
}

func (c *Client) CreateJob(orgName, wsName string, in CreateJobInput) (*Job, error) {
	var job Job
	path := fmt.Sprintf("/api/organizations/%s/workspaces/%s/jobs/", orgName, wsName)
	if err := c.Post(path, in, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (c *Client) GetJob(orgName, wsName, jobID string) (*Job, error) {
	var job Job
	path := fmt.Sprintf("/api/organizations/%s/workspaces/%s/jobs/%s/", orgName, wsName, jobID)
	if err := c.Get(path, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (c *Client) ApproveJob(orgName, jobID string) error {
	path := fmt.Sprintf("/api/organizations/%s/jobs/%s/approve/", orgName, jobID)
	return c.Post(path, nil, nil)
}

func (c *Client) CancelJob(orgName, wsName, jobID string) error {
	path := fmt.Sprintf("/api/organizations/%s/workspaces/%s/jobs/%s/cancel/", orgName, wsName, jobID)
	return c.Post(path, nil, nil)
}

func (c *Client) DiscardJob(orgName, jobID string) error {
	path := fmt.Sprintf("/api/organizations/%s/jobs/%s/discard/", orgName, jobID)
	return c.Post(path, nil, nil)
}

// GetJobStage fetches the log output for a single job stage. The endpoint is
// scoped by the caller's organization membership server-side, so it needs only
// the job ID.
func (c *Client) GetJobStage(jobID, stage string) (*JobStage, error) {
	var js JobStage
	path := fmt.Sprintf("/api/workers/jobs/%s/stages/%s/", jobID, stage)
	if err := c.Get(path, &js); err != nil {
		return nil, err
	}
	return &js, nil
}

// ListJobStages fetches every stage of a job, ordered for display. The API
// returns them keyed by stage type rather than as an array.
func (c *Client) ListJobStages(jobID string) ([]JobStage, error) {
	var resp struct {
		Stages map[string]JobStage `json:"stages"`
	}
	path := fmt.Sprintf("/api/workers/jobs/%s/stages/", jobID)
	if err := c.Get(path, &resp); err != nil {
		return nil, err
	}

	stages := make([]JobStage, 0, len(resp.Stages))
	for name, s := range resp.Stages {
		if s.Type == "" {
			s.Type = name
		}
		stages = append(stages, s)
	}
	sort.Slice(stages, func(i, j int) bool {
		return stageOrder(stages[i].Type) < stageOrder(stages[j].Type)
	})
	return stages, nil
}

// stageOrder sorts stages into execution order rather than alphabetically, so
// `idp job output` reads top-to-bottom the way the run happened.
func stageOrder(stage string) int {
	order := []string{"init", "details", "validate", "tflint", "plan", "apply", "debug"}
	for i, s := range order {
		if s == stage {
			return i
		}
	}
	return len(order)
}
