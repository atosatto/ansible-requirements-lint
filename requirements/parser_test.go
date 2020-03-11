package requirements

import (
	"testing"
)

func rolesEqual(a, b Role) bool {
	switch {
	case a.Name != b.Name:
		return false
	case a.Source != b.Source:
		return false
	case a.Scm != b.Scm:
		return false
	case a.Version != b.Version:
		return false
	case a.Include != b.Include:
		return false
	default:
		return true
	}
}

func parseAndCompare(t *testing.T, requirements string, expected Requirements) {
	parsed, err := Unmarshal([]byte(requirements))
	if err != nil {
		t.Errorf("expected no error, obtained %+v", err)
		return
	}

	if len(expected.Roles) != len(parsed.Roles) {
		t.Errorf("expecting %d roles, parsed %d roles", len(expected.Roles), len(parsed.Roles))
	}

	for i, r := range parsed.Roles {
		expR := expected.Roles[i]
		if !rolesEqual(*expR, *r) {
			t.Errorf("expecting role %+v, parsed %+v", expR, r)
		}
	}
}

// TestParseInlineRequirementsFile tests parsing of roles
// requirements from Ansible requirements files using the
// legacy inline syntax.
func TestParseInlineRequirementsFile(t *testing.T) {
	var requirements = `
---
# Test ansible-requirements-lint 
- name: test.ansible-requirements-lint-name
  version: v1.0.0

- src: test.ansible-requirements-lint-src
  version: v1.0.0

- src: test.ansible-requirements-lint-scm
  version: v1.0.0
  scm: git
`

	var expected = Requirements{
		Roles: []*Role{
			&Role{
				Name:    "test.ansible-requirements-lint-name",
				Version: "v1.0.0",
			},
			&Role{
				Source:  "test.ansible-requirements-lint-src",
				Version: "v1.0.0",
			},
			&Role{
				Source:  "test.ansible-requirements-lint-scm",
				Version: "v1.0.0",
				Scm:     "git",
			},
		},
	}

	parseAndCompare(t, requirements, expected)
}

// TestParseRolesAndCollectionsRequirementsFile tests parsing of roles
// requirements from Ansible requirements files using the new
// dictionary based syntax introduced to add support to Ansible collections.
func TestParseRolesAndCollectionsRequirementsFile(t *testing.T) {
	var requirements = `
---
  roles:
  - name: test.ansible-requirements-lint-name
    version: v1.0.0

  - src: test.ansible-requirements-lint-src
    version: v1.0.0

  - src: test.ansible-requirements-lint-scm
    version: v1.0.0
    scm: git
`
	var expected = Requirements{
		Roles: []*Role{
			&Role{
				Name:    "test.ansible-requirements-lint-name",
				Version: "v1.0.0",
			},
			&Role{
				Source:  "test.ansible-requirements-lint-src",
				Version: "v1.0.0",
			},
			&Role{
				Source:  "test.ansible-requirements-lint-scm",
				Version: "v1.0.0",
				Scm:     "git",
			},
		},
	}

	parseAndCompare(t, requirements, expected)
}

// TestParseMetaRequirementsFile tests parsing of roles
// requirements from Ansible roles meta definitions.
func TestParseMetaRequirementsFile(t *testing.T) {
	var requirements = `
---
  dependencies:
  - test.ansible-requirements-lint-inline

  - role: test.ansible-requirements-lint-role
    version: v1.0.0
    ansible_requirements_lint_role_variable: "ansible_requirements_lint"

  - name: test.ansible-requirements-lint-name
    version: v1.0.0

  - src: test.ansible-requirements-lint-src
    version: v1.0.0

  - src: test.ansible-requirements-lint-scm
    version: v1.0.0
    scm: git
`
	var expected = Requirements{
		Roles: []*Role{
			&Role{
				Name: "test.ansible-requirements-lint-inline",
			},
			&Role{
				Name:    "test.ansible-requirements-lint-role",
				Version: "v1.0.0",
			},
			&Role{
				Name:    "test.ansible-requirements-lint-name",
				Version: "v1.0.0",
			},
			&Role{
				Source:  "test.ansible-requirements-lint-src",
				Version: "v1.0.0",
			},
			&Role{
				Source:  "test.ansible-requirements-lint-scm",
				Version: "v1.0.0",
				Scm:     "git",
			},
		},
	}

	parseAndCompare(t, requirements, expected)
}
