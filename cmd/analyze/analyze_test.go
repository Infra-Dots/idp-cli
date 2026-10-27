package analyze

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/cli"
	"github.com/infradots/idp-cli/internal/output"
)

type req struct {
	Method, Path string
	Body         map[string]any
}

// newScope serves handler and returns an org scope pointed at it, its captured stdout, and the requests seen.
func newScope(t *testing.T, format string, handler func(w http.ResponseWriter, r *http.Request, n int)) (*cli.OrgScope, *bytes.Buffer, *[]req) {
	t.Helper()
	var (
		mu   sync.Mutex
		reqs []req
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := req{Method: r.Method, Path: r.URL.Path}
		_ = json.NewDecoder(r.Body).Decode(&rec.Body)
		mu.Lock()
		reqs = append(reqs, rec)
		n := len(reqs)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		handler(w, r, n)
	}))
	t.Cleanup(srv.Close)
	buf := &bytes.Buffer{}
	p := output.New(format, false)
	p.Out = buf
	return &cli.OrgScope{Client: api.NewClient(srv.URL, "tok"), Printer: p, Org: "acme"}, buf, &reqs
}

func reply(w http.ResponseWriter, status int, body string) {
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

const proposed = `{"id":"a1","status":"proposed","environment":"prod","needs":[],"mentions":[],
 "repositories":[{"label":"R1","url":"github.com/x/y","commit_sha":"abc123","error":""}],
 "proposal":{"summary":"one db","questions":["which region?"],
  "targets":[{"workspace_name":"ws1","selected":true,"kinds":["postgresql"]},
             {"workspace_name":"ws2","selected":false,"kinds":["s3"]},
             {"workspace_name":"ws3","selected":true,"kinds":["redis"]}],
  "decisions":[{"kind":"postgresql","decision":"provision","target_workspace":"ws1","resource_type":"aws_db_instance",
   "rank":1,"below_top_rank":false,"unverified":true,"rationale":"needs a db"}]}}`

func TestRepoSendsNeedsTagsNotesAndSourceCLI(t *testing.T) {
	scope, buf, reqs := newScope(t, "table", func(w http.ResponseWriter, r *http.Request, n int) {
		reply(w, 202, `{"id":"a1","status":"queued","environment":"prod",
		 "needs":[{"alias":"postgresql","kinds":["aws_rds"],"pin":[]}],
		 "mentions":[{"alias":"cache","kinds":["redis"],"pin":[],"start":0,"end":5,"text":"cache"}]}`)
	})
	err := runRepo(scope, repoOptions{
		urls: []string{"github.com/x/y", "github.com/x/z"}, environment: "prod", ref: "main", vcs: "gh",
		tags: []string{"team=payments", "app=y"}, needs: []string{"postgresql", "s3"}, notes: "a cache please",
	})
	if err != nil {
		t.Fatal(err)
	}
	got := (*reqs)[0]
	if got.Method != "POST" || got.Path != "/api/agents/acme/repository-analyses/" {
		t.Fatalf("request %s %s", got.Method, got.Path)
	}
	if got.Body["source"] != "cli" || got.Body["environment"] != "prod" || got.Body["notes"] != "a cache please" {
		t.Errorf("body %v", got.Body)
	}
	if tags := got.Body["tags"].(map[string]any); tags["team"] != "payments" || tags["app"] != "y" || len(tags) != 2 {
		t.Errorf("tags %v", tags)
	}
	if needs := got.Body["needs"].([]any); len(needs) != 2 || needs[1] != "s3" {
		t.Errorf("needs %v", needs)
	}
	repos := got.Body["repositories"].([]any)
	if len(repos) != 2 || repos[0].(map[string]any)["ref"] != "main" || repos[1].(map[string]any)["vcs"] != "gh" {
		t.Errorf("repositories %v", repos)
	}
	out := buf.String()
	for _, want := range []string{"a1", "queued", "postgresql -> aws_rds", `"cache" (cache) -> redis`} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
}

func TestRepoRejectsUnknownTagKey(t *testing.T) {
	scope, _, reqs := newScope(t, "table", nil)
	if err := runRepo(scope, repoOptions{urls: []string{"x"}, tags: []string{"owner=me"}}); err == nil {
		t.Fatal("tag key accepted")
	}
	if len(*reqs) != 0 {
		t.Error("a request was sent")
	}
}

func TestRepoWaitPollsToTerminalStatusAndShowsProposal(t *testing.T) {
	pollInterval = time.Millisecond
	t.Cleanup(func() { pollInterval = 5 * time.Second })
	scope, buf, reqs := newScope(t, "table", func(w http.ResponseWriter, r *http.Request, n int) {
		switch {
		case r.Method == "POST":
			reply(w, 202, `{"id":"a1","status":"queued"}`)
		case n < 4:
			reply(w, 200, `{"id":"a1","status":"scanning"}`)
		default:
			reply(w, 200, proposed)
		}
	})
	if err := runRepo(scope, repoOptions{urls: []string{"x"}, wait: true, timeout: time.Minute}); err != nil {
		t.Fatal(err)
	}
	if len(*reqs) != 4 {
		t.Errorf("%d requests, want POST + 3 GETs", len(*reqs))
	}
	out := buf.String()
	for _, want := range []string{"proposed", "[x] ws1", "[ ] ws2", "aws_db_instance", "unverified", "which region?"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
}

func TestRepoWaitTimesOut(t *testing.T) {
	pollInterval = time.Millisecond
	t.Cleanup(func() { pollInterval = 5 * time.Second })
	scope, _, _ := newScope(t, "table", func(w http.ResponseWriter, r *http.Request, n int) {
		reply(w, 200, `{"id":"a1","status":"analyzing"}`)
	})
	err := runRepo(scope, repoOptions{urls: []string{"x"}, wait: true, timeout: 20 * time.Millisecond})
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Errorf("got %v", err)
	}
}

func TestShowJSONPrintsTheRawResponse(t *testing.T) {
	scope, buf, _ := newScope(t, "json", func(w http.ResponseWriter, r *http.Request, n int) {
		reply(w, 200, `{"id":"a1","status":"proposed","field_the_cli_does_not_model":7}`)
	})
	a, err := scope.Client.GetAnalysis("acme", "a1")
	if err != nil {
		t.Fatal(err)
	}
	if err := showAnalysis(scope, a); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "field_the_cli_does_not_model") {
		t.Errorf("raw field lost:\n%s", buf)
	}
}

