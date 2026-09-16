package api

import "fmt"

// VCSConnection represents a VCS provider connection.
type VCSConnection struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	VCSType     string `json:"vcsType"`
	APIURL      string `json:"apiUrl"`
	// Status is "pending" until the OAuth callback completes, and "completed"
	// once the connection can actually reach the provider.
	Status      string `json:"status"`
	Callback    string `json:"callback"`
	CreatedDate string `json:"created_date"`
}

// CreateVCSInput holds fields for creating a VCS connection.
//
// InfraDots authenticates to providers over OAuth, so a connection is created
// with an OAuth app's client credentials — there is no personal-access-token
// field. The connection stays in "pending" until the browser callback runs.
type CreateVCSInput struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	VCSType      string `json:"vcsType"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	APIURL       string `json:"apiUrl,omitempty"`
}

func (c *Client) ListVCS(orgName string) ([]VCSConnection, error) {
	return getList[VCSConnection](c, fmt.Sprintf("/api/organizations/%s/vcs/", orgName))
}

func (c *Client) GetVCS(orgName, vcsID string) (*VCSConnection, error) {
	var vcs VCSConnection
	path := fmt.Sprintf("/api/organizations/%s/vcs/%s/", orgName, vcsID)
	if err := c.Get(path, &vcs); err != nil {
		return nil, err
	}
	return &vcs, nil
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
	return c.Delete(fmt.Sprintf("/api/organizations/%s/vcs/%s/", orgName, vcsID))
}
