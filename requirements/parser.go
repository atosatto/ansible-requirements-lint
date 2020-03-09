package requirements

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// UnmarshalFromFile parses the Requirements defined in the
// file stored at the given path.
func UnmarshalFromFile(path string) (*Requirements, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Unmarshal(content)
}

// Unmarshal parses the Ansible Requirements defined in data.
func Unmarshal(data []byte) (*Requirements, error) {
	var root yaml.Node
	err := yaml.Unmarshal(data, &root)
	if err != nil {
		return nil, err
	}

	var requirements = Requirements{node: &root}
	if len(root.Content) == 0 {
		// the file is empty
		return &requirements, nil
	}

	var node = root.Content[0]
	var childrens = node.Content
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

		// In case of maps,we should have at least a key and a value,
		// if not we break the loop
		if len(childrens)/2 <= 0 {
			break
		}

		for {
			// Note that
			// - node.Content[0] will be the key of the dictionary
			// - node.Content[1] will contain the list of roles or collections
			var k, v = childrens[0], childrens[1]
			switch {
			case k.Kind == yaml.ScalarNode && k.Value == "roles":
				roles, err := parseRolesFromNodesList(v.Content)
				if err != nil {
					return nil, err
				}
				requirements.Roles = roles
			case k.Kind == yaml.ScalarNode && k.Value == "collections":
				// collections are not supported yet
			case k.Kind == yaml.ScalarNode && k.Value == "collections":
				return nil, fmt.Errorf("unknown dictionary key at line %d: %s", k.Line, k.Value)
			default:
				return nil, fmt.Errorf("unexpected node type at line %d: %v", k.Line, k.Kind)
			}

			// if there are no more nodes to parse
			if len(childrens) <= 2 {
				break
			}

			// parse the next key/value pair
			childrens = childrens[2:]
		}
	default:
		// mh, something wrong is happening here
		return nil, fmt.Errorf("unexpected node type at line %d: %v", node.Line, node.Kind)
	}

	return &requirements, nil
}

func parseRolesFromNodesList(nodes []*yaml.Node) ([]Role, error) {
	var res []Role

	for _, n := range nodes {
		var role Role
		err := n.Decode(&role)
		if err != nil {
			return nil, fmt.Errorf("decoding line %d as role: %w", n.Line, err)
		}
		role.node = n
		res = append(res, role)
	}

	return res, nil
}
