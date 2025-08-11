package docscraper

import "testing"

func TestGetDocCategory(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "resources path",
			filePath: "resources/aws_instance",
			expected: "resources",
		},
		{
			name:     "resources nested path",
			filePath: "resources/aws/instance",
			expected: "resources",
		},
		{
			name:     "datasources path",
			filePath: "datasources/aws_vpc",
			expected: "datasources",
		},
		{
			name:     "datasources nested path",
			filePath: "datasources/aws/vpc",
			expected: "datasources",
		},
		{
			name:     "functions path",
			filePath: "functions/file",
			expected: "functions",
		},
		{
			name:     "functions nested path",
			filePath: "functions/crypto/file",
			expected: "functions",
		},
		{
			name:     "guides path",
			filePath: "guides/getting-started",
			expected: "guides",
		},
		{
			name:     "guides nested path",
			filePath: "guides/advanced/configuration",
			expected: "guides",
		},
		{
			name:     "ephemeral path",
			filePath: "ephemeral/resource",
			expected: "ephemeral",
		},
		{
			name:     "ephemeral nested path",
			filePath: "ephemeral/aws/resource",
			expected: "ephemeral",
		},
		{
			name:     "index file",
			filePath: "index",
			expected: "index",
		},
		{
			name:     "index.md file",
			filePath: "index.md",
			expected: "index",
		},
		{
			name:     "nested index file",
			filePath: "cdktf/python/index",
			expected: "index",
		},
		{
			name:     "index in filename",
			filePath: "resources/index_resource",
			expected: "index",
		},
		{
			name:     "CDKTF resources",
			filePath: "cdktf/python/resources/aws_instance",
			expected: "",
		},
		{
			name:     "CDKTF datasources",
			filePath: "cdktf/typescript/datasources/aws_vpc",
			expected: "",
		},
		{
			name:     "unknown path",
			filePath: "unknown/something",
			expected: "",
		},
		{
			name:     "empty path",
			filePath: "",
			expected: "",
		},
		{
			name:     "just filename",
			filePath: "readme.md",
			expected: "",
		},
		{
			name:     "partial match",
			filePath: "resource/test",
			expected: "",
		},
		{
			name:     "path ending with slash",
			filePath: "resources/",
			expected: "resources",
		},
		{
			name:     "exact match without slash",
			filePath: "resources",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := getDocCategory(tt.filePath)
			if actual != tt.expected {
				t.Errorf("getDocCategory(%q) = %q; want %q", tt.filePath, actual, tt.expected)
			}
		})
	}
}
