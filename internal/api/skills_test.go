package api

import (
	"net/http"
	"testing"
)

// Mirrors idp's AgentSkillSerializer: an organization's list holds its rows and the platform rows it uses.
const skillsJSON = `[{"id":"p1","name":"terraform-style","platform":true,"enabled":true},
	{"id":"o1","name":"tagging","platform":false,"enabled":true,"agents":["review"]}]`

func TestSaveSkillPostsANewSkillEvenWhenAPlatformSkillHasTheName(t *testing.T) {
	c, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			jsonHandler(200, skillsJSON)(w, r)
			return
		}
		jsonHandler(201, `{"id":"o2","name":"terraform-style"}`)(w, r)
	})
	_, created, err := c.SaveSkill("acme", SkillInput{Name: "terraform-style", Instructions: "Ours."})
	if err != nil || !created {
		t.Fatalf("created=%v err=%v", created, err)
	}
	got := (*reqs)[1]
	if got.Method != "POST" || got.Path != "/api/agents/acme/skills/" {
		t.Fatalf("request %s %s, want a POST: the platform row is not the organization's to change", got.Method, got.Path)
	}
	if v, ok := got.Body["source_repo"]; !ok || v != "" {
		t.Errorf("source_repo must be sent empty so an inline skill clears a git one: %v", got.Body)
	}
}

func TestSaveSkillReplacesTheOrganizationsOwn(t *testing.T) {
	c, reqs := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			jsonHandler(200, skillsJSON)(w, r)
			return
		}
		jsonHandler(200, `{"id":"o1","name":"tagging"}`)(w, r)
	})
	if _, created, err := c.SaveSkill("acme", SkillInput{Name: "tagging"}); err != nil || created {
		t.Fatalf("created=%v err=%v", created, err)
	}
	if got := (*reqs)[1]; got.Method != "PATCH" || got.Path != "/api/agents/acme/skills/o1/" {
		t.Fatalf("request %s %s", got.Method, got.Path)
	}
}

func TestPlatformSkillsUseTheAdminAPI(t *testing.T) {
	c, reqs := newTestClient(t, jsonHandler(200, `{"count":1,"next":null,"results":[{"id":"p1","name":"terraform-style"}]}`))
	skill, err := c.FindSkill("", "terraform-style")
	if err != nil || skill == nil || skill.ID != "p1" {
		t.Fatalf("skill=%v err=%v", skill, err)
	}
	if (*reqs)[0].Path != "/api/admin/ai/skills/" {
		t.Errorf("path %s", (*reqs)[0].Path)
	}
}
