package api

import (
	"encoding/json"
	"strings"
	"testing"
)

const keyJSON = `{"id":"11111111-2222-3333-4444-555555555555","name":"modules","fingerprint":"SHA256:abc",
	"public_key":"ssh-ed25519 AAAA","known_hosts":"","workspaces":["app"]}`

func TestCreateSSHKeySendsTheKeyAndOmitsUnsetKnownHosts(t *testing.T) {
	c, reqs := newTestClient(t, jsonHandler(201, keyJSON))
	key, err := c.CreateSSHKey("acme", SSHKeyInput{Name: "modules", PrivateKey: "-----BEGIN OPENSSH PRIVATE KEY-----\n"})
	if err != nil {
		t.Fatal(err)
	}
	got := (*reqs)[0]
	if got.Method != "POST" || got.Path != "/api/organizations/acme/ssh-keys/" {
		t.Fatalf("request %s %s", got.Method, got.Path)
	}
	if !strings.HasPrefix(got.Body["private_key"].(string), "-----BEGIN OPENSSH") {
		t.Errorf("private_key not sent: %v", got.Body)
	}
	if _, ok := got.Body["known_hosts"]; ok {
		t.Error("known_hosts sent although unset")
	}
	if key.Fingerprint != "SHA256:abc" || key.Workspaces[0] != "app" {
		t.Errorf("key %+v", key)
	}
}

func TestResolveSSHKeyByNameOrID(t *testing.T) {
	c, _ := newTestClient(t, jsonHandler(200, "["+keyJSON+"]"))
	for _, ref := range []string{"modules", "11111111-2222-3333-4444-555555555555"} {
		key, err := c.ResolveSSHKey("acme", ref)
		if err != nil || key.Name != "modules" {
			t.Errorf("%s: %v %v", ref, key, err)
		}
	}
	if _, err := c.ResolveSSHKey("acme", "missing"); err == nil || !strings.Contains(err.Error(), "idp ssh-key list") {
		t.Errorf("got %v", err)
	}
}

func TestUpdateSSHKeyCanClearKnownHosts(t *testing.T) {
	c, reqs := newTestClient(t, jsonHandler(200, keyJSON))
	empty := ""
	if _, err := c.UpdateSSHKey("acme", "k1", SSHKeyInput{KnownHosts: &empty}); err != nil {
		t.Fatal(err)
	}
	got := (*reqs)[0]
	if got.Method != "PATCH" || got.Path != "/api/organizations/acme/ssh-keys/k1/" {
		t.Fatalf("request %s %s", got.Method, got.Path)
	}
	if v, ok := got.Body["known_hosts"]; !ok || v != "" {
		t.Errorf("known_hosts not cleared: %v", got.Body)
	}
	if _, ok := got.Body["private_key"]; ok {
		t.Error("private_key sent although unchanged")
	}
}

func TestWorkspaceSSHKeyCanBeSetClearedOrLeftAlone(t *testing.T) {
	encode := func(in UpdateWorkspaceInput) map[string]any {
		raw, _ := json.Marshal(in)
		var out map[string]any
		_ = json.Unmarshal(raw, &out)
		return out
	}
	if _, ok := encode(UpdateWorkspaceInput{Branch: "main"})["ssh_key"]; ok {
		t.Error("ssh_key sent although untouched")
	}
	id := "k1"
	set := &id
	if encode(UpdateWorkspaceInput{SSHKey: &set})["ssh_key"] != "k1" {
		t.Error("ssh_key not set")
	}
	var none *string
	cleared := encode(UpdateWorkspaceInput{SSHKey: &none})
	if v, ok := cleared["ssh_key"]; !ok || v != nil {
		t.Errorf("ssh_key not sent as null: %v", cleared)
	}
}
