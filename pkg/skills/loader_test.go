package skills

import "testing"

func TestEmbeddedGPIOSkillFallback(t *testing.T) {
	loader := NewSkillsLoader(t.TempDir(), "")

	content, ok := loader.LoadSkill("gpio")
	if !ok {
		t.Fatal("expected embedded gpio skill to load")
	}
	if content == "" {
		t.Fatal("expected embedded gpio skill content")
	}

	found := false
	for _, skill := range loader.ListSkills(false) {
		if skill.Name == "gpio" && skill.Source == "builtin" && skill.Available {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected embedded gpio skill in builtin skill list")
	}
}
