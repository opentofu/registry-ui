// Package storage handles database operations for provider metadata and documents
package storage

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"

	"github.com/opentofu/registry-ui/pkg/license"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

// Queryable is an interface that both pgx.Tx and *pgxpool.Pool implement for write operations
type Queryable interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// StoreProviderVersion stores provider version information in the database
func StoreProviderVersion(ctx context.Context, tx pgx.Tx, namespace, name, version string, docCount, licenseCount int, tagCreatedAt *time.Time, licenseAccepted bool, scrapeStatus, skipReason, errorMessage string) error {
	query := `
		INSERT INTO provider_versions (provider_namespace, provider_name, version, tag_created_at, license_accepted, scrape_status, skip_reason, error_message, last_attempt_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (provider_namespace, provider_name, version)
		DO UPDATE SET
			tag_created_at = EXCLUDED.tag_created_at,
			license_accepted = EXCLUDED.license_accepted,
			scrape_status = EXCLUDED.scrape_status,
			skip_reason = EXCLUDED.skip_reason,
			error_message = EXCLUDED.error_message,
			last_attempt_at = NOW(),
			updated_at = NOW()`

	_, err := tx.Exec(ctx, query, namespace, name, version, tagCreatedAt, licenseAccepted, scrapeStatus, skipReason, errorMessage)
	if err != nil {
		return fmt.Errorf("failed to store provider version: %w", err)
	}

	return nil
}

// StoreRepository stores repository information in the database
// Accepts Queryable interface so it can be called with either pgx.Tx or *pgxpool.Pool
func StoreRepository(ctx context.Context, db Queryable, organisation, name string) error {
	query := `
		INSERT INTO repositories (organisation, name, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (organisation, name)
		DO UPDATE SET
			updated_at = NOW()`

	_, err := db.Exec(ctx, query, organisation, name)
	if err != nil {
		return fmt.Errorf("failed to store repository: %w", err)
	}

	return nil
}

// StoreProvider stores provider information in the database
// Accepts Queryable interface so it can be called with either pgx.Tx or *pgxpool.Pool
func StoreProvider(ctx context.Context, db Queryable, namespace, name, repoOrganisation, repoName string, warnings []string) error {
	query := `
		INSERT INTO providers (namespace, name, warnings, repo_organisation, repo_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (namespace, name)
		DO UPDATE SET
			warnings = EXCLUDED.warnings,
			repo_organisation = EXCLUDED.repo_organisation,
			repo_name = EXCLUDED.repo_name,
			updated_at = NOW()`

	_, err := db.Exec(ctx, query, namespace, name, warnings, repoOrganisation, repoName)
	if err != nil {
		return fmt.Errorf("failed to store provider: %w", err)
	}

	return nil
}

// ProviderInfo represents provider information from the database
type ProviderInfo struct {
	Namespace        string
	Name             string
	RepoOrganisation string
	RepoName         string
}

// GetProvider retrieves a provider from the database by namespace and name
func GetProvider(ctx context.Context, pool *pgxpool.Pool, namespace, name string) (*ProviderInfo, error) {
	query := `
		SELECT namespace, name, repo_organisation, repo_name 
		FROM providers 
		WHERE namespace = $1 AND name = $2`

	var provider ProviderInfo
	err := pool.QueryRow(ctx, query, namespace, name).Scan(
		&provider.Namespace,
		&provider.Name,
		&provider.RepoOrganisation,
		&provider.RepoName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider %s/%s: %w", namespace, name, err)
	}

	return &provider, nil
}

