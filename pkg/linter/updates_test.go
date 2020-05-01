package linter

import (
	"context"
	"reflect"
	"testing"

	"github.com/atosatto/ansible-requirements-lint/pkg/errors"
	"github.com/atosatto/ansible-requirements-lint/pkg/provider"
	"github.com/atosatto/ansible-requirements-lint/pkg/types"
)

type mockGitProvider struct{}

func (g mockGitProvider) VersionsForRole(ctx context.Context, r types.Role) ([]string, error) {
	switch {
	case r.Source == "https://github.com/test/ansible-requirements-lint":
		return []string{"v1.0.0", "v1.1.0"}, nil
	default:
		return nil, errors.NewRoleNotFoundError(r, "mockGitProvider")
	}
}

type mockAnsibleGalaxyProvider struct{}

func (g mockAnsibleGalaxyProvider) VersionsForRole(ctx context.Context, r types.Role) ([]string, error) {
	switch {
	case r.Source == "test.ansible-requirements-lint":
		return []string{"v1.0.0", "v1.1.0"}, nil
	default:
		return nil, errors.NewRoleNotFoundError(r, "mockAnsibleGalaxyProvider")
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
		role   types.Role
		update Update
		level  Level
		err    error
	}{
		"galaxy:update": {
			role: types.Role{
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
			role: types.Role{
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
			role: types.Role{
				Source: "test.ansible-requirements-lint",
			},
			update: Update{
				ToVersion: "v1.1.0",
				IsUpdate:  false,
			},
			level: LevelWarning,
		},
		"galaxy:notFound": {
			role: types.Role{
				Source:  "test.ansible-requirements-lint-notfound",
				Version: "v1.0.0",
			},
			level: LevelError,
			err:   &errors.RoleNotFoundError{},
		},
		// no scm (defaulting to git)
		"noscm:update": {
			role: types.Role{
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
			role: types.Role{
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
			role: types.Role{
				Source:  "https://github.com/test/ansible-requirements-lint-notfound",
				Version: "v1.1.0",
			},
			level: LevelError,
			err:   &errors.RoleNotFoundError{},
		},
		"noscm:tarball": {
			role: types.Role{
				Source:  "https://github.com/test/ansible-requirements-lint.tar.gz",
				Version: "v1.0.0",
			},
			level: LevelInfo,
		},
		"git:update": {
			role: types.Role{
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
			role: types.Role{
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
			role: types.Role{
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
			role: types.Role{
				Source:  "https://github.com/test/ansible-requirements-lint-notfound",
				Scm:     "git",
				Version: "v1.0.0",
			},
			level: LevelError,
			err:   &errors.RoleNotFoundError{},
		},
		"uknownscm": {
			role: types.Role{
				Source:  "https://github.com/test/ansible-requirements-lint",
				Scm:     "hg",
				Version: "v1.0.0",
			},
			level: LevelError,
			err:   errors.NewUnknownScmError("hg"),
		},
	}

	for k, c := range cases {
		results := make(chan Result)
		requirements := types.Requirements{
			Roles: []types.Role{c.role},
		}
		go updatesLinter.Lint(context.Background(), &requirements, results)

		res := <-results
		if !reflect.DeepEqual(c.role, res.Role) {
			t.Errorf("%s: expecting role %+v, found %+v", k, c.role, res.Role)
		}
		if !reflect.DeepEqual(c.update, Update{}) && !reflect.DeepEqual(c.update, res.Metadata) {
			t.Errorf("%s: expecting update %+v, obtained update %+v", k, c.update, res.Metadata)
		}
		if !reflect.DeepEqual(c.level, res.Level) {
			t.Errorf("%s: expecting level %+v, found %+v", k, c.level, res.Level)
		}
		if c.err != nil && reflect.TypeOf(c.err) != reflect.TypeOf(res.Err) {
			t.Errorf("%s: expecting error of type %T, obtained type %T", k, c.err, res.Err)
		}
	}
}
