package skills

import "testing"

func TestRepositoryResolvesLatestAndConstraints(t *testing.T) {
	skills := []Skill{
		{Frontmatter: Frontmatter{Name: "demo", Version: "1.0.0", Description: "old"}},
		{Frontmatter: Frontmatter{Name: "demo", Version: "1.10.0", Description: "latest"}},
		{Frontmatter: Frontmatter{Name: "demo", Version: "1.5.0", Description: "middle"}},
	}
	repository, err := NewRepository(skills)
	if err != nil {
		t.Fatal(err)
	}
	latest, err := repository.Resolve("demo", "")
	if err != nil || latest.Version() != "1.10.0" {
		t.Fatalf("latest = %s, err=%v", latest.Version(), err)
	}
	selected, err := repository.Resolve("demo", ">=1.2.0 <1.6.0")
	if err != nil || selected.Version() != "1.5.0" {
		t.Fatalf("selected = %s, err=%v", selected.Version(), err)
	}
}

func TestRepositoryListsByNameAndDescendingVersion(t *testing.T) {
	repository, err := NewRepository([]Skill{
		{Frontmatter: Frontmatter{Name: "zeta", Version: "1.0.0", Description: "z"}},
		{Frontmatter: Frontmatter{Name: "alpha", Version: "1.0.0", Description: "a"}},
		{Frontmatter: Frontmatter{Name: "alpha", Version: "2.0.0", Description: "a2"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	list := repository.List()
	if len(list) != 3 || list[0].Name() != "alpha" || list[0].Version() != "2.0.0" || list[1].Version() != "1.0.0" {
		t.Fatalf("list = %#v", list)
	}
}

func TestRepositoryRejectsDuplicateVersions(t *testing.T) {
	_, err := NewRepository([]Skill{
		{Frontmatter: Frontmatter{Name: "demo", Version: "1.0.0", Description: "one"}},
		{Frontmatter: Frontmatter{Name: "demo", Version: "1.0.0", Description: "two"}},
	})
	if err == nil {
		t.Fatal("NewRepository() accepted duplicate versions")
	}
}