// ProviderVersionExists checks if a provider version already exists in the database
func ProviderVersionExists(ctx context.Context, db interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}, namespace, name, version string,
) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM provider_versions WHERE provider_namespace = $1 AND provider_name = $2 AND version = $3)`

	err := db.QueryRow(ctx, query, namespace, name, version).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if provider version exists: %w", err)
	}

	return exists, nil
}

// ProviderVersionInfo represents information about a provider version
type ProviderVersionInfo struct {
	DocumentsCount int
	LicensesCount  int
	ProcessedAt    time.Time
}

// GetProviderVersionInfo retrieves information about an existing provider version
func GetProviderVersionInfo(ctx context.Context, db interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}, namespace, name, version string,
) (*ProviderVersionInfo, error) {
	var info ProviderVersionInfo
	query := `SELECT updated_at FROM provider_versions WHERE provider_namespace = $1 AND provider_name = $2 AND version = $3`

	err := db.QueryRow(ctx, query, namespace, name, version).Scan(&info.ProcessedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider version info: %w", err)
	}

	// Set counts to 0 since we don't track them in the schema
	info.DocumentsCount = 0
	info.LicensesCount = 0

	return &info, nil
}

// GetExistingProviderVersions returns all versions that already exist in the database for a given provider
// Includes 'failed' status to avoid retrying failed versions on every run
func GetExistingProviderVersions(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}, namespace, name string,
) ([]string, error) {
	query := `
		SELECT version
		FROM provider_versions
		WHERE provider_namespace = $1
		  AND provider_name = $2
		  AND scrape_status IN ('completed', 'skipped', 'failed')`

	rows, err := db.Query(ctx, query, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to query existing versions: %w", err)
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		versions = append(versions, version)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over versions: %w", err)
	}

	return versions, nil
}

// UpdateProviderVersionStatus updates the scraping status of a provider version
func UpdateProviderVersionStatus(ctx context.Context, db interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}, namespace, name, version, status, skipReason, errorMessage string,
) error {
	query := `
		UPDATE provider_versions
		SET scrape_status = $4,
		    skip_reason = $5,
		    error_message = $6,
		    last_attempt_at = NOW(),
		    updated_at = NOW()
		WHERE provider_namespace = $1
		  AND provider_name = $2
		  AND version = $3`

	result, err := db.Exec(ctx, query, namespace, name, version, status, skipReason, errorMessage)
	if err != nil {
		return fmt.Errorf("failed to update provider version status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("provider version %s/%s@%s not found", namespace, name, version)
	}

	return nil
}

// ProviderVersionStatusInfo represents status information about a provider version
type ProviderVersionStatusInfo struct {
	Version         string
	ScrapeStatus    string
	SkipReason      string
	ErrorMessage    string
	LastAttemptAt   *time.Time
	LicenseAccepted bool
}

// GetSkippedProviderVersions returns all skipped versions for a given provider
func GetSkippedProviderVersions(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}, namespace, name string,
) ([]ProviderVersionStatusInfo, error) {
	query := `
		SELECT version, scrape_status, skip_reason, error_message, last_attempt_at, license_accepted
		FROM provider_versions
		WHERE provider_namespace = $1
		  AND provider_name = $2
		  AND scrape_status = 'skipped'
		ORDER BY version DESC`

	rows, err := db.Query(ctx, query, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to query skipped versions: %w", err)
	}
	defer rows.Close()

	var versions []ProviderVersionStatusInfo
	for rows.Next() {
		var v ProviderVersionStatusInfo
		var skipReason, errorMessage *string
		var lastAttempt *time.Time

		if err := rows.Scan(&v.Version, &v.ScrapeStatus, &skipReason, &errorMessage, &lastAttempt, &v.LicenseAccepted); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}

		if skipReason != nil {
			v.SkipReason = *skipReason
		}
		if errorMessage != nil {
			v.ErrorMessage = *errorMessage
		}
		if lastAttempt != nil {
			v.LastAttemptAt = lastAttempt
		}

		versions = append(versions, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over versions: %w", err)
	}

	return versions, nil
}

// GetFailedProviderVersions returns all failed versions for a given provider
func GetFailedProviderVersions(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}, namespace, name string,
) ([]ProviderVersionStatusInfo, error) {
	query := `
		SELECT version, scrape_status, skip_reason, error_message, last_attempt_at, license_accepted
		FROM provider_versions
		WHERE provider_namespace = $1
		  AND provider_name = $2
		  AND scrape_status = 'failed'
		ORDER BY last_attempt_at DESC`

	rows, err := db.Query(ctx, query, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to query failed versions: %w", err)
	}
	defer rows.Close()

	var versions []ProviderVersionStatusInfo
	for rows.Next() {
		var v ProviderVersionStatusInfo
		var skipReason, errorMessage *string
		var lastAttempt *time.Time

		if err := rows.Scan(&v.Version, &v.ScrapeStatus, &skipReason, &errorMessage, &lastAttempt, &v.LicenseAccepted); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}

		if skipReason != nil {
			v.SkipReason = *skipReason
		}
		if errorMessage != nil {
			v.ErrorMessage = *errorMessage
		}
		if lastAttempt != nil {
			v.LastAttemptAt = lastAttempt
		}

		versions = append(versions, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over versions: %w", err)
	}

	return versions, nil
}

// DocItem represents a documentation item for storage operations
type DocItem struct {
	Name        string
	Title       string
	Subcategory string
	Description string
	EditLink    string
	MD5Checksum string
}

// DocTypeConfig defines the configuration for a documentation type
type DocTypeConfig struct {
	SourceDirs []string
	TargetPath string
	FieldName  string
}

// DocTypes defines the mapping between source directories and target categorization paths.
//
// TO ADD A NEW DOCUMENTATION TYPE:
//  1. Add a new field to the scraper.ProviderDocs struct (e.g., `MyType []DocItem `json:"mytype,omitempty"`)
//  2. Add a new entry to this map with:
//     - SourceDirs: directories to search for docs (legacy and modern formats)
//     - TargetPath: the normalized path used in storage and JSON output (must match DB constraint)
//     - FieldName: the struct field name from step 1 (e.g., "MyType")
//  3. Add a database migration to update the CHECK constraint on provider_documents.document_type
//     (see pkg/db/migrate.go, migration 26 for the current constraint)
//
// The scraping, categorization, and JSON building all derive from this map automatically.
var DocTypes = map[string]DocTypeConfig{
	"resources": {
		SourceDirs: []string{"r", "resources"},
		TargetPath: "resources",
		FieldName:  "Resources",
	},
	"datasources": {
		SourceDirs: []string{"d", "data-sources", "datasources"},
		TargetPath: "datasources",
		FieldName:  "DataSources",
	},
	"functions": {
		SourceDirs: []string{"f", "functions"},
		TargetPath: "functions",
		FieldName:  "Functions",
	},
	"guides": {
		SourceDirs: []string{"guides"},
		TargetPath: "guides",
		FieldName:  "Guides",
	},
	"ephemeral": {
		SourceDirs: []string{"ephemeral", "ephemeral-resources"},
		TargetPath: "ephemeral",
		FieldName:  "Ephemeral",
	},
	"actions": {
		SourceDirs: []string{"actions"},
		TargetPath: "actions",
		FieldName:  "Actions",
	},
	"listresources": {
		SourceDirs: []string{"list-resources", "listresources"},
		TargetPath: "listresources",
		FieldName:  "ListResources",
	},
}

