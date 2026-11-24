package agent

import (
	"reflect"
	"testing"
)

func TestParseSkillFileReadsFrontmatterScopeAndInstructions(t *testing.T) {
	front, body, err := parseSkillFile("---\r\nname: tagging\r\ndescription: How we tag.\r\nmetadata:\r\n" +
		"  infradots-resource-types: \"aws_*, google_*\"\r\n---\r\n# Tagging\r\n---\r\nEvery resource.")
	if err != nil {
		t.Fatal(err)
	}
	if front.Name != "tagging" || front.Description != "How we tag." {
		t.Errorf("front %+v", front)
	}
	if got := splitList(front.Metadata["infradots-resource-types"]); !reflect.DeepEqual(got, []string{"aws_*", "google_*"}) {
		t.Errorf("resource types %v", got)
	}
	if body != "# Tagging\n---\nEvery resource." {
		t.Errorf("body %q: only the first closing --- ends the frontmatter", body)
	}
}

func TestParseSkillFileNeedsFrontmatter(t *testing.T) {
	for _, text := range []string{"# no frontmatter", "---\nname: x\n"} {
		if _, _, err := parseSkillFile(text); err == nil {
			t.Errorf("%q parsed", text)
		}
	}
}
