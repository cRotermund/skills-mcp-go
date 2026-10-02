package skills

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/Masterminds/semver/v3"
)

type repositorySkill struct {
	skill   Skill
	version *semver.Version
}

type Repository struct {
	mu     sync.RWMutex
	skills map[string][]repositorySkill
}

func NewRepository(skills []Skill) (*Repository, error) {
	repository := &Repository{skills: make(map[string][]repositorySkill)}
	if err := repository.Replace(skills); err != nil {
		return nil, err
	}
	return repository, nil
}

func (r *Repository) Replace(skills []Skill) error {
	next := make(map[string][]repositorySkill)
	for _, skill := range skills {
		version, err := semver.NewVersion(skill.Version())
		if err != nil {
			return fmt.Errorf("skill %q has invalid version %q: %w", skill.Name(), skill.Version(), err)
		}
		name := strings.TrimSpace(skill.Name())
		if name == "" {
			return fmt.Errorf("skill name is empty")
		}
		for _, existing := range next[name] {
			if existing.version.Equal(version) {
				return fmt.Errorf("duplicate skill %s@%s", name, skill.Version())
			}
		}
		next[name] = append(next[name], repositorySkill{skill: skill, version: version})
	}
	for name := range next {
		sort.Slice(next[name], func(i, j int) bool { return next[name][i].version.GreaterThan(next[name][j].version) })
	}
	r.mu.Lock()
	r.skills = next
	r.mu.Unlock()
	return nil
}

func (r *Repository) List() []Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Skill, 0)
	for _, versions := range r.skills {
		for _, item := range versions {
			result = append(result, item.skill)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name() == result[j].Name() {
			left, _ := semver.NewVersion(result[i].Version())
			right, _ := semver.NewVersion(result[j].Version())
			return left.GreaterThan(right)
		}
		return result[i].Name() < result[j].Name()
	})
	return result
}

func (r *Repository) Resolve(name, constraint string) (Skill, error) {
	r.mu.RLock()
	versions := append([]repositorySkill(nil), r.skills[name]...)
	available := make([]string, 0, len(r.skills))
	for availableName := range r.skills {
		available = append(available, availableName)
	}
	r.mu.RUnlock()
	if len(versions) == 0 {
		sort.Strings(available)
		if len(available) == 0 {
			return Skill{}, fmt.Errorf("skill %q not found; no skills are available", name)
		}
		return Skill{}, fmt.Errorf("skill %q not found; available skills: %s", name, strings.Join(available, ", "))
	}
	if strings.TrimSpace(constraint) == "" {
		return versions[0].skill, nil
	}
	query, err := semver.NewConstraint(constraint)
	if err != nil {
		return Skill{}, fmt.Errorf("invalid version constraint %q: %w", constraint, err)
	}
	for _, item := range versions {
		if query.Check(item.version) {
			return item.skill, nil
		}
	}
	return Skill{}, fmt.Errorf("skill %q has no version satisfying %q", name, constraint)
}

func (r *Repository) Get(name string) (Skill, bool) {
	skill, err := r.Resolve(name, "")
	return skill, err == nil
}

func (r *Repository) ReadAsset(name, constraint, relativePath string) (Asset, string, error) {
	skill, err := r.Resolve(name, constraint)
	if err != nil {
		return Asset{}, "", err
	}
	asset, err := readAsset(skill, relativePath)
	if err != nil {
		return Asset{}, "", err
	}
	return asset, skill.Version(), nil
}
