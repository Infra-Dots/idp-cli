package api

import (
	"net/http"
	"testing"
)

// These tests pin the CLI's request and response shapes to what the idp API
// actually sends and accepts. The response bodies mirror the Django serializers
// field-for-field — in particular the `*_date` timestamps, which the CLI used to
// read as `*_at` and render as blank columns.

func TestCreateTokenCapturesSecret(t *testing.T) {
	// POST /api/users/tokens/ returns the signed JWT exactly once, at creation.
	// If the client drops it, the credential is gone for good.
	body := `{"id":"tok-1","description":"ci","created":"2026-08-25T10:00:00Z",
	          "last_used":null,"expiration":"2036-08-22T10:00:00Z","token":"eyJhbGciOi.signed.jwt"}`
	c, reqs := newTestClient(t, jsonHandler(201, body))

	tok, err := c.CreateToken(CreateTokenInput{Description: "ci", Expiration: 90})
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}
	if tok.Token != "eyJhbGciOi.signed.jwt" {
		t.Errorf("Token = %q, want the JWT from the response", tok.Token)
	}
	if tok.Created != "2026-08-25T10:00:00Z" {
		t.Errorf("Created = %q, want the `created` field", tok.Created)
	}
	if got := (*reqs)[0].Body["expiration"]; got != float64(90) {
		t.Errorf("expiration sent as %v, want 90", got)
	}
}

func TestCreateTokenOmitsDefaultExpiration(t *testing.T) {
	c, reqs := newTestClient(t, jsonHandler(201, `{"id":"t","token":"x"}`))
	if _, err := c.CreateToken(CreateTokenInput{Description: "ci"}); err != nil {
		t.Fatalf("CreateToken: %v", err)
	}
	if _, present := (*reqs)[0].Body["expiration"]; present {
		t.Error("expiration should be omitted when unset, letting the API apply its default")
	}
}

func TestCreateWorkspaceSendsRequiredFields(t *testing.T) {
	// WorkspaceSerializer exposes every model field, and source/branch/
	// terraform_version are all non-blank on the model — omitting any of them
	// is a 400. `repository` is not a field at all.
	c, reqs := newTestClient(t, jsonHandler(201, `{"name":"ws","source":"org/infra"}`))

	_, err := c.CreateWorkspace("my-org", CreateWorkspaceInput{
		Name:             "ws",
		Source:           "org/infra",
		Branch:           "main",
		TerraformVersion: "1.9.0",
		VcsID:            "vcs-1",
	})
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	body := (*reqs)[0].Body
	for _, key := range []string{"source", "branch", "terraform_version"} {
		if body[key] == nil || body[key] == "" {
			t.Errorf("request body is missing required field %q: %+v", key, body)
		}
	}
	if _, present := body["repository"]; present {
		t.Error("`repository` is not an API field and must not be sent")
	}
	// The view reads the VCS id straight off the request body under "vcs".
	if body["vcs"] != "vcs-1" {
		t.Errorf("vcs = %v, want vcs-1", body["vcs"])
	}
}

func TestWorkspaceDecodesDateFields(t *testing.T) {
	body := `[{"id":"w1","name":"prod","source":"org/infra","branch":"main",
	           "terraform_version":"1.9.0","auto_apply":true,"agents_enabled":false,
	           "created_date":"2026-01-02T03:04:05Z","updated_date":"2026-02-03T04:05:06Z"}]`
	c, _ := newTestClient(t, jsonHandler(200, body))

	ws, err := c.ListWorkspaces("my-org")
	if err != nil {
		t.Fatalf("ListWorkspaces: %v", err)
	}
	if ws[0].UpdatedDate != "2026-02-03T04:05:06Z" {
		t.Errorf("UpdatedDate = %q, want the `updated_date` value", ws[0].UpdatedDate)
	}
	if ws[0].Source != "org/infra" || ws[0].Branch != "main" {
		t.Errorf("source/branch not decoded: %+v", ws[0])
	}
}

