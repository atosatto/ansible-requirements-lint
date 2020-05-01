package linter

import (
	"context"
	"fmt"
	"strings"

	"github.com/atosatto/ansible-requirements-lint/pkg/errors"
	"github.com/atosatto/ansible-requirements-lint/pkg/provider"
	"github.com/atosatto/ansible-requirements-lint/pkg/types"
)

// Update represents an Ansible role update.
type Update struct {
	// FromVersion represents the current version of the role.
	FromVersion string

	// ToVersion is a more recent version for a role.
	ToVersion string

	// IsUpdate is true if an update has been found for the role.
	IsUpdate bool
}

const (
	// git is the value of the key used in the rolesProviders
	// map of the UpdatesLinter to refer to the Git implementation
	// of the RolesProvider interface.
	git = "git"

	// ansibleGalaxy is the value of they used in the rolesProviders
	// map of the UpdatesLinter to refer to the AnsibleGalaxy implementation
	// of the RolesProvider interface.
	ansibleGalaxy = "galaxy"
)

// UpdatesLinter checks for updates for roles declarations.
type UpdatesLinter struct {
	cache map[string]Result

	// rolesProviders are defined as attribute of the
	// UpdatesLinter struct to allow mocking during
	// unit tests
	rolesProviders map[string]provider.RolesProvider
}

// NewUpdatesLinter returns a new UpdatesLinter.
func NewUpdatesLinter() *UpdatesLinter {
	providers := make(map[string]provider.RolesProvider)
	providers[git] = provider.NewGit()
	providers[ansibleGalaxy] = provider.NewAnsibleGalaxy(provider.DefaultAnsibleGalaxyURL)

	return &UpdatesLinter{
		// register the roles providers
		rolesProviders: providers,
	}
}

// WithAnsibleGalaxyURL configures the UpdatesLinter
// to use the given URL for Ansible Galaxy instead of
// the default one.
func (u *UpdatesLinter) WithAnsibleGalaxyURL(url string) {
	u.rolesProviders[ansibleGalaxy] = provider.NewAnsibleGalaxy(url)
}

// Lint checks for updates to the Roles defined in the given Requirements.
// Linter Results will be sent on the output channel.
// In case an update exists for a given Role, the corresponding
// Result will have the Metadata field set to an Update holding additional information
// on the new version available for the role.
func (u *UpdatesLinter) Lint(ctx context.Context, requirements *types.Requirements, output chan<- Result) error {
	// make sure to close the results chan on exit
	defer close(output)
	for _, role := range requirements.Roles {
		select {
		case <-ctx.Done():
			return nil
		default:
			// provider to be used to fetch updates to the role
			var scm provider.RolesProvider

			// otherwise, we check if the role has any update
			h := roleHash(role)
			res, ok := u.cache[h]
			switch {
			case ok:
				// it's cached, so we just have to
				// send back the result we've found in the cache
				output <- res
				continue
			case strings.HasSuffix(role.Source, ".tar.gz"):
				fallthrough
			case strings.HasSuffix(role.Source, ".gz"):
				fallthrough
			case strings.HasSuffix(role.Source, ".tar"):
				fallthrough
			case strings.HasSuffix(role.Source, ".zip"):
				// we can't detect updates of tarballs uploaded on a custom webserver
				output <- Result{
					Role:  role,
					Level: LevelInfo,
					Err:   fmt.Errorf("unable to detect updates for roles distributed via custom webservers"),
				}
				continue
			case role.Scm == "git":
				scm = u.rolesProviders[git]
			case role.Scm == "" && strings.HasPrefix(role.Source, "http"):
				// if it's just an URL, try with the git provider
				scm = u.rolesProviders[git]
			case role.Scm == "":
				scm = u.rolesProviders[ansibleGalaxy]
			default:
				output <- Result{
					Role:  role,
					Level: LevelError,
					Err:   errors.NewUnknownScmError(role.Scm),
				}
				continue
			}

			// fetch the versions available for the role
			versions, err := scm.VersionsForRole(ctx, role)
			if err != nil {
				output <- Result{
					Role:  role,
					Level: LevelError,
					Err:   err,
				}
				continue
			}

			// check if the current version of the role is the latest
			latest := latestVersion(versions)
			if latest == role.Version {
				output <- Result{
					Role:     role,
					Level:    LevelInfo,
					Metadata: Update{FromVersion: latest, ToVersion: latest, IsUpdate: false},
				}
				continue
			}

			// check if there are new versions for the role
			versionFound := false
			for _, v := range versions {
				if v == role.Version {
					versionFound = true
					break
				}
			}
			if !versionFound {
				output <- Result{
					Role:     role,
					Level:    LevelWarning,
					Err:      errors.NewRoleVersionNotFoundError(role, versions),
					Metadata: Update{FromVersion: role.Version, ToVersion: latest, IsUpdate: false},
				}
			} else {
				output <- Result{
					Role:     role,
					Level:    LevelWarning,
					Metadata: Update{FromVersion: role.Version, ToVersion: latest, IsUpdate: true},
				}
			}
		}
	}
	return nil
}
