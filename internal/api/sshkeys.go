package api

import (
	"fmt"
	"regexp"
)

// SSHKey is an organization's SSH key for module sources fetched over SSH. The API never returns the
// private key: only its public half and fingerprint.
type SSHKey struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	PublicKey   string   `json:"public_key"`
	Fingerprint string   `json:"fingerprint"`
	KnownHosts  string   `json:"known_hosts"`
	Workspaces  []string `json:"workspaces"`
	CreatedAt   string   `json:"created_at"`
}

// SSHKeyInput creates or updates a key. Empty fields are left out of an update; KnownHosts is a pointer
// so it can be cleared ("").
type SSHKeyInput struct {
	Name       string  `json:"name,omitempty"`
	PrivateKey string  `json:"private_key,omitempty"`
	KnownHosts *string `json:"known_hosts,omitempty"`
}

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func (c *Client) ListSSHKeys(orgName string) ([]SSHKey, error) {
	return getList[SSHKey](c, fmt.Sprintf("/api/organizations/%s/ssh-keys/", orgName))
}

// ResolveSSHKey finds a key by ID or by name.
func (c *Client) ResolveSSHKey(orgName, idOrName string) (*SSHKey, error) {
	keys, err := c.ListSSHKeys(orgName)
	if err != nil {
		return nil, err
	}
	for i := range keys {
		if keys[i].Name == idOrName || (uuidPattern.MatchString(idOrName) && keys[i].ID == idOrName) {
			return &keys[i], nil
		}
	}
	return nil, fmt.Errorf("no SSH key %q in organization %s (see `idp ssh-key list`)", idOrName, orgName)
}

func (c *Client) CreateSSHKey(orgName string, in SSHKeyInput) (*SSHKey, error) {
	var key SSHKey
	if err := c.Post(fmt.Sprintf("/api/organizations/%s/ssh-keys/", orgName), in, &key); err != nil {
		return nil, err
	}
	return &key, nil
}

func (c *Client) UpdateSSHKey(orgName, id string, in SSHKeyInput) (*SSHKey, error) {
	var key SSHKey
	if err := c.Patch(fmt.Sprintf("/api/organizations/%s/ssh-keys/%s/", orgName, id), in, &key); err != nil {
		return nil, err
	}
	return &key, nil
}

func (c *Client) DeleteSSHKey(orgName, id string) error {
	return c.Delete(fmt.Sprintf("/api/organizations/%s/ssh-keys/%s/", orgName, id))
}
