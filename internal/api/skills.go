package api

import "fmt"

// Skill is an agent skill (idp's AgentSkillSerializer): inline instructions, or a skill directory in git
// pinned to SourceCommit when it was saved. Platform skills reach every organization; an organization's
// skill of the same name replaces one, or turns it off when Enabled is false.
type Skill struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	DisplayName       string   `json:"display_name"`
	Description       string   `json:"description"`
	Instructions      string   `json:"instructions"`
	Agents            []string `json:"agents"`
	ResourceTypes     []string `json:"resource_types"`
	IacTypes          []string `json:"iac_types"`
	Enabled           bool     `json:"enabled"`
	SourceRepo        string   `json:"source_repo"`
	SourcePath        string   `json:"source_path"`
	SourceRef         string   `json:"source_ref"`
	SourceCommit      string   `json:"source_commit"`
	Platform          bool     `json:"platform"`
	OverridesPlatform bool     `json:"overrides_platform"`
}

// SkillInput creates or replaces a skill. Every field is sent: the API refuses a skill with both
// instructions and a git source, so switching one to the other has to clear the old.
type SkillInput struct {
	Name          string   `json:"name"`
	DisplayName   string   `json:"display_name"`
	Description   string   `json:"description"`
	Instructions  string   `json:"instructions"`
	Agents        []string `json:"agents"`
	ResourceTypes []string `json:"resource_types"`
	IacTypes      []string `json:"iac_types"`
	Enabled       bool     `json:"enabled"`
	SourceRepo    string   `json:"source_repo"`
	SourcePath    string   `json:"source_path"`
	SourceRef     string   `json:"source_ref"`
}

// skillsPath is an organization's skills, or with org "" the platform's (staff only).
func skillsPath(org string) string {
	if org == "" {
		return "/api/admin/ai/skills/"
	}
	return fmt.Sprintf("/api/agents/%s/skills/", org)
}

// ListSkills returns an organization's skills and the platform skills it hasn't replaced; with org "" the
// platform's own.
func (c *Client) ListSkills(org string) ([]Skill, error) {
	return getList[Skill](c, skillsPath(org))
}

// FindSkill returns the skill named name that this scope owns (never a platform one, for an organization),
// or nil.
func (c *Client) FindSkill(org, name string) (*Skill, error) {
	skills, err := c.ListSkills(org)
	if err != nil {
		return nil, err
	}
	for i := range skills {
		if skills[i].Name == name && (org == "" || !skills[i].Platform) {
			return &skills[i], nil
		}
	}
	return nil, nil
}

// SaveSkill creates the skill, or replaces the one with that name the scope already owns.
func (c *Client) SaveSkill(org string, in SkillInput) (*Skill, bool, error) {
	existing, err := c.FindSkill(org, in.Name)
	if err != nil {
		return nil, false, err
	}
	var out Skill
	if existing == nil {
		err = c.Post(skillsPath(org), in, &out)
	} else {
		err = c.Patch(skillsPath(org)+existing.ID+"/", in, &out)
	}
	if err != nil {
		return nil, false, err
	}
	return &out, existing == nil, nil
}

func (c *Client) DeleteSkill(org, id string) error {
	return c.Delete(skillsPath(org) + id + "/")
}
