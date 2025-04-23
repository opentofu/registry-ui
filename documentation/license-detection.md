# License Detection

This document describes how the OpenTofu Registry detects licenses in repositories and determines if they are redistributable.

## Detection Process

License detection happens in two places:

1. When indexing modules
2. When indexing providers

Both use the same license detection code in `backend/internal/license/license_detector.go`, which leverages the [go-license-detector](https://github.com/go-enry/go-license-detector) library.

## License File Inclusion/Exclusion

### License Files to Include

We should specifically look for these license files:

- `LICENSE` (and variations like `License`, `LICENCE`, etc.)
- `LICENSE.md`, `LICENSE.txt`, `LICENSE.rst`
- `COPYING` (and variations)
- `COPYING.md`, `COPYING.txt`
- `COPYRIGHT`
- `UNLICENSE`
- `MIT-LICENSE`
- `APACHE-LICENSE-2.0`
- `GPL-LICENSE`

### License Files to Ignore

We should explicitly ignore these files:

- `THIRD_PARTY_LICENSES.txt`
- `THIRD_PARTY_LICENSE`
- `3RD_PARTY_LICENSES`
- `PATENTS`
- `NOTICE` (contains additional information, not the main license)
- Any license files within `vendor/`, `node_modules/`, or other dependency directories
- License files in test fixtures
- Example licenses in `examples/` directories

## License Detection Hierarchy

When detecting licenses, we follow these rules:

1. **Documentation Directory Priority**:
   - License files in documentation directories (`website/docs/`, `docs/`) take precedence over license files in the root directory
   - This allows projects to specify different licensing terms for their documentation

2. **License File Evaluation**:
   - Each license file is evaluated with a confidence score for various license types
   - Only license types with confidence scores above our threshold (default: 0.85) are considered
   - For a repository to be redistributable, all detected license types (that meet the threshold) must be in our approved license list

3. **No Valid License**:
   - If no license files meet our confidence threshold, we do not scrape the documentation
   - We display a "451 Unavailable for Legal Reasons" message

## Confidence Scoring and Evaluation

1. **Confidence Scoring**:
   - Confidence scores are per license type, not per file
   - Each potential license type receives a confidence score between 0 and 1
   - Only licenses with confidence above our threshold (default: 0.85) are considered

2. **High Confidence Override**:
   - A license type with extremely high confidence (default: 0.98) may override other detected licenses
   - This handles cases where a repository contains multiple license files but one is clearly dominant

3. **Sorting License Files**:
   - When multiple license files are found, we sort them consistently to ensure deterministic results
   - Sort criteria:
     - Documentation directory files first
     - Files with shorter paths (closer to root) next
     - Alphabetical order for ties

## Repository Redistributability

A repository is redistributable if:

1. At least one license file is found with confidence above our threshold
2. ALL license types detected (that meet the threshold) are in our compatible license list
3. If multiple license files exist, we respect the directory hierarchy (docs/ overrides root)

## Reporting and Handling

When licenses are detected, we:

1. Check each license type against our allowlist of compatible licenses (`licenses.json`)
2. Mark the module/provider as redistributable only if all detected license types are compatible
3. Include links to the license files in the repository when available
4. For non-redistributable content, display a "451 Unavailable for Legal Reasons" message
