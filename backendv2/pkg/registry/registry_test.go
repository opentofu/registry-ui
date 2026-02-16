package registry

import (
	"testing"
)

func TestParseFilter(t *testing.T) {
	tests := []struct {
		filter   string
		expected []string
	}{
		{"", nil},
		{"*", nil},
		{"hashicorp", []string{"hashicorp"}},
		{"hashicorp/aws", []string{"hashicorp", "aws"}},
		{"terraform-aws-modules/vpc/aws", []string{"terraform-aws-modules", "vpc", "aws"}},
	}

	for _, tt := range tests {
		t.Run(tt.filter, func(t *testing.T) {
			result := parseFilter(tt.filter)
			if len(result) != len(tt.expected) {
				t.Errorf("parseFilter(%q) = %v, want %v", tt.filter, result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("parseFilter(%q)[%d] = %v, want %v", tt.filter, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		pattern string
		value   string
		want    bool
	}{
		{"*", "anything", true},
		{"hashicorp", "hashicorp", true},
		{"hashicorp", "other", false},
		{"hash*", "hashicorp", true},
		{"*corp", "hashicorp", true},
		{"*shi*", "hashicorp", true},
		{"hash*", "other", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.value, func(t *testing.T) {
			if got := matchPattern(tt.pattern, tt.value); got != tt.want {
				t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.pattern, tt.value, got, tt.want)
			}
		})
	}
}

func TestMatchesFilter(t *testing.T) {
	tests := []struct {
		name        string
		parts       []string
		filterParts []string
		want        bool
	}{
		{
			name:        "empty filter matches all",
			parts:       []string{"hashicorp", "aws"},
			filterParts: nil,
			want:        true,
		},
		{
			name:        "exact namespace match",
			parts:       []string{"hashicorp", "aws"},
			filterParts: []string{"hashicorp"},
			want:        true,
		},
		{
			name:        "exact full match",
			parts:       []string{"hashicorp", "aws"},
			filterParts: []string{"hashicorp", "aws"},
			want:        true,
		},
		{
			name:        "wildcard namespace",
			parts:       []string{"hashicorp", "aws"},
			filterParts: []string{"*", "aws"},
			want:        true,
		},
		{
			name:        "wildcard name",
			parts:       []string{"hashicorp", "aws"},
			filterParts: []string{"hashicorp", "*"},
			want:        true,
		},
		{
			name:        "no match",
			parts:       []string{"hashicorp", "aws"},
			filterParts: []string{"other", "gcp"},
			want:        false,
		},
		{
			name:        "filter longer than parts",
			parts:       []string{"hashicorp"},
			filterParts: []string{"hashicorp", "aws", "extra"},
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesFilter(tt.parts, tt.filterParts); got != tt.want {
				t.Errorf("matchesFilter(%v, %v) = %v, want %v", tt.parts, tt.filterParts, got, tt.want)
			}
		})
	}
}

func TestMatchesModuleFilter(t *testing.T) {
	tests := []struct {
		name        string
		parts       []string
		filterParts []string
		want        bool
	}{
		{
			name:        "empty filter matches all",
			parts:       []string{"terraform-aws-modules", "vpc", "aws"},
			filterParts: nil,
			want:        true,
		},
		{
			name:        "namespace only",
			parts:       []string{"terraform-aws-modules", "vpc", "aws"},
			filterParts: []string{"terraform-aws-modules"},
			want:        true,
		},
		{
			name:        "namespace and name",
			parts:       []string{"terraform-aws-modules", "vpc", "aws"},
			filterParts: []string{"terraform-aws-modules", "vpc"},
			want:        true,
		},
		{
			name:        "full match",
			parts:       []string{"terraform-aws-modules", "vpc", "aws"},
			filterParts: []string{"terraform-aws-modules", "vpc", "aws"},
			want:        true,
		},
		{
			name:        "wildcard in namespace",
			parts:       []string{"terraform-aws-modules", "vpc", "aws"},
			filterParts: []string{"terraform-*", "vpc", "aws"},
			want:        true,
		},
		{
			name:        "wildcard in name",
			parts:       []string{"terraform-aws-modules", "vpc", "aws"},
			filterParts: []string{"terraform-aws-modules", "*", "aws"},
			want:        true,
		},
		{
			name:        "wildcard in target",
			parts:       []string{"terraform-aws-modules", "vpc", "aws"},
			filterParts: []string{"terraform-aws-modules", "vpc", "*"},
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesModuleFilter(tt.parts, tt.filterParts); got != tt.want {
				t.Errorf("matchesModuleFilter(%v, %v) = %v, want %v", tt.parts, tt.filterParts, got, tt.want)
			}
		})
	}
}
