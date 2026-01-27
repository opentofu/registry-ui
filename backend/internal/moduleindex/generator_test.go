package moduleindex

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestInferTypeFromDefault(t *testing.T) {
	tests := []struct {
		name         string
		defaultValue any
		expectedType cty.Type
	}{
		{
			name:         "string default",
			defaultValue: "ALL",
			expectedType: cty.String,
		},
		{
			name:         "empty string",
			defaultValue: "",
			expectedType: cty.String,
		},
		{
			name:         "number default (float64 from JSON)",
			defaultValue: float64(42),
			expectedType: cty.Number,
		},
		{
			name:         "zero number",
			defaultValue: float64(0),
			expectedType: cty.Number,
		},
		{
			name:         "bool true",
			defaultValue: true,
			expectedType: cty.Bool,
		},
		{
			name:         "bool false",
			defaultValue: false,
			expectedType: cty.Bool,
		},
		{
			name:         "list with elements",
			defaultValue: []any{"a", "b", "c"},
			expectedType: cty.DynamicPseudoType,
		},
		{
			name:         "empty list",
			defaultValue: []any{},
			expectedType: cty.DynamicPseudoType,
		},
		{
			name:         "map with elements",
			defaultValue: map[string]any{"key": "value"},
			expectedType: cty.DynamicPseudoType,
		},
		{
			name:         "empty map",
			defaultValue: map[string]any{},
			expectedType: cty.DynamicPseudoType,
		},
		{
			name:         "nil default",
			defaultValue: nil,
			expectedType: cty.DynamicPseudoType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferTypeFromDefault(tt.defaultValue)
			if !result.Equals(tt.expectedType) {
				t.Errorf("inferTypeFromDefault(%v) = %v, want %v",
					tt.defaultValue, result.FriendlyName(), tt.expectedType.FriendlyName())
			}
		})
	}
}
