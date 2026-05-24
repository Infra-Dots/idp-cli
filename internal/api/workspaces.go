package api

import "fmt"

// Workspace represents an InfraDots workspace.
type Workspace struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	TerraformVersion string `json:"terraform_version"`
	AutoApply        bool   `json:"auto_apply"`
	AgentsEnabled    bool   `json:"agents_enabled"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// CreateWorkspaceInput holds the fields for workspace creation.
type CreateWorkspaceInput struct {
	Name             string `json:"name"`
	VcsID            string `json:"vcs,omitempty"`
	Repository       string `json:"repository,omitempty"`
	TerraformVersion string `json:"terraform_version,omitempty"`
	AutoApply        *bool  `json:"auto_apply,omitempty"`
	AgentsEnabled    *bool  `json:"agents_enabled,omitempty"`
}

// UpdateWorkspaceInput holds updatable workspace fields.
type UpdateWorkspaceInput struct {
	TerraformVersion string `json:"terraform_version,omitempty"`
	AutoApply        *bool  `json:"auto_apply,omitempty"`
	AgentsEnabled    *bool  `json:"agents_enabled,omitempty"`
}

func (c *Client) ListWorkspaces(orgName string) ([]Workspace, error) {
	var workspaces []Workspace
	path := fmt.Sprintf("/api/organizations/%s/workspaces/", orgName)
	if err := c.Get(path, &workspaces); err != nil {
		return nil, err
	}
	return workspaces, nil
}

func (c *Client) GetWorkspace(orgName, wsName string) (*Workspace, error) {
	var ws Workspace
	path := fmt.Sprintf("/api/organizations/%s/workspaces/%s/", orgName, wsName)
	if err := c.Get(path, &ws); err != nil {
		return nil, err
	}
	return &ws, nil
}

func (c *Client) CreateWorkspace(orgName string, in CreateWorkspaceInput) (*Workspace, error) {
	var ws Workspace
	path := fmt.Sprintf("/api/organizations/%s/workspaces/", orgName)
	if err := c.Post(path, in, &ws); err != nil {
		return nil, err
	}
	return &ws, nil
}

func (c *Client) UpdateWorkspace(orgName, wsName string, in UpdateWorkspaceInput) (*Workspace, error) {
	var ws Workspace
	path := fmt.Sprintf("/api/organizations/%s/workspaces/%s/", orgName, wsName)
	if err := c.Patch(path, in, &ws); err != nil {
		return nil, err
	}
	return &ws, nil
}

func (c *Client) DeleteWorkspace(orgName, wsName string) error {
	path := fmt.Sprintf("/api/organizations/%s/workspaces/%s/", orgName, wsName)
	return c.Delete(path)
}
