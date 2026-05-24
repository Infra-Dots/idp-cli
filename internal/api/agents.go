package api

import "fmt"

// AgentHistory represents a single agent execution record.
type AgentHistory struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (c *Client) ListAgentHistory(orgName string) ([]AgentHistory, error) {
	var history []AgentHistory
	path := fmt.Sprintf("/api/agents/history/%s/", orgName)
	if err := c.Get(path, &history); err != nil {
		return nil, err
	}
	return history, nil
}

func (c *Client) GetAgentHistory(jobID string) (*AgentHistory, error) {
	var ah AgentHistory
	path := fmt.Sprintf("/api/agents/history/%s/", jobID)
	if err := c.Get(path, &ah); err != nil {
		return nil, err
	}
	return &ah, nil
}
