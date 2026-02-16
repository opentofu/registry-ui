package scraper

import (
	"slices"
	"strings"

	"github.com/opentofu/registry-ui/pkg/provider/storage"
)

// getDocCategory determines the category of a document based on its file path
// Returns the category name (resources, datasources, functions, guides, ephemeral, index) or empty string if no match
func getDocCategory(filePath string) string {
	if strings.Contains(filePath, "index") {
		return "index"
	}
	parts := strings.Split(filePath, "/")

	for _, config := range storage.DocTypes {
		// First check if it starts with the target path (regular docs)
		if strings.HasPrefix(filePath, config.TargetPath+"/") {
			return config.TargetPath
		}

		// Then check for source directories in the path segments (handles both regular and CDKTF docs)
		for _, sourceDir := range config.SourceDirs {
			if slices.Contains(parts, sourceDir) {
				return config.TargetPath
			}
		}
	}

	return ""
}
