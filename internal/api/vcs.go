package api

import "fmt"

// VCSConnection represents a VCS provider connection.
type VCSConnection struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	VCSType   string `json:"vcsType"`
	CreatedAt string `json:"created_at"`
}

// CreateVCSInput holds fields for creating a VCS connection.
type CreateVCSInput struct {
	Name    string `json:"name"`
	VCSType string `json:"vcsType"`
	Token   string `json:"token"`
}

func (c *Client) ListVCS(orgName string) ([]VCSConnection, error) {
	var connections []VCSConnection
	path := fmt.Sprintf("/api/organizations/%s/vcs/", orgName)
	if err := c.Get(path, &connections); err != nil {
		return nil, err
	}
	return connections, nil
}

func (c *Client) CreateVCS(orgName string, in CreateVCSInput) (*VCSConnection, error) {
	var vcs VCSConnection
	path := fmt.Sprintf("/api/organizations/%s/vcs/", orgName)
	if err := c.Post(path, in, &vcs); err != nil {
		return nil, err
	}
	return &vcs, nil
}

func (c *Client) DeleteVCS(orgName, vcsID string) error {
	path := fmt.Sprintf("/api/organizations/%s/vcs/%s/", orgName, vcsID)
	return c.Delete(path)
}