func TestConfirmWithNoFlagsSendsEmptyBody(t *testing.T) {
	scope, buf, reqs := newScope(t, "table", func(w http.ResponseWriter, r *http.Request, n int) {
		if r.Method == "GET" {
			reply(w, 200, proposed)
			return
		}
		reply(w, 201, `{"request_group_id":"g1","targets":[{"workspace_name":"ws1","status":"dispatched"},
		 {"workspace_name":"ws3","status":"blocked","reason":"run in progress"}]}`)
	})
	if err := runConfirm(scope, confirmOptions{id: "a1", yes: true}, strings.NewReader("")); err != nil {
		t.Fatal(err)
	}
	post := (*reqs)[1]
	if post.Path != "/api/agents/acme/repository-analyses/a1/confirm/" || len(post.Body) != 0 {
		t.Errorf("post %s %v", post.Path, post.Body)
	}
	for _, want := range []string{"dispatched", "blocked", "run in progress", "g1"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, buf)
		}
	}
}

func TestConfirmOnlyAndIncludeBuildTargets(t *testing.T) {
	selected := func(body map[string]any) map[string]bool {
		out := map[string]bool{}
		for _, t := range body["targets"].([]any) {
			m := t.(map[string]any)
			out[m["workspace_name"].(string)] = m["selected"].(bool)
		}
		return out
	}
	handler := func(w http.ResponseWriter, r *http.Request, n int) {
		if r.Method == "GET" {
			reply(w, 200, proposed)
			return
		}
		reply(w, 201, `{"request_group_id":"g1","targets":[]}`)
	}

	scope, _, reqs := newScope(t, "table", handler)
	if err := runConfirm(scope, confirmOptions{id: "a1", yes: true, only: []string{"ws2"}}, strings.NewReader("")); err != nil {
		t.Fatal(err)
	}
	if got := selected((*reqs)[1].Body); got["ws1"] || !got["ws2"] || got["ws3"] || len(got) != 3 {
		t.Errorf("--only: %v", got)
	}

	scope, _, reqs = newScope(t, "table", handler)
	if err := runConfirm(scope, confirmOptions{id: "a1", yes: true, include: []string{"ws2"}}, strings.NewReader("")); err != nil {
		t.Fatal(err)
	}
	if got := selected((*reqs)[1].Body); !got["ws1"] || !got["ws2"] || !got["ws3"] {
		t.Errorf("--include: %v", got)
	}

	scope, _, _ = newScope(t, "table", handler)
	err := runConfirm(scope, confirmOptions{id: "a1", yes: true, only: []string{"nope"}}, strings.NewReader(""))
	if err == nil || !strings.Contains(err.Error(), "nope") {
		t.Errorf("unknown workspace: %v", err)
	}
}

func TestConfirmAsksUnlessYes(t *testing.T) {
	scope, buf, reqs := newScope(t, "table", func(w http.ResponseWriter, r *http.Request, n int) {
		reply(w, 200, proposed)
	})
	err := runConfirm(scope, confirmOptions{id: "a1"}, strings.NewReader("n\n"))
	if err == nil || !strings.Contains(err.Error(), "aborted") {
		t.Errorf("got %v", err)
	}
	if len(*reqs) != 1 {
		t.Error("confirmed despite the answer being no")
	}
	if !strings.Contains(buf.String(), "ws1") || strings.Contains(buf.String(), "ws2") {
		t.Errorf("prompt should list the selected targets only:\n%s", buf)
	}
}

func TestErrorExtrasArePrinted(t *testing.T) {
	scope, _, _ := newScope(t, "table", func(w http.ResponseWriter, r *http.Request, n int) {
		reply(w, 400, `{"error":"Unknown environment.","environments":["prod","staging"],"known_kinds":["postgresql","s3"]}`)
	})
	err := runRepo(scope, repoOptions{urls: []string{"x"}, environment: "dev"})
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, want := range []string{"Unknown environment.", "environments: prod, staging", "known_kinds: postgresql, s3"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error lacks %q: %v", want, err)
		}
	}
}

func TestNotFoundSaysTheFeatureIsNotEnabled(t *testing.T) {
	scope, _, _ := newScope(t, "table", func(w http.ResponseWriter, r *http.Request, n int) {
		reply(w, 404, `{"error":"Not found."}`)
	})
	_, err := scope.Client.GetAnalysis("acme", "a1")
	if err == nil || !strings.Contains(err.Error(), "not enabled") {
		t.Errorf("got %v", err)
	}
}

func TestRepoNotVisibleExplainsTheConnection(t *testing.T) {
	scope, _, _ := newScope(t, "table", func(w http.ResponseWriter, r *http.Request, n int) {
		reply(w, 422, fmt.Sprintf(`{"error":"Repository not visible.","connections":["%s"]}`, "gh-main"))
	})
	err := runRepo(scope, repoOptions{urls: []string{"x"}})
	if err == nil || !strings.Contains(err.Error(), "connections: gh-main") || !strings.Contains(err.Error(), "--vcs") {
		t.Errorf("got %v", err)
	}
}
