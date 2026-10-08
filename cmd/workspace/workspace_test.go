package workspace

import "testing"

func TestIacTypeCode(t *testing.T) {
	for in, want := range map[string]string{
		"": "", "terraform": "TF", "opentofu": "OT", "tofu": "OT", "terragrunt": "TG",
		"TF": "TF", "OT": "OT", "TG": "TG", " Terragrunt ": "TG",
	} {
		got, err := iacTypeCode(in)
		if err != nil || got != want {
			t.Errorf("iacTypeCode(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	if _, err := iacTypeCode("pulumi"); err == nil {
		t.Error("an unknown tool was accepted")
	}
}
