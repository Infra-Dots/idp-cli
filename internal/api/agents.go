package api

import "fmt"

// AgentHistory represents a single agent execution record.
type AgentHistory struct {
	// An integer, not a UUID: AgentHistory is the one model in idp's ai app with an implicit
	// AutoField pk, while its neighbours (conversations, skills, plugins…) use UUIDs. Typed as a
	// string here, every `idp agent list` failed with "cannot unmarshal number into ... .id".
	ID          int    `json:"id"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	PRURL       string `json:"pr_url"`
	Implemented bool   `json:"implemented"`
	Workspace   string `json:"workspace"`
	Job         string `json:"job"`
	// The API names this `timestamp`; there is no updated field.
	Timestamp string `json:"timestamp"`
}

// ListAgentHistory returns agent executions across an organization.
func (c *Client) ListAgentHistory(orgName string) ([]AgentHistory, error) {
	return getList[AgentHistory](c, fmt.Sprintf("/api/agents/history/%s/", orgName))
}

// ListWorkspaceAgentHistory returns agent executions for a single workspace.
func (c *Client) ListWorkspaceAgentHistory(wsName string) ([]AgentHistory, error) {
	return getList[AgentHistory](c, fmt.Sprintf("/api/agents/history/workspace/%s/", wsName))
}

// GetJobAgentHistory returns the agent executions for one job. A job can have
// several (a review, then an implementation), so this is always a list.
func (c *Client) GetJobAgentHistory(jobID string) ([]AgentHistory, error) {
	return getList[AgentHistory](c, fmt.Sprintf("/api/agents/history/%s/", jobID))
}
