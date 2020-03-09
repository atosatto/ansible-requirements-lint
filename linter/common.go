package linter

import (
	"fmt"
	"hash/fnv"
	"sort"

	"github.com/atosatto/ansible-requirements-lint/requirements"
	version "github.com/hashicorp/go-version"
)

// roleHash returns an hash of the given Role to be used
// as key for the UpdatesLinter cache.
func roleHash(r requirements.Role) string {
	h := fnv.New32a()
	h.Write([]byte(r.Source))
	return fmt.Sprintf("%x", h.Sum32())
}

// latestVersion returns the latest semanting version
// in the provided list of version tags.
func latestVersion(tags []string) string {
	versions := make([]*version.Version, len(tags))
	for i, t := range tags {
		v, _ := version.NewVersion(t)
		versions[i] = v
	}

	sort.Sort(version.Collection(versions))
	last := versions[len(versions)-1]
	return last.Original()
}