// StoreProviderDocuments stores all document metadata in the database using bulk inserts
func StoreProviderDocuments(ctx context.Context, tx pgx.Tx, namespace, name, version string, docs map[string]*DocItem) error {
	ctx, span := telemetry.Tracer().Start(ctx, "storage.StoreProviderDocuments")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
		attribute.String("provider.version", version),
		attribute.Int("docs.count", len(docs)),
	)

	if len(docs) == 0 {
		slog.InfoContext(ctx, "No documents to store in database",
			"namespace", namespace, "name", name, "version", version)
		return nil
	}

	slog.DebugContext(ctx, "Storing document metadata in database",
		"namespace", namespace, "name", name, "version", version, "docs_count", len(docs))

	// Use pgx.Batch for efficient bulk insert within provided transaction
	batch := &pgx.Batch{}

	for filePath, doc := range docs {
		docType := getDocCategory(filePath)
		if docType == "" {
			slog.WarnContext(ctx, "Unknown document type, skipping database storage",
				"filePath", filePath, "document_name", doc.Name)
			continue
		}

		language := extractLanguage(filePath)
		s3Key := fmt.Sprintf("providers/%s/%s/%s/%s.md", namespace, name, version, filePath)

		batch.Queue(`
			INSERT INTO provider_documents
			(provider_namespace, provider_name, version, document_type, document_name,
			 title, subcategory, description, edit_link, s3_key, language, md5_checksum)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (provider_namespace, provider_name, version, document_type, document_name, language)
			DO UPDATE SET
				title = EXCLUDED.title,
				subcategory = EXCLUDED.subcategory,
				description = EXCLUDED.description,
				edit_link = EXCLUDED.edit_link,
				s3_key = EXCLUDED.s3_key,
				md5_checksum = EXCLUDED.md5_checksum,
				updated_at = NOW()
		`, namespace, name, version, docType, doc.Name,
			doc.Title, doc.Subcategory, doc.Description, doc.EditLink, s3Key, language, doc.MD5Checksum)
	}

	batchResults := tx.SendBatch(ctx, batch)
	defer batchResults.Close()

	insertCount := 0
	for i := 0; i < batch.Len(); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			span.RecordError(err)
			slog.ErrorContext(ctx, "Failed to insert document metadata",
				"batch_index", i, "error", err)
			return fmt.Errorf("failed to insert document metadata at batch index %d: %w", i, err)
		}
		insertCount++
	}

	// Transaction commit is handled by the caller

	span.SetAttributes(attribute.Int("docs.inserted_count", insertCount))
	slog.DebugContext(ctx, "Successfully stored document metadata in database",
		"namespace", namespace, "name", name, "version", version,
		"inserted_count", insertCount, "total_docs", len(docs))

	return nil
}

