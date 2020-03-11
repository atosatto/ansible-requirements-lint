package linter

import (
	"context"
	"fmt"
	"strings"

	"github.com/atosatto/ansible-requirements-lint/provider"
	"github.com/atosatto/ansible-requirements-lint/requirements"
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

// UnknownScmError is the error returned by the UpdatesLinter
// when it can't determine whether a role as updates due
// to the value of the Scm field being not valid or corresponding to
// an unknown Source Code Management system.
type UnknownScmError struct {
	scm string
}

// NewUnknownScmError creates a new UnknownScmError
func NewUnknownScmError(scm string) *UnknownScmError {
	return &UnknownScmError{scm: scm}
}

func (e *UnknownScmError) Error() string {
	return fmt.Sprintf("unknown or unsupported scm: %s", e.scm)
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

// UpdatesLinterOptions represent a set of
// options to be passed to the NewUpdatesLinter function
// to customize the UpdatesLinter configuration.
type UpdatesLinterOptions struct {
	// AnsibleGalaxyURL is the URL of the Ansible Galaxy APIs to be
	// used by the AnsibleGalaxy provider to fetch updates from Ansible Galaxy.
	AnsibleGalaxyURL string
}

// UpdatesLinter checks for updates for roles declarations.
type UpdatesLinter struct {
	cache map[string]Result

	// rolesProviders are defined as attribute of the
	// UpdatesLinter struct to allow mocking during
	// unit tests
	rolesProviders map[string]provider.RolesProvider
}

// NewUpdatesLinter returns a new UpdatesLinter.
func NewUpdatesLinter(opt UpdatesLinterOptions) *UpdatesLinter {
	providers := make(map[string]provider.RolesProvider)

	providers[git] = provider.NewGit()
	if len(opt.AnsibleGalaxyURL) == 0 {
		providers[ansibleGalaxy] = provider.NewAnsibleGalaxy(provider.DefaultAnsibleGalaxyURL)
	} else {
		providers[ansibleGalaxy] = provider.NewAnsibleGalaxy(opt.AnsibleGalaxyURL)
	}

	return &UpdatesLinter{
		// register the roles providers
		rolesProviders: providers,
	}
}

// RunWithContext checks for updates to Roles received on the roles channel.
// In case an update exists, the Metadata field of the corresponding Result
// will hold an Update struct with additional information on the new version available for the
// Role.
func (u *UpdatesLinter) RunWithContext(ctx context.Context, roles <-chan requirements.Role, results chan<- Result) {
	for r := range roles {
		var roleName string
		if len(r.Name) == 0 {
			roleName = r.Source
		} else {
			roleName = r.Name
		}

		var scm provider.RolesProvider
		h := roleHash(r)
		res, ok := u.cache[h]
		switch {
		case ok:
			// it's cached, so we just have to
			// send back the result we've found in the cache
			results <- res
			continue
		case strings.HasSuffix(r.Source, ".tar.gz"):
			fallthrough
		case strings.HasSuffix(r.Source, ".gz"):
			fallthrough
		case strings.HasSuffix(r.Source, ".tar"):
			fallthrough
		case strings.HasSuffix(r.Source, ".zip"):
			// we can't detect updates of tarballs uploaded on a custom webserver
			results <- Result{
				Role:  r,
				Level: LevelInfo,
				Msg:   fmt.Sprintf("%s: unable to detect updates for roles distributed via a custom webservers", roleName),
			}
			continue
		case r.Scm == "git":
			scm = u.rolesProviders[git]
		case r.Scm == "" && strings.HasPrefix(r.Source, "http"):
			// if it's just an URL, try with the git provider
			scm = u.rolesProviders[git]
		case r.Scm == "":
			scm = u.rolesProviders[ansibleGalaxy]
		default:
			results <- Result{
				Role:  r,
				Level: LevelError,
				Err:   NewUnknownScmError(r.Scm),
			}
			continue
		}

		versions, err := scm.VersionsForRole(ctx, r)
		if err != nil {
			results <- Result{
				Role:  r,
				Level: LevelError,
				Err:   err,
			}
			continue
		}

		latest := latestVersion(versions)
		if latest == r.Version {
			results <- Result{
				Role:     r,
				Level:    LevelInfo,
				Msg:      fmt.Sprintf("%s [%s]: the role is already at the latest version", roleName, r.Version),
				Metadata: Update{FromVersion: latest, ToVersion: latest, IsUpdate: false},
			}
			continue
		}

		versionFound := false
		for _, v := range versions {
			if v == r.Version {
				versionFound = true
				break
			}
		}
		if !versionFound {
			results <- Result{
				Role:     r,
				Level:    LevelWarning,
				Msg:      fmt.Sprintf("%s unable to find [%s] between the available versions, tag a new release or use [%s]", roleName, r.Version, latest),
				Metadata: Update{FromVersion: r.Version, ToVersion: latest, IsUpdate: false},
			}
		} else {
			results <- Result{
				Role:     r,
				Level:    LevelWarning,
				Msg:      fmt.Sprintf("%s: role not at the latest version, upgrade from %s to %s", roleName, r.Version, latest),
				Metadata: Update{FromVersion: r.Version, ToVersion: latest, IsUpdate: true},
			}
		}
	}
}
