package agent

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/cli"
	"github.com/infradots/idp-cli/internal/output"
)

func newSkillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Manage the skills your agents load",
		Long: `Manage agent skills. Platform skills reach every organization; an organization's skill with the
same name replaces one, and adding one with --disabled turns it off.

Agents see each skill's one-line description and load the instructions only when a skill fits the run.
--agents, --resource-types and --iac-types limit which runs list a skill at all.

--platform manages the platform's own skills (InfraDots staff).`,
	}
	cmd.PersistentFlags().Bool("platform", false, "Manage platform skills instead of an organization's (staff)")
	cmd.AddCommand(newSkillListCmd(), newSkillAddCmd(), newSkillRemoveCmd())
	return cmd
}

// skillScope is the client, printer and organization ("" for --platform) a skill command works on.
func skillScope(cmd *cobra.Command) (*api.Client, *output.Printer, string, error) {
	if platform, _ := cmd.Flags().GetBool("platform"); platform {
		client, err := cli.Client()
		return client, cli.Printer(), "", err
	}
	scope, err := cli.NewOrgScope()
	if err != nil {
		return nil, nil, "", err
	}
	return scope.Client, scope.Printer, scope.Org, nil
}

var skillHeaders = []string{"NAME", "STATUS", "SOURCE", "SCOPE", "DESCRIPTION"}

func skillStatus(s api.Skill) string {
	switch {
	case s.Platform:
		return "platform"
	case s.OverridesPlatform && !s.Enabled:
		return "platform skill off"
	case s.OverridesPlatform:
		return "replaces platform"
	case !s.Enabled:
		return "off"
	}
	return "on"
}

func skillRow(s api.Skill) []string {
	source := "inline"
	if s.SourceRepo != "" {
		source = fmt.Sprintf("%s//%s@%.12s", s.SourceRepo, s.SourcePath, s.SourceCommit)
	}
	scope := strings.Join(append(append(append([]string{}, s.Agents...), s.ResourceTypes...), s.IacTypes...), ",")
	return []string{s.Name, skillStatus(s), source, scope, s.Description}
}

func newSkillListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the organization's skills and the platform skills it uses",
		Example: `  idp agent skill list --org my-org
  idp agent skill list --platform`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, org, err := skillScope(cmd)
			if err != nil {
				return err
			}
			skills, err := client.ListSkills(org)
			if err != nil {
				return err
			}
			if printer.Quiet {
				for _, s := range skills {
					printer.PrintID(s.Name)
				}
				return nil
			}
			rows := make([][]string, len(skills))
			for i, s := range skills {
				rows[i] = skillRow(s)
			}
			return printer.Print(skills, skillHeaders, rows)
		},
	}
}

// skillFrontmatter is what `add --file` reads from a SKILL.md (agentskills.io), including the scope a
// skill may declare in metadata.
type skillFrontmatter struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Metadata    map[string]string `yaml:"metadata"`
}

// parseSkillFile splits a SKILL.md into its frontmatter and the instructions after it.
func parseSkillFile(text string) (skillFrontmatter, string, error) {
	var front skillFrontmatter
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return front, "", fmt.Errorf("no '---' frontmatter at the top of the file")
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			if err := yaml.Unmarshal([]byte(strings.Join(lines[1:i], "\n")), &front); err != nil {
				return front, "", fmt.Errorf("frontmatter is not valid YAML: %w", err)
			}
			return front, strings.Join(lines[i+1:], "\n"), nil
		}
	}
	return front, "", fmt.Errorf("the frontmatter is not closed with '---'")
}

