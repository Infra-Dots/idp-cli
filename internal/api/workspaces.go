package api

import "fmt"

// Workspace represents an InfraDots workspace.
type Workspace struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Source           string `json:"source"`
	Branch           string `json:"branch"`
	Folder           string `json:"folder"`
	TerraformVersion string `json:"terraform_version"`
	IacType          string `json:"iac_type"`
	ExecutionMode    string `json:"execution_mode"`
	AutoApply        bool   `json:"auto_apply"`
	AgentsEnabled    bool   `json:"agents_enabled"`
	Locked           bool   `json:"locked"`
	// SSHKey is the ID of the SSH key its module sources over SSH use, or nil.
	SSHKey *string `json:"ssh_key"`
	// The API names these `*_date`, not `*_at`.
	CreatedDate string `json:"created_date"`
	UpdatedDate string `json:"updated_date"`
}

// CreateWorkspaceInput holds the fields for workspace creation.
//
// Source, Branch and TerraformVersion are required by the API: WorkspaceSerializer
// exposes every model field, and all three are non-blank on the model.
type CreateWorkspaceInput struct {
	Name             string `json:"name"`
	Source           string `json:"source"`
	Branch           string `json:"branch"`
	TerraformVersion string `json:"terraform_version"`
	Folder           string `json:"folder,omitempty"`
	Description      string `json:"description,omitempty"`
	// VcsID is read straight from the request body by the view (the serializer
	// field is read-only), so the key must be "vcs".
	VcsID         string `json:"vcs,omitempty"`
	IacType       string `json:"iac_type,omitempty"`
	AutoApply     *bool  `json:"auto_apply,omitempty"`
	AgentsEnabled *bool  `json:"agents_enabled,omitempty"`
	SSHKey        string `json:"ssh_key,omitempty"`
}

// UpdateWorkspaceInput holds updatable workspace fields.
type UpdateWorkspaceInput struct {
	Source           string `json:"source,omitempty"`
	Branch           string `json:"branch,omitempty"`
	Folder           string `json:"folder,omitempty"`
	Description      string `json:"description,omitempty"`
	TerraformVersion string `json:"terraform_version,omitempty"`
	AutoApply        *bool  `json:"auto_apply,omitempty"`
	AgentsEnabled    *bool  `json:"agents_enabled,omitempty"`
	// SSHKey: nil leaves it as is; a pointer to nil sends null, which detaches the key.
	SSHKey **string `json:"ssh_key,omitempty"`
}

func (c *Client) ListWorkspaces(orgName string) ([]Workspace, error) {
	return getList[Workspace](c, fmt.Sprintf("/api/organizations/%s/workspaces/", orgName))
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
