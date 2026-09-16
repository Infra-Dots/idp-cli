package cli

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func reset(t *testing.T) {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
}

// Missing credentials should say how to fix themselves, not fail later with an
// opaque URL error from the HTTP layer.
func TestClientReportsMissingCredentials(t *testing.T) {
	tests := []struct {
		name       string
		host       string
		token      string
		wantSubstr string
	}{
		{"no host", "", "tok", "no host configured"},
		{"no token", "https://api.example", "", "no token configured"},
		{"neither", "", "", "no host configured"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reset(t)
			viper.Set("host", tc.host)
			viper.Set("token", tc.token)

			_, err := Client()
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tc.wantSubstr) {
				t.Errorf("error = %q, want it to mention %q", err, tc.wantSubstr)
			}
			if !strings.Contains(err.Error(), "idp auth login") {
				t.Errorf("error = %q, want it to point at `idp auth login`", err)
			}
		})
	}
}

func TestClientSucceedsWithCredentials(t *testing.T) {
	reset(t)
	viper.Set("host", "https://api.example")
	viper.Set("token", "tok")

	c, err := Client()
	if err != nil {
		t.Fatalf("Client: %v", err)
	}
	if c.Host != "https://api.example" || c.Token != "tok" {
		t.Errorf("client = %+v, want the configured host and token", c)
	}
}

func TestRequireOrgMentionsBothWaysToSetIt(t *testing.T) {
	reset(t)
	_, err := RequireOrg()
	if err == nil {
		t.Fatal("expected an error when --org is unset")
	}
	for _, want := range []string{"--org", "default_org"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want it to mention %q", err, want)
		}
	}
}

func TestNewOrgScopeReportsOrgBeforeCredentials(t *testing.T) {
	reset(t)
	// Nothing is configured; the org error is the one the user can act on first.
	if _, err := NewOrgScope(); err == nil || !strings.Contains(err.Error(), "--org") {
		t.Errorf("error = %v, want the missing --org to be reported", err)
	}
}

func TestNewOrgScopeResolvesAllThree(t *testing.T) {
	reset(t)
	viper.Set("host", "https://api.example")
	viper.Set("token", "tok")
	viper.Set("org", "my-org")
	viper.Set("output", "json")

	scope, err := NewOrgScope()
	if err != nil {
		t.Fatalf("NewOrgScope: %v", err)
	}
	if scope.Org != "my-org" || scope.Client == nil || scope.Printer == nil {
		t.Errorf("scope = %+v, want all three resolved", scope)
	}
	if string(scope.Printer.Format) != "json" {
		t.Errorf("printer format = %q, want json", scope.Printer.Format)
	}
}

func TestRequire(t *testing.T) {
	if err := Require("workspace", ""); err == nil {
		t.Error("expected an error for an empty value")
	} else if !strings.Contains(err.Error(), "--workspace") {
		t.Errorf("error = %q, want it to name the flag", err)
	}
	if err := Require("workspace", "prod"); err != nil {
		t.Errorf("Require with a value returned %v, want nil", err)
	}
}
