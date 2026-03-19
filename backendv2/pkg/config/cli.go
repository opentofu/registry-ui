package config

import (
	"log"

	"github.com/urfave/cli/v3"
)

const metadataKey = "config"

// StoreToCLI stores the config in the CLI command's metadata for later retrieval.
func StoreToCLI(cmd *cli.Command, cfg *BackendConfig) {
	cmd.Metadata[metadataKey] = cfg
}

// FromCLI retrieves the BackendConfig from the CLI command's root metadata.
// This should only be called from commands that run after the Before hook has loaded config.
func FromCLI(cmd *cli.Command) *BackendConfig {
	v, ok := cmd.Root().Metadata[metadataKey]
	if !ok {
		log.Fatal("config not loaded — this command requires configuration but the Before hook did not run. Is this command in the configFree list by mistake?")
	}
	return v.(*BackendConfig)
}