func splitList(s string) []string {
	out := []string{}
	for _, v := range strings.Split(s, ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func newSkillAddCmd() *cobra.Command {
	var name, displayName, description, file, repo, skillPath, ref, agents, resourceTypes, iacTypes string
	var disabled bool

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a skill, or replace the one with the same name",
		Long: `Add a skill from a SKILL.md (--file), from a skill directory in git (--repo/--path), or, with only
--name and --disabled, turn a platform skill off for the organization.

A skill that already exists under that name is replaced. A git skill is pinned to the commit --ref points to
when it is saved: run add again to move it. Organization skills pass the same input check as implement
requests; a refused skill is not saved.`,
		Example: `  idp agent skill add --org my-org --file skills/tagging/SKILL.md --resource-types 'aws_*'
  idp agent skill add --org my-org --repo https://github.com/acme/skills --path skills/iac-review --ref v1.2.0
  idp agent skill add --org my-org --name iac-module-scaffold --disabled`,
		RunE: func(cmd *cobra.Command, args []string) error {
			in := api.SkillInput{Name: name, DisplayName: displayName, Description: description, Enabled: !disabled,
				Agents: splitList(agents), ResourceTypes: splitList(resourceTypes), IacTypes: splitList(strings.ToUpper(iacTypes))}
			switch {
			case file != "" && repo != "":
				return output.NewError("--file and --repo are two kinds of skill: pass one")
			case file != "":
				data, err := os.ReadFile(file)
				if err != nil {
					return fmt.Errorf("reading %s: %w", file, err)
				}
				front, body, err := parseSkillFile(string(data))
				if err != nil {
					return output.NewError("%s: %v", file, err)
				}
				in.Instructions = body
				if in.Name == "" {
					in.Name = front.Name
				}
				if in.Description == "" {
					in.Description = front.Description
				}
				for flag, key := range map[*[]string]string{&in.Agents: "infradots-agents",
					&in.ResourceTypes: "infradots-resource-types", &in.IacTypes: "infradots-iac-types"} {
					if len(*flag) == 0 {
						*flag = splitList(front.Metadata[key])
					}
				}
				in.IacTypes = splitList(strings.ToUpper(strings.Join(in.IacTypes, ",")))
			case repo != "":
				if err := cli.Require("path", skillPath); err != nil {
					return err
				}
				in.SourceRepo, in.SourcePath, in.SourceRef = repo, skillPath, ref
				if in.Name == "" {
					in.Name = path.Base(skillPath)
				}
			case !disabled:
				return output.NewError("pass --file or --repo, or --name with --disabled to turn a platform skill off")
			}
			if err := cli.Require("name", in.Name); err != nil {
				return err
			}
			if in.DisplayName == "" {
				in.DisplayName = in.Name
			}
			client, printer, org, err := skillScope(cmd)
			if err != nil {
				return err
			}
			skill, created, err := client.SaveSkill(org, in)
			if err != nil {
				return err
			}
			if printer.Quiet {
				printer.PrintID(skill.Name)
				return nil
			}
			verb := "replaced"
			if created {
				verb = "added"
			}
			fmt.Fprintf(os.Stderr, "Skill %s %s\n", skill.Name, verb)
			return printer.Print(skill, skillHeaders, [][]string{skillRow(*skill)})
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Skill name (default: SKILL.md's name, or the --path directory)")
	cmd.Flags().StringVar(&displayName, "display-name", "", "Name shown in the dashboard (default: the skill name)")
	cmd.Flags().StringVar(&description, "description", "", "One line, at most 128 characters (default: SKILL.md's)")
	cmd.Flags().StringVar(&file, "file", "", "A SKILL.md: its frontmatter and instructions")
	cmd.Flags().StringVar(&repo, "repo", "", "https:// repository holding the skill directory")
	cmd.Flags().StringVar(&skillPath, "path", "", "The skill directory in --repo, e.g. skills/tagging")
	cmd.Flags().StringVar(&ref, "ref", "", "Branch, tag or commit of --repo (default: its default branch)")
	cmd.Flags().StringVar(&agents, "agents", "", "Only these agents: review,implement,remediate (default: all)")
	cmd.Flags().StringVar(&resourceTypes, "resource-types", "", "Only runs touching these, e.g. 'aws_iam_*,aws_security_group'")
	cmd.Flags().StringVar(&iacTypes, "iac-types", "", "Only TF, OT and/or TG workspaces")
	cmd.Flags().BoolVar(&disabled, "disabled", false, "Save the skill turned off (a platform skill's name turns that skill off)")
	return cmd
}

func newSkillRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a skill (an organization's version of a platform skill brings the platform one back)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, org, err := skillScope(cmd)
			if err != nil {
				return err
			}
			skill, err := client.FindSkill(org, args[0])
			if err != nil {
				return err
			}
			if skill == nil {
				return output.NewError("no skill %q to remove (platform skills are turned off with `idp agent skill add --name %s --disabled`)",
					args[0], args[0])
			}
			if err := client.DeleteSkill(org, skill.ID); err != nil {
				return err
			}
			fmt.Printf("Skill %s removed\n", skill.Name)
			return nil
		},
	}
}
