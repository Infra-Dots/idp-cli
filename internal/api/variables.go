package api

import "fmt"

// Variable categories accepted by the API.
const (
	CategoryTerraform = "terraform"
	CategoryEnv       = "env"
)

// Variable represents an org or workspace variable.
type Variable struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Sensitive   bool   `json:"sensitive"`
	HCL         bool   `json:"hcl"`
	Category    string `json:"category"`
	Description string `json:"description,omitempty"`
	Workspace   string `json:"workspace,omitempty"`
}

// SetVariableInput holds fields for creating or updating a variable.
//
// The bools deliberately omit `omitempty`: an explicit false is meaningful on
// update (clearing `sensitive`), and omitting it would silently keep the old value.
type SetVariableInput struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Sensitive   bool   `json:"sensitive"`
	HCL         bool   `json:"hcl"`
	Category    string `json:"category,omitempty"`
	Description string `json:"description,omitempty"`
}

func (c *Client) ListOrgVariables(orgName string) ([]Variable, error) {
	return getList[Variable](c, fmt.Sprintf("/api/organizations/%s/variables/", orgName))
}

func (c *Client) ListWorkspaceVariables(orgName, wsName string) ([]Variable, error) {
	return getList[Variable](c, fmt.Sprintf("/api/organizations/%s/workspaces/%s/variables/", orgName, wsName))
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

// UpdateVariable patches an existing variable. Variables have a single detail
// route scoped to the organization, whatever their own scope — there is no
// workspace-nested detail route.
func (c *Client) UpdateVariable(orgName, varID string, in SetVariableInput) (*Variable, error) {
	var v Variable
	path := fmt.Sprintf("/api/organizations/%s/variables/%s/", orgName, varID)
	if err := c.Patch(path, in, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// DeleteVariable removes a variable by ID, org-level or workspace-level alike —
// see UpdateVariable for why there is only one route.
func (c *Client) DeleteVariable(orgName, varID string) error {
	return c.Delete(fmt.Sprintf("/api/organizations/%s/variables/%s/", orgName, varID))
}