// getDocCategory determines the category of a document based on its file path
// Returns the category name (resources, datasources, functions, guides, ephemeral, index) or empty string if no match
func getDocCategory(filePath string) string {
	if strings.Contains(filePath, "index") {
		return "index"
	}
	parts := strings.Split(filePath, "/")

	for _, config := range DocTypes {
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

// GetAllDocumentChecksums retrieves all MD5 checksums for a provider version in one query
// Returns a map with key "docType:docName:language" -> checksum
func GetAllDocumentChecksums(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}, namespace, name, version string,
) (map[string]string, error) {
	query := `
		SELECT document_type, document_name, language, md5_checksum
		FROM provider_documents
		WHERE provider_namespace = $1
		  AND provider_name = $2
		  AND version = $3
		  AND md5_checksum IS NOT NULL`

	rows, err := db.Query(ctx, query, namespace, name, version)
	if err != nil {
		return nil, fmt.Errorf("failed to query document checksums: %w", err)
	}
	defer rows.Close()

	checksums := make(map[string]string)
	for rows.Next() {
		var docType, docName, language, checksum string
		if err := rows.Scan(&docType, &docName, &language, &checksum); err != nil {
			return nil, fmt.Errorf("failed to scan document checksum: %w", err)
		}
		key := fmt.Sprintf("%s:%s:%s", docType, docName, language)
		checksums[key] = checksum
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating document checksums: %w", err)
	}

	return checksums, nil
}

// StoreProviderLicenses stores detailed license information for a provider version
func StoreProviderLicenses(ctx context.Context, tx pgx.Tx, namespace, name, version string, licenses license.List) error {
	// First ensure all licenses exist in the licenses table
	for _, lic := range licenses {
		licenseQuery := `
			INSERT INTO licenses (spdx_id, name, redistributable)
			VALUES ($1, $2, $3)
			ON CONFLICT (spdx_id) DO NOTHING`

		// Determine if license is redistributable (compatible)
		redistributable := lic.IsCompatible

		// Use SPDX ID as name if no specific name available
		licenseName := lic.SPDX
		if licenseName == "" {
			licenseName = "Unknown"
		}

		_, err := tx.Exec(ctx, licenseQuery, lic.SPDX, licenseName, redistributable)
		if err != nil {
			return fmt.Errorf("failed to ensure license %s exists: %w", lic.SPDX, err)
		}
	}

	// Delete existing licenses for this provider version
	deleteQuery := `
		DELETE FROM provider_version_licenses
		WHERE provider_namespace = $1 AND provider_name = $2 AND version = $3`

	_, err := tx.Exec(ctx, deleteQuery, namespace, name, version)
	if err != nil {
		return fmt.Errorf("failed to delete existing provider licenses: %w", err)
	}

	// Insert new licenses
	insertQuery := `
		INSERT INTO provider_version_licenses (
			provider_namespace, provider_name, version,
			license_spdx_id, confidence_score, file_path, match_type
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	for _, lic := range licenses {
		matchType := "detected"
		if lic.IsCompatible {
			matchType = "compatible"
		}

		_, err = tx.Exec(ctx, insertQuery,
			namespace, name, version,
			lic.SPDX, float64(lic.Confidence), lic.File, matchType)
		if err != nil {
			return fmt.Errorf("failed to insert provider license %s: %w", lic.SPDX, err)
		}
	}

	return nil
}
