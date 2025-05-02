package parser

import (
	"encoding/json"
	"fmt"

	"github.com/papidb/drift-detector/internal/types"
)

// ParseTerraformStateFile parses a Terraform state file and extracts only the attributes of each instance.
func ParseTerraformStateFile(data []byte) ([]types.Resource, error) {
	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	results := make([]types.Resource, 0) // Initialize as empty slice
	resources, ok := state["resources"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no resources found in state file")
	}

	for _, r := range resources {
		resourceMap, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		resourceType, _ := resourceMap["type"].(string)

		instances, ok := resourceMap["instances"].([]interface{})
		if !ok {
			continue
		}

		for _, inst := range instances {
			instanceMap, ok := inst.(map[string]interface{})
			if !ok {
				continue
			}

			attributes, ok := instanceMap["attributes"].(map[string]interface{})
			if !ok {
				continue
			}

			// Convert vpc_security_group_ids to []string
			var securityGroups []string
			if sgIds, ok := attributes["vpc_security_group_ids"].([]interface{}); ok {
				for _, sgId := range sgIds {
					if sgIdStr, ok := sgId.(string); ok {
						securityGroups = append(securityGroups, sgIdStr)
					}
				}
			}

			// Convert tags to map[string]string
			var tags map[string]string
			if tagsRaw, ok := attributes["tags"].(map[string]interface{}); ok {
				tags = make(map[string]string)
				for k, v := range tagsRaw {
					if vStr, ok := v.(string); ok {
						tags[k] = vStr
					}
				}
			}

			normalized := map[string]interface{}{
				"instance_id":       attributes["id"],
				"instance_type":     attributes["instance_type"],
				"ami":               attributes["ami"],
				"key_name":          attributes["key_name"],
				"subnet_id":         attributes["subnet_id"],
				"availability_zone": attributes["availability_zone"],
				"state":             attributes["instance_state"],
				"tags":              tags,
				"private_ip":        attributes["private_ip"],
				"public_ip":         attributes["public_ip"],
				"security_groups":   securityGroups,
			}

			results = append(results, types.NewResource(
				fmt.Sprintf("%v", attributes["id"]),
				types.ResourceType(resourceType),
				normalized,
			))
		}
	}

	return results, nil
}
