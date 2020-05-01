package parser

import (
	"io/ioutil"

	"github.com/atosatto/ansible-requirements-lint/pkg/types"

	"gopkg.in/yaml.v3"
)

// UnmarshalFromFile parses the Requirements defined in the
// file stored at the given path.
func UnmarshalFromFile(path string) (*types.Requirements, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Unmarshal(content)
}

// Unmarshal parses the Ansible Requirements defined in data.
func Unmarshal(data []byte) (*types.Requirements, error) {
	var root yaml.Node
	err := yaml.Unmarshal(data, &root)
	if err != nil {
		return nil, err
	}

	var requirements = types.Requirements{}
	if len(root.Content) == 0 {
		// the file is empty
		return &requirements, nil
	}

	var node = root.Content[0]
	var childrens = make([]*yaml.Node, len(node.Content))
	copy(childrens, node.Content)

	switch {
	case node.Kind == yaml.SequenceNode:
		// we have a sequence of roles
		roles, err := parseRolesFromNodesList(childrens)
		if err != nil {
			return nil, err
		}
		requirements.Roles = roles
	case node.Kind == yaml.MappingNode:
		// we have a dictionary with either a roles or a collections list
		for {
			if len(childrens) < 2 {
				break
			}

			// Note that
			// - node.Content[0] will be the key of the dictionary
			// - node.Content[1] will contain the list of roles or collections
			var k, v = childrens[0], childrens[1]
			childrens = childrens[2:]
			switch {
			case k.Kind == yaml.ScalarNode && k.Value == "roles":
				// we are parsing a requirements file using the new
				// dictionary based syntax introduced with Collections
				fallthrough
			case k.Kind == yaml.ScalarNode && k.Value == "dependencies":
				// we are parsing roles dependencies contained in the meta/main.yml
				// manifest of an Ansible role
				roles, err := parseRolesFromNodesList(v.Content)
				if err != nil {
					return nil, err
				}
				requirements.Roles = roles
			case k.Kind == yaml.ScalarNode && k.Value == "collections":
				// collections are not supported yet
			case k.Kind == yaml.ScalarNode:
				// when parsing dependencies in the meta/main.yml format
				// we might encounter here the meta specific fields
			default:
				return nil, NewUnexpectedNodeKindError(k.Line, k.Kind)
			}
		}
	default:
		// mh, something wrong is happening here
		return nil, NewUnexpectedNodeKindError(node.Line, node.Kind)
	}

	return &requirements, nil
}

func parseRolesFromNodesList(nodes []*yaml.Node) ([]types.Role, error) {
	var res []types.Role

	for _, n := range nodes {
		var role = types.Role{}

		switch {
		case n.Kind == yaml.ScalarNode:
			// entries in which the name of the role
			// is provided directly (e.g. meta/main.yml dependencies)
			role.Name = n.Value
		case n.Kind == yaml.MappingNode:
			// this is a standard dict-based role definition
			var childrens = make([]*yaml.Node, len(n.Content))
			copy(childrens, n.Content)

			for {
				if len(childrens) < 2 {
					break
				}

				var k, v = childrens[0], childrens[1]
				childrens = childrens[2:]
				switch {
				case k.Kind == yaml.ScalarNode && k.Value == "src":
					role.Source = v.Value
				case k.Kind == yaml.ScalarNode && k.Value == "scm":
					role.Scm = v.Value
				case k.Kind == yaml.ScalarNode && k.Value == "version":
					role.Version = v.Value
				case k.Kind == yaml.ScalarNode && (k.Value == "name" || k.Value == "role"):
					role.Name = v.Value
				case k.Kind == yaml.ScalarNode && k.Value == "include":
					role.Include = v.Value
				case k.Kind == yaml.ScalarNode:
					// when parsing dependencies in the meta/main.yml format
					// we might encounter some variables names, let's ignore them
				default:
					return nil, NewUnexpectedNodeKindError(k.Line, k.Kind)
				}
			}
		default:
			return nil, NewUnexpectedNodeKindError(n.Line, n.Kind)
		}

		res = append(res, role)
	}

	return res, nil
}
