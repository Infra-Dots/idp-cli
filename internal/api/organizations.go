package api

import "fmt"

// Organization represents an InfraDots organization.
type Organization struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	CreatedAt   string `json:"created_at"`
}

// UserToken represents a personal API token.
type UserToken struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	LastUsed    string `json:"last_used,omitempty"`
}

// ListOrganizations returns all organizations the authenticated user belongs to.
func (c *Client) ListOrganizations() ([]Organization, error) {
	var orgs []Organization
	if err := c.Get("/api/organizations/", &orgs); err != nil {
		return nil, err
	}
	return orgs, nil
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
	var tokens []UserToken
	if err := c.Get("/api/users/tokens/", &tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}

// CreateToken creates a new personal API token.
func (c *Client) CreateToken(description string) (*UserToken, error) {
	body := map[string]string{"description": description}
	var token UserToken
	if err := c.Post("/api/users/tokens/", body, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

// RevokeToken deletes an API token by ID.
func (c *Client) RevokeToken(tokenID string) error {
	return c.Delete(fmt.Sprintf("/api/users/tokens/%s/", tokenID))
}
