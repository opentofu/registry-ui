package moduleindex

import (
	"testing"
)

func TestFormatVariableType(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "simple string type",
			input:    "string",
			expected: "string",
		},
		{
			name:     "number type",
			input:    "number",
			expected: "number",
		},
		{
			name:     "bool type",
			input:    "bool",
			expected: "bool",
		},
		{
			name:     "nil input",
			input:    nil,
			expected: "",
		},
		{
			name:     "list of strings",
			input:    []any{"list", "string"},
			expected: "list(string)",
		},
		{
			name:     "map of strings",
			input:    []any{"map", "string"},
			expected: "map(string)",
		},
		{
			name:     "set of numbers",
			input:    []any{"set", "number"},
			expected: "set(number)",
		},
		{
			name:  "simple object",
			input: []any{"object", map[string]any{"name": "string", "age": "number"}},
			// Fields are sorted alphabetically
			expected: "object({age = number, name = string})",
		},
		{
			name: "object with optional fields",
			input: []any{
				"object",
				map[string]any{"name": "string", "nickname": "string"},
				[]any{"nickname"},
			},
			expected: "object({name = string, nickname = optional(string)})",
		},
		{
			name: "map of objects (like disks variable from issue #3442)",
			input: []any{
				"map",
				[]any{
					"object",
					map[string]any{
						"size":       "number",
						"filesystem": "string",
					},
					[]any{"filesystem"},
				},
			},
			expected: "map(object({filesystem = optional(string), size = number}))",
		},
		{
			name: "list of objects (like firewall_rules)",
			input: []any{
				"list",
				[]any{
					"object",
					map[string]any{
						"protocol":  "string",
						"range":     "string",
						"rule_type": "string",
					},
				},
			},
			expected: "list(object({protocol = string, range = string, rule_type = string}))",
		},
		{
			name: "nested map of map of strings",
			input: []any{
				"map",
				[]any{"map", "string"},
			},
			expected: "map(map(string))",
		},
		{
			name: "tuple type",
			input: []any{
				"tuple",
				[]any{"string", "number", "bool"},
			},
			expected: "tuple([string, number, bool])",
		},
		{
			name:     "list without element type",
			input:    []any{"list"},
			expected: "list",
		},
		{
			name:     "object without fields",
			input:    []any{"object"},
			expected: "object",
		},
		{
			name:     "empty array",
			input:    []any{},
			expected: "",
		},
		{
			name:     "unknown type name",
			input:    []any{"custom"},
			expected: "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatVariableType(tt.input)
			if result != tt.expected {
				t.Errorf("formatVariableType(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatObjectFields(t *testing.T) {
	tests := []struct {
		name           string
		fields         map[string]any
		optionalFields []string
		expected       string
	}{
		{
			name:           "simple fields",
			fields:         map[string]any{"name": "string", "age": "number"},
			optionalFields: nil,
			expected:       "age = number, name = string",
		},
		{
			name:           "with optional field",
			fields:         map[string]any{"name": "string", "nickname": "string"},
			optionalFields: []string{"nickname"},
			expected:       "name = string, nickname = optional(string)",
		},
		{
			name:           "all optional",
			fields:         map[string]any{"a": "string", "b": "number"},
			optionalFields: []string{"a", "b"},
			expected:       "a = optional(string), b = optional(number)",
		},
		{
			name:           "nested type",
			fields:         map[string]any{"items": []any{"list", "string"}},
			optionalFields: nil,
			expected:       "items = list(string)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatObjectFields(tt.fields, tt.optionalFields)
			if result != tt.expected {
				t.Errorf("formatObjectFields(%v, %v) = %q, want %q", tt.fields, tt.optionalFields, result, tt.expected)
			}
		})
	}
}