func TestJobDecodesDateFields(t *testing.T) {
	body := `{"id":"j1","type":"plan","status":"Waiting Approval","created_by":"42","via":"idp-ui",
	          "created_date":"2026-01-02T03:04:05Z","updated_date":"2026-01-02T03:09:05Z"}`
	c, _ := newTestClient(t, jsonHandler(200, body))

	j, err := c.GetJob("my-org", "ws", "j1")
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if j.CreatedDate == "" || j.UpdatedDate == "" {
		t.Errorf("date fields not decoded: %+v", j)
	}
	if j.Status != "Waiting Approval" {
		t.Errorf("Status = %q, want the literal API spelling", j.Status)
	}
}

func TestJobTypeValidation(t *testing.T) {
	for _, valid := range []string{"plan", "apply", "destroy", "refresh"} {
		if !ValidJobType(valid) {
			t.Errorf("ValidJobType(%q) = false, want true", valid)
		}
	}
	// plan_only is the Python constant name, not the wire value; accepting it
	// would produce an opaque 400 from the API.
	for _, invalid := range []string{"plan_only", "PLAN", "", "destroy_all"} {
		if ValidJobType(invalid) {
			t.Errorf("ValidJobType(%q) = true, want false", invalid)
		}
	}
}

func TestDeleteVariableUsesOrgScopedRoute(t *testing.T) {
	// There is no workspace-nested variable detail route; workspace-level
	// variables are deleted through the org-scoped one.
	c, reqs := newTestClient(t, jsonHandler(204, ``))
	if err := c.DeleteVariable("my-org", "var-1"); err != nil {
		t.Fatalf("DeleteVariable: %v", err)
	}
	if got, want := (*reqs)[0].Path, "/api/organizations/my-org/variables/var-1/"; got != want {
		t.Errorf("path = %q, want %q", got, want)
	}
	if (*reqs)[0].Method != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", (*reqs)[0].Method)
	}
}

func TestSetVariableSendsExplicitBools(t *testing.T) {
	// `sensitive` and `hcl` must be sent even when false, so an update can
	// clear them rather than silently keeping the stored value.
	c, reqs := newTestClient(t, jsonHandler(200, `{"id":"v1","key":"K"}`))
	if _, err := c.UpdateVariable("my-org", "v1", SetVariableInput{
		Key: "K", Value: "v", Sensitive: false, HCL: false,
	}); err != nil {
		t.Fatalf("UpdateVariable: %v", err)
	}
	body := (*reqs)[0].Body
	for _, key := range []string{"sensitive", "hcl"} {
		if _, present := body[key]; !present {
			t.Errorf("%q must be sent explicitly, even when false: %+v", key, body)
		}
	}
	if (*reqs)[0].Method != http.MethodPatch {
		t.Errorf("method = %q, want PATCH", (*reqs)[0].Method)
	}
}

func TestVariableCategoryRoundTrips(t *testing.T) {
	c, reqs := newTestClient(t, jsonHandler(201, `{"id":"v1","key":"AWS_REGION","category":"env"}`))
	v, err := c.CreateOrgVariable("my-org", SetVariableInput{
		Key: "AWS_REGION", Value: "us-east-1", Category: CategoryEnv,
	})
	if err != nil {
		t.Fatalf("CreateOrgVariable: %v", err)
	}
	if (*reqs)[0].Body["category"] != "env" {
		t.Errorf("category not sent: %+v", (*reqs)[0].Body)
	}
	if v.Category != "env" {
		t.Errorf("Category = %q, want env", v.Category)
	}
}

