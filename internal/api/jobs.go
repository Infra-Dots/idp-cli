package api

import "fmt"

// Job represents a workspace job (plan or apply).
type Job struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	CreatedBy   string `json:"created_by"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// CreateJobInput holds the fields for job creation.
type CreateJobInput struct {
	Type string `json:"type"` // "plan", "apply", "plan_only"
}

func (c *Client) ListJobs(orgName, wsName string) ([]Job, error) {
	var jobs []Job
	path := fmt.Sprintf("/api/organizations/%s/workspaces/%s/jobs/", orgName, wsName)
	if err := c.Get(path, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
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

// GetJobStageOutput fetches the log output for a specific job stage.
func (c *Client) GetJobStageOutput(orgName, wsName, jobID, stage string) (string, error) {
	var result struct {
		Output string `json:"output"`
	}
	path := fmt.Sprintf("/api/workers/jobs/%s/stages/%s/", jobID, stage)
	if err := c.Get(path, &result); err != nil {
		return "", err
	}
	return result.Output, nil
}
