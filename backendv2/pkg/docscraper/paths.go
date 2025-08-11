package docscraper

import "strings"

// getDocCategory determines the category of a document based on its file path
// Returns the category name (resources, datasources, functions, guides, ephemeral, index) or empty string if no match
func getDocCategory(filePath string) string {
	if strings.Contains(filePath, "index") {
		return "index"
	}
	parts := strings.Split(filePath, "/")

	for _, config := range docTypes {
		// First check if it starts with the target path (regular docs)
		if strings.HasPrefix(filePath, config.TargetPath+"/") {
			return config.TargetPath
		}

		// Then check for source directories in the path segments (handles both regular and CDKTF docs)
		for _, sourceDir := range config.SourceDirs {
			for _, part := range parts {
				if part == sourceDir {
					return config.TargetPath
				}
			}
		}
	}

	return ""
}

// extractLanguage determines the language from a file path
// Returns the language code (typescript, python, java, csharp, go) for CDKTF docs or "default" for regular docs
func extractLanguage(filePath string) string {
	// Check if this is a CDKTF document path
	if strings.Contains(filePath, "cdktf/") {
		parts := strings.Split(filePath, "/")
		for i, part := range parts {
			if part == "cdktf" && i+1 < len(parts) {
				return parts[i+1] // Return the language (python, typescript, etc.)
			}
		}
	}
	return "default" // Regular terraform docs
}
