package provider

import (
	"context"
	"fmt"

	"github.com/atosatto/ansible-requirements-lint/requirements"
	"gopkg.in/src-d/go-billy.v4/memfs"
	gogit "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

// Git fetches Ansible Roles information
// from remote Git repositories.
type Git struct{}

// NewGit creates a new Git provider.
func NewGit() Git {
	return Git{}
}

// VersionsForRole returns the list of versions available on the upstream Git repository for Role r.
func (g Git) VersionsForRole(ctx context.Context, r requirements.Role) ([]string, error) {
	// Use an in-memory filesystem for the repo and git objects
	fs := memfs.New()
	storer := memory.NewStorage()

	// Clone the repository
	repo, err := gogit.CloneContext(ctx, storer, fs, &gogit.CloneOptions{URL: r.Source})
	if err != nil {
		return nil, fmt.Errorf("cloning %s: %v", r.Source, err)
	}

	// Fetch tags
	tags, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("listing tags for %s: %v", r.Source, err)
	}

	// Look for the most recent tag by commit date
	var versions []string
	err = tags.ForEach(func(tagRef *plumbing.Reference) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			versions = append(versions, tagRef.Name().Short())
			return nil
		}
	})
	if err != nil {
		return nil, err
	}

	return versions, nil
}
