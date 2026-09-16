package api

import "fmt"

// Organization represents an InfraDots organization.
type Organization struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	ExecutionMode         string `json:"execution_mode"`
	AgentsEnabled         bool   `json:"agents_enabled"`
	DriftDetectionEnabled bool   `json:"drift_detection_enabled"`
	IsTrial               bool   `json:"is_trial"`
	TrialEnddate          string `json:"trial_enddate"`
	CreatedAt             string `json:"created_at"`
	UpdatedAt             string `json:"updated_at"`
}

// UserToken represents a personal API token.
//
// Token carries the signed JWT, which the API returns only in the response to
// creation. It is never retrievable afterwards, so `token create` has to print it.
type UserToken struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Token       string `json:"token,omitempty"`
	Created     string `json:"created"`
	LastUsed    string `json:"last_used,omitempty"`
	Expiration  string `json:"expiration,omitempty"`
}

// CreateTokenInput holds the fields for minting a token. ExpirationDays of 0
// means the API's default (3650 days).
type CreateTokenInput struct {
	Description string `json:"description"`
	Expiration  int    `json:"expiration,omitempty"`
}

// ListOrganizations returns all organizations the authenticated user belongs to.
func (c *Client) ListOrganizations() ([]Organization, error) {
	return getList[Organization](c, "/api/organizations/")
}

// GetOrganization returns a single organization by name.
func (c *Client) GetOrganization(name string) (*Organization, error) {
	var org Organization
	if err := c.Get(fmt.Sprintf("/api/organizations/%s/", name), &org); err != nil {
		return nil, err
	}
	return &org, nil
}

// ListTokens returns the current user's API tokens.
func (c *Client) ListTokens() ([]UserToken, error) {
	return getList[UserToken](c, "/api/users/tokens/")
}

// CreateToken mints a new personal API token. The returned UserToken carries
// the only copy of the secret the caller will ever see.
func (c *Client) CreateToken(in CreateTokenInput) (*UserToken, error) {
	var token UserToken
	if err := c.Post("/api/users/tokens/", in, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

// RevokeToken deletes an API token by ID.
func (c *Client) RevokeToken(tokenID string) error {
	return c.Delete(fmt.Sprintf("/api/users/tokens/%s/", tokenID))
}
