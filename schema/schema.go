package schema

import (
	"encoding/json"
	"fmt"
	"strings"
)

type (
	JSON = string
)

// Build takes a schema definition of the forms
//
//	name:type|...n
//	name:type:description|...n
//
// and builds the equivalent OpenAPI schema JSON from it
func Build(definition string) (JSON, error) {
	if definition == "" || definition[0] == '{' {
		return JSON(definition), nil
	}

	array := strings.HasPrefix(definition, "[]")
	definition = strings.TrimPrefix(definition, "[]")

	schema := map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}

	properties := schema["properties"].(map[string]any)

	if array {
		schema = map[string]any{
			"type":  "array",
			"items": schema,
		}
	}

	for _, field := range strings.Split(definition, "|") {
		field = strings.TrimSpace(field)
		if field == "" {
			return "", fmt.Errorf("missing field definition. expected 'name:type|...' or 'name:type:description|...'. got empty string in %q", definition)
		}

		attributes := strings.Split(field, ":")

		if len(attributes) != 2 && len(attributes) != 3 {
			return "", fmt.Errorf("invalid definition format: '%s'.expected 'name:type' or 'name:type:description'", field)
		}

		name := strings.TrimSpace(attributes[0])
		datatype := strings.TrimSpace(attributes[1])
		description := ""

		if len(attributes) == 3 {
			description = strings.TrimSpace(attributes[2])
		}

		if name == "" || datatype == "" {
			return "", fmt.Errorf("name and type cannot be empty in field definition: '%s'", field)
		}

		property := map[string]any{
			"type":        datatype,
			"description": description,
		}

		properties[name] = property
	}

	data, err := json.Marshal(schema)
	if err != nil {
		return "", fmt.Errorf("error marshalling schema to json. %w", err)
	}

	return JSON(data), nil
}
