# Documentation Sources

This document describes how the OpenTofu Registry scrapes and normalizes documentation from provider and module repositories.

## Provider Documentation Sources

The OpenTofu Registry supports multiple formats for provider documentation, accommodating both the latest format and legacy formats used by providers.

### Directory Structure

The registry looks for documentation in two possible locations, in order of preference:

1. `website/docs/` (legacy format)
2. `docs/` (modern format)

We prefer `website/docs` over just `docs` because some providers (such as AWS) use the `docs` directory for internal documentation rather than user-facing documentation.

### Documentation Format

Provider documentation is organized in the following structure:

#### Latest Format

```sh
docs/
├── index.md                      # Provider overview documentation
├── guides/                       # Usage guides
│   └── <guide>.md
├── resources/                    # Resource documentation
│   └── <resource>.md
├── data-sources/                 # Data source documentation
│   └── <data-source>.md
└── functions/                    # Functions documentation
    └── <function>.md
```

#### Legacy Format

```sh
website/docs/
├── index.html.markdown           # Provider overview (can be .html.md as well)
├── guides/                       # Usage guides
│   └── <guide>.html.markdown
├── r/                            # Resource documentation
│   └── <resource>.html.markdown
├── d/                            # Data source documentation
│   └── <data-source>.html.markdown
└── f/                            # Functions documentation
    └── <function>.html.markdown
```

### CDKTF Documentation

Both formats also support CDKTF (Cloud Development Kit for Terraform) documentation in the following structures:

#### Latest Format

```sh
docs/cdktf/[python|typescript]/
├── resources/
│   └── <resource>.md
├── data-sources/
│   └── <data-source>.md
└── functions/
    └── <function>.md
```

#### Legacy Format

```sh
website/docs/cdktf/[python|typescript]/
├── r/
│   └── <resource>.html.markdown
├── d/
│   └── <data-source>.html.markdown
└── f/
    └── <function>.html.markdown
```

### Normalization Process

When scraping documentation, we follow these steps:

1. **License Check**: Documentation is only processed if the repository has compatible licenses
2. **Directory Location**: Check for documentation in `website/docs/` first, then `docs/`
3. **Root Document**: Extract the provider overview from an `index.*` file in the root of the docs directory
4. **Document Types**: Scrape the following document types:
   - `r/` or `resources/` for resources
   - `d/` or `data-sources/` for data sources
   - `f/` or `functions/` for functions
   - `guides/` for guides
5. **CDKTF Languages**: For each supported CDKTF language (Python, TypeScript), scrape language-specific documentation
6. **File Extensions**: Support multiple file extensions:
   - `.html.md`
   - `.html.markdown`
   - `.md.html`
   - `.markdown.html`
   - `.md`
   - `.markdown`
7. **Frontmatter Extraction**: Parse YAML frontmatter to extract metadata (title, subcategory, description)
8. **Edit Links**: Generate links to the original source files in the repository

### File Size Limits

To prevent denial-of-service attacks through memory exhaustion, we limit the size of documentation files to 1MB. For files larger than this limit, we generate an error message with a link to view the file in the repository directly.

### Storage Format

After scraping, documentation is stored in a normalized structure that separates:

1. Provider overview
2. Resources
3. Data sources
4. Functions
5. Guides
6. CDKTF language-specific versions of the above

This normalized structure makes it easier to present the documentation in the web interface while preserving the original organization and metadata from the source repository.

## Module Documentation and Schema Extraction

Unlike providers, which primarily share documentation, modules share both documentation and their schema - including variables, outputs, resources, and submodules. To extract this information, we need to parse the Terraform/OpenTofu configuration files.

### The Challenge with terraform-config-inspect

Traditionally, the `terraform-config-inspect` package has been used to extract information from Terraform modules. However, this package doesn't fully support OpenTofu's specific features and syntax extensions. Additionally, it doesn't provide the complete schema information we need for the registry.

### Our Solution: OpenTofu Metadata Command

Instead of using third-party parsing libraries, we utilize OpenTofu's built-in capabilities to extract module information. We use a custom branch of the OpenTofu binary that includes a special `metadata dump` command designed specifically for the registry.

The process works as follows:

1. We maintain a custom branch of OpenTofu with the `metadata dump` command
2. This command analyzes module configuration files and outputs structured JSON data
3. Our registry backend invokes this command and processes the output

### Custom OpenTofu Binary

The custom OpenTofu binary is packaged with the registry backend. When extracting module information, the backend:

1. Checks if the OpenTofu binary is available
2. If not, downloads it using the `tofudl` package
3. Runs the `metadata dump -json` command against the module directory
4. Parses the JSON output into a structured schema

### Schema Information Extracted

The module extraction process collects the following information:

- **Provider Requirements**: What providers the module requires and their version constraints
- **Variables**: Input variables with their types, default values, and descriptions
- **Outputs**: Module outputs with descriptions and references
- **Resources**: All resources defined in the module
- **Module Calls**: References to other modules, including source and version constraints
- **Submodules**: The schema of each submodule, recursively extracted

### Integration with the Registry

The extracted module schema is:

1. Stored in the registry's database
2. Used to generate API responses for the frontend
3. Indexed for search functionality
4. Displayed in the web interface to help users understand module usage

### Future Plans

Eventually, we hope to merge the `metadata dump` command into the main OpenTofu codebase. This would allow us to use the standard OpenTofu binary instead of maintaining a custom branch. Until then, we'll continue using our custom implementation to ensure the registry has the information it needs.

## Future Improvements

- Better normalization of links between documentation pages
- Improved module schema extraction with better error handling and validation