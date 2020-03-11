package linter

        import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/atosatto/ansible-requirements-lint/provider"
	"github.com/atosatto/ansible-requirements-lint/requirements"
)

var errMockRoleNotFound = fmt.Errorf("role not found")

type mockGitProvider struct{}

func (g mockGitProvider) VersionsForRole(ctx context.Context, r requirements.Role) ([]string, error) {
	switch {
	case r.Source == "https://github.com/test/ansible-requirements-lint":
		return []string{"v1.0.0", "v1.1.0"}, nil
	default:
		return nil, errMockRoleNotFound
	}
}

type mockAnsibleGalaxyProvider struct{}

func (g mockAnsibleGalaxyProvider) VersionsForRole(ctx context.Context, r requirements.Role) ([]string, error) {
	switch {
	case r.Source == "test.ansible-requirements-lint":
		return []string{"v1.0.0", "v1.1.0"}, nil
	default:
		return nil, errMockRoleNotFound
	}
}

func TestUpdatesLinter(t *testing.T) {
	updatesLinter := &UpdatesLinter{
		rolesProviders: map[string]provider.RolesProvider{
			git:           mockGitProvider{},
			ansibleGalaxy: mockAnsibleGalaxyProvider{},
		},
	}

	// test cases
	cases := map[string]struct {
		role   requirements.Role
		update Update
		level  Level
		err    error
	}{
		"galaxy:update": {
			role: requirements.Role{
				Source:  "test.ansible-requirements-lint",
				Version: "v1.0.0",
			},
			update: Update{
				FromVersion: "v1.0.0",
				ToVersion:   "v1.1.0",
				IsUpdate:    true,
			},
			level: LevelWarning,
		},
		"galaxy:latest": {
			role: requirements.Role{
				Source:  "test.ansible-requirements-lint",
				Version: "v1.1.0",
			},
			update: Update{
				FromVersion: "v1.1.0",
				ToVersion:   "v1.1.0",
				IsUpdate:    false,
			},
			level: LevelInfo,
		},
		"galaxy:noVersion": {
			role: requirements.Role{
				Source: "test.ansible-requirements-lint",
			},
			update: Update{
				ToVersion: "v1.1.0",
				IsUpdate:  false,
			},
			level: LevelWarning,
		},
		"galaxy:notFound": {
			role: requirements.Role{
				Source:  "test.ansible-requirements-lint-notfound",
				Version: "v1.0.0",
			},
			level: LevelError,
			err:   errMockRoleNotFound,
		},
		// no scm (defaulting to git)
		"noscm:update": {
			role: requirements.Role{
				Source:  "https://github.com/test/ansible-requirements-lint",
				Version: "v1.0.0",
			},
			update: Update{
				FromVersion: "v1.0.0",
				ToVersion:   "v1.1.0",
				IsUpdate:    true,
			},
			level: LevelWarning,
		},
		"noscm:latest": {
			role: requirements.Role{
				Source:  "https://github.com/test/ansible-requirements-lint",
				Version: "v1.1.0",
			},
			update: Update{
				FromVersion: "v1.1.0",
				ToVersion:   "v1.1.0",
				IsUpdate:    false,
			},
			level: LevelInfo,
		},
		"noscm:notFound": {
			role: requirements.Role{
				Source:  "https://github.com/test/ansible-requirements-lint-notfound",
				Version: "v1.1.0",
			},
			level: LevelError,
			err:   errMockRoleNotFound,
		},
		"noscm:tarball": {
			role: requirements.Role{
				Source:  "https://github.com/test/ansible-requirements-lint.tar.gz",
				Version: "v1.0.0",
			},
			level: LevelInfo,
		},
		"git:update": {
			role: requirements.Role{
				Source:  "https://github.com/test/ansible-requirements-lint",
				Scm:     "git",
				Version: "v1.0.0",
			},
			update: Update{
				FromVersion: "v1.0.0",
				ToVersion:   "v1.1.0",
				IsUpdate:    true,
			},
			level: LevelWarning,
		},
		"git:latest": {
			role: requirements.Role{
				Source:  "https://github.com/test/ansible-requirements-lint",
				Scm:     "git",
				Version: "v1.0.0",
			},
			update: Update{
				FromVersion: "v1.0.0",
				ToVersion:   "v1.1.0",
				IsUpdate:    true,
			},
			level: LevelWarning,
		},
		"git:branch": {
			role: requirements.Role{
				Source:  "https://github.com/test/ansible-requirements-lint",
				Scm:     "git",
				Version: "master",
			},
			update: Update{
				FromVersion: "master",
				ToVersion:   "v1.1.0",
				IsUpdate:    false,
			},
			level: LevelWarning,
		},
		"git:notfound": {
			role: requirements.Role{
				Source:  "https://github.com/test/ansible-requirements-lint-notfound",
				Scm:     "git",
				Version: "v1.0.0",
			},
			level: LevelError,
			err:   errMockRoleNotFound,
		},
		"uknownscm": {
			role: requirements.Role{
				Source:  "https://github.com/test/ansible-requirements-lint",
				Scm:     "hg",
				Version: "v1.0.0",
			},
			level: LevelError,
			err:   NewUnknownScmError("hg"),
		},
	}

	roles := make(chan requirements.Role)
	results := make(chan Result)
	go updatesLinter.RunWithContext(context.Background(), roles, results)
	for k, c := range cases {
		roles <- c.role
		res := <-results

		if !reflect.DeepEqual(c.role, res.Role) {
			t.Errorf("%s: expecting role %+v, found %+v", k, c.role, res.Role)
		}
		if !reflect.DeepEqual(c.update, Update{}) && !reflect.DeepEqual(c.update, res.Metadata) {
			if len(res.Msg) > 0 {
				t.Errorf("%s: expecting update %+v, obtained update %+v with msg: '%s'", k, c.update, res.Metadata, res.Msg)
			} else if res.Err != nil {
				t.Errorf("%s: expecting update %+v, obtained update %+v with err: '%s'", k, c.update, res.Metadata, res.Err)
			}
		}
		if !reflect.DeepEqual(c.level, res.Level) {
			t.Errorf("%s: expecting level %+v, found %+v", k, c.level, res.Level)
		}
		if c.err != nil && c.err.Error() != res.Err.Error() {
			t.Errorf("%s: expecting error %+v, obtained %+v", k, c.err, res.Err)
		}
	}

	t.Errorf("test")
}
