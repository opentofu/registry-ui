package moduleschema

// This command generates a usable tofu binary and tofudl cache directory. This is a temporary measure until the
// tofu metadata dump command makes it into a released version.

//go:generate go run github.com/opentofu/registry-ui/internal/moduleindex/moduleschema/tools/build-tofu-binary
