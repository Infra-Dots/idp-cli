package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// recorded captures what the client actually put on the wire, so tests can
// assert on the request as well as the decoded response.
type recorded struct {
	Method string
	Path   string
	Query  string
	Auth   string
	Body   map[string]any
}

// newTestClient serves handler and returns a client pointed at it, plus a slice
// that accumulates every request the client made.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *[]recorded) {
	t.Helper()
	var reqs []recorded

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := recorded{
			Method: r.Method,
			Path:   r.URL.Path,
			Query:  r.URL.RawQuery,
			Auth:   r.Header.Get("Authorization"),
		}
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&rec.Body)
		}
		reqs = append(reqs, rec)
		handler(w, r)
	}))
	t.Cleanup(srv.Close)

	return NewClient(srv.URL, "test-token"), &reqs
}

// jsonHandler replies with a fixed body for every request.
func jsonHandler(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func TestClientSendsBearerToken(t *testing.T) {
	c, reqs := newTestClient(t, jsonHandler(200, `[]`))
	if _, err := c.ListOrganizations(); err != nil {
		t.Fatalf("ListOrganizations: %v", err)
	}
	if got := (*reqs)[0].Auth; got != "Bearer test-token" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer test-token")
	}
}

func TestDecodeListBareArray(t *testing.T) {
	items, next, err := decodeList[Organization](json.RawMessage(`[{"name":"a"},{"name":"b"}]`))
	if err != nil {
		t.Fatalf("decodeList: %v", err)
	}
	if len(items) != 2 || items[0].Name != "a" {
		t.Errorf("items = %+v, want two orgs starting with a", items)
	}
	if next != "" {
		t.Errorf("next = %q, want empty for a bare array", next)
	}
}

// The job list endpoint is the one paginated collection the CLI consumes;
// decoding it as a bare array is what broke `idp job list`.
func TestDecodeListPaginationEnvelope(t *testing.T) {
	raw := json.RawMessage(`{"count":2,"next":"https://api.example/next","results":[{"id":"j1"}]}`)
	items, next, err := decodeList[Job](raw)
	if err != nil {
		t.Fatalf("decodeList: %v", err)
	}
	if len(items) != 1 || items[0].ID != "j1" {
		t.Errorf("items = %+v, want one job j1", items)
	}
	if next != "https://api.example/next" {
		t.Errorf("next = %q, want the envelope's next URL", next)
	}
}

func TestDecodeListEmptyAndNull(t *testing.T) {
	for _, raw := range []string{``, `null`} {
		items, next, err := decodeList[Job](json.RawMessage(raw))
		if err != nil {
			t.Fatalf("decodeList(%q): %v", raw, err)
		}
		if len(items) != 0 || next != "" {
			t.Errorf("decodeList(%q) = %v/%q, want empty", raw, items, next)
		}
	}
}

// getList must follow `next` to the end, or a workspace with more than one page
// of jobs silently reports only the first 25.
func TestGetListFollowsPagination(t *testing.T) {
	var srv *httptest.Server
	page := 0
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		w.Header().Set("Content-Type", "application/json")
		if page == 1 {
			fmt.Fprintf(w, `{"count":2,"next":"%s/api/next","results":[{"id":"j1"}]}`, srv.URL)
			return
		}
		fmt.Fprint(w, `{"count":2,"next":null,"results":[{"id":"j2"}]}`)
	}))
	t.Cleanup(srv.Close)

	jobs, err := NewClient(srv.URL, "t").ListJobs("org", "ws")
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if len(jobs) != 2 || jobs[0].ID != "j1" || jobs[1].ID != "j2" {
		t.Errorf("jobs = %+v, want both pages", jobs)
	}
}

func TestErrorMessagesAreActionable(t *testing.T) {
	tests := []struct {
		status int
		body   string
		want   string
	}{
		{401, `{"detail":"no"}`, "authentication failed — run `idp auth login`"},
		{403, `{"error":"nope"}`, "permission denied"},
		{404, `{}`, "resource not found"},
		{400, `{"source":["This field is required."]}`, `API error 400`},
	}
	for _, tc := range tests {
		c, _ := newTestClient(t, jsonHandler(tc.status, tc.body))
		_, err := c.ListOrganizations()
		if err == nil {
			t.Fatalf("status %d: expected an error", tc.status)
		}
		var apiErr *APIError
		if !errorAs(err, &apiErr) {
			t.Fatalf("status %d: error is not *APIError: %v", tc.status, err)
		}
		if apiErr.StatusCode != tc.status {
			t.Errorf("status = %d, want %d", apiErr.StatusCode, tc.status)
		}
		if tc.status != 400 && apiErr.Message != tc.want {
			t.Errorf("message = %q, want %q", apiErr.Message, tc.want)
		}
	}
}

// errorAs keeps the table test above readable.
func errorAs(err error, target **APIError) bool {
	e, ok := err.(*APIError)
	if ok {
		*target = e
	}
	return ok
}