func TestCreateVCSSendsOAuthCredentials(t *testing.T) {
	// The Vcs model has no token field — auth is OAuth client credentials.
	c, reqs := newTestClient(t, jsonHandler(201, `{"id":"v1","name":"gh","vcsType":"github","status":"pending"}`))
	vcs, err := c.CreateVCS("my-org", CreateVCSInput{
		Name: "gh", Description: "gh", VCSType: "github",
		ClientID: "id", ClientSecret: "secret",
	})
	if err != nil {
		t.Fatalf("CreateVCS: %v", err)
	}
	body := (*reqs)[0].Body
	for _, key := range []string{"clientId", "clientSecret", "vcsType", "description"} {
		if body[key] == nil || body[key] == "" {
			t.Errorf("request body is missing %q: %+v", key, body)
		}
	}
	if _, present := body["token"]; present {
		t.Error("`token` is not an API field and must not be sent")
	}
	if vcs.Status != "pending" {
		t.Errorf("Status = %q, want pending", vcs.Status)
	}
}

func TestAgentHistoryForJobIsAList(t *testing.T) {
	// One job can have several agent runs (a review, then an implementation),
	// so the retrieve endpoint returns an array, not an object.
	//
	// `id` is a JSON *number*: AgentHistory is the one model in idp's ai app whose pk is an
	// implicit AutoField rather than a UUID. This fixture used to quote it, which is why the
	// string-typed field looked correct in tests while `idp agent list` failed against every
	// real organization.
	body := `[{"id":1,"type":"review","status":"completed","timestamp":"2026-01-02T03:04:05Z",
	           "pr_url":"https://github.com/o/r/pull/7","implemented":true,"job":"j1","workspace":"w1"}]`
	c, _ := newTestClient(t, jsonHandler(200, body))

	history, err := c.GetJobAgentHistory("j1")
	if err != nil {
		t.Fatalf("GetJobAgentHistory: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("got %d records, want 1", len(history))
	}
	if history[0].PRURL == "" || history[0].Timestamp == "" || !history[0].Implemented {
		t.Errorf("agent fields not decoded: %+v", history[0])
	}
	if history[0].ID != 1 {
		t.Errorf("ID = %d, want 1 (numeric pk)", history[0].ID)
	}
}

func TestOrganizationHasNoDisplayName(t *testing.T) {
	// OrganizationSerializer has no display_name; these are the fields it does
	// expose that are worth showing.
	body := `[{"id":"o1","name":"my-org","execution_mode":"remote","agents_enabled":true,
	           "drift_detection_enabled":false,"is_trial":false,"created_at":"2026-01-02T03:04:05Z"}]`
	c, _ := newTestClient(t, jsonHandler(200, body))

	orgs, err := c.ListOrganizations()
	if err != nil {
		t.Fatalf("ListOrganizations: %v", err)
	}
	if orgs[0].ExecutionMode != "remote" || !orgs[0].AgentsEnabled {
		t.Errorf("org fields not decoded: %+v", orgs[0])
	}
	if orgs[0].CreatedAt == "" {
		t.Error("Organization is the one resource that really does use created_at")
	}
}

func TestListJobStagesOrdersByExecution(t *testing.T) {
	// The endpoint returns stages keyed by type, in no particular order.
	body := `{"stages":{"apply":{"id":"3","type":"apply","output":"c"},
	                    "init":{"id":"1","type":"init","output":"a"},
	                    "plan":{"id":"2","type":"plan","output":"b"}}}`
	c, reqs := newTestClient(t, jsonHandler(200, body))

	stages, err := c.ListJobStages("j1")
	if err != nil {
		t.Fatalf("ListJobStages: %v", err)
	}
	got := make([]string, len(stages))
	for i, s := range stages {
		got[i] = s.Type
	}
	want := []string{"init", "plan", "apply"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("stage order = %v, want %v", got, want)
		}
	}
	// Stage routes are keyed on the job ID alone — no org or workspace segment.
	if p := (*reqs)[0].Path; p != "/api/workers/jobs/j1/stages/" {
		t.Errorf("path = %q, want the workers stage route", p)
	}
}
