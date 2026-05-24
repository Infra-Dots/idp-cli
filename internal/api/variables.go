package api

import "fmt"

// Variable represents an org or workspace variable.
type Variable struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Sensitive   bool   `json:"sensitive"`
	HCL         bool   `json:"hcl"`
	Category    string `json:"category"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// SetVariableInput holds fields for creating/updating a variable.
type SetVariableInput struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Sensitive bool   `json:"sensitive,omitempty"`
	HCL       bool   `json:"hcl,omitempty"`
}

func (c *Client) ListOrgVariables(orgName string) ([]Variable, error) {
	var vars []Variable
	path := fmt.Sprintf("/api/organizations/%s/variables/", orgName)
	if err := c.Get(path, &vars); err != nil {
		return nil, err
	}
	return vars, nil
}

func (c *Client) ListWorkspaceVariables(orgName, wsName string) ([]Variable, error) {
	var vars []Variable
	path := fmt.Sprintf("/api/organizations/%s/workspaces/%s/variables/", orgName, wsName)
	if err := c.Get(path, &vars); err != nil {
		return nil, err
	}
	return vars, nil
}

func (c *Client) CreateOrgVariable(orgName string, in SetVariableInput) (*Variable, error) {
	var v Variable
	path := fmt.Sprintf("/api/organizations/%s/variables/", orgName)
	if err := c.Post(path, in, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (c *Client) CreateWorkspaceVariable(orgName, wsName string, in SetVariableInput) (*Variable, error) {
	var v Variable
	path := fmt.Sprintf("/api/organizations/%s/workspaces/%s/variables/", orgName, wsName)
	if err := c.Post(path, in, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (c *Client) DeleteOrgVariable(orgName, varID string) error {
	path := fmt.Sprintf("/api/organizations/%s/variables/%s/", orgName, varID)
	return c.Delete(path)
}

func (c *Client) DeleteWorkspaceVariable(orgName, wsName, varID string) error {
	path := fmt.Sprintf("/api/organizations/%s/workspaces/%s/variables/%s/", orgName, wsName, varID)
	return c.Delete(path)
}
