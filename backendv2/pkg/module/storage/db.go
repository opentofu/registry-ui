package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/opentofu/registry-ui/pkg/license"
)

// Queryable is an interface that both pgx.Tx and *pgxpool.Pool implement for database operations
type Queryable interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// StoreModule ensures a module record exists in the database.
// Accepts Queryable interface so it can be called with either pgx.Tx or *pgxpool.Pool
func StoreModule(ctx context.Context, db Queryable, namespace, name, target, repoOrganisation, repoName string) error {
	query := `
		INSERT INTO modules (namespace, name, target, repo_organisation, repo_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (namespace, name, target)
		DO UPDATE SET
			updated_at = NOW()`

	_, err := db.Exec(ctx, query, namespace, name, target, repoOrganisation, repoName)
	if err != nil {
		return fmt.Errorf("failed to store module: %w", err)
	}

	return nil
}

// StoreRepository stores repository metadata in the database
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

// StoreModuleVersion stores module version information in the database
func StoreModuleVersion(ctx context.Context, tx pgx.Tx, namespace, name, target, version string, tofuJSON any, tagCreatedAt *time.Time, scrapeStatus, skipReason, errorMessage, indexChecksum, readmeChecksum string) error {
	// Convert the moduleData to JSON
	jsonData, err := json.Marshal(tofuJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal tofu JSON: %w", err)
	}

	query := `
		INSERT INTO module_versions (module_namespace, module_name, module_target, version, tofu_json, processed_at, tag_created_at, scrape_status, skip_reason, error_message, last_attempt_at, index_md5_checksum, readme_md5_checksum)
		VALUES ($1, $2, $3, $4, $5, NOW(), $6, $7, $8, $9, NOW(), $10, $11)
		ON CONFLICT (module_namespace, module_name, module_target, version)
		DO UPDATE SET
			tofu_json = EXCLUDED.tofu_json,
			processed_at = EXCLUDED.processed_at,
			tag_created_at = EXCLUDED.tag_created_at,
			scrape_status = EXCLUDED.scrape_status,
			skip_reason = EXCLUDED.skip_reason,
			error_message = EXCLUDED.error_message,
			last_attempt_at = NOW(),
			index_md5_checksum = EXCLUDED.index_md5_checksum,
			readme_md5_checksum = EXCLUDED.readme_md5_checksum`

	_, err = tx.Exec(ctx, query, namespace, name, target, version, jsonData, tagCreatedAt, scrapeStatus, skipReason, errorMessage, indexChecksum, readmeChecksum)
	if err != nil {
		return fmt.Errorf("failed to store module version: %w", err)
	}

	return nil
}

// StoreModuleSubmodule stores submodule information in the database
func StoreModuleSubmodule(ctx context.Context, tx pgx.Tx, namespace, name, target, version, submoduleName, submodulePath string, tofuJSON any, indexChecksum, readmeChecksum string) error {
	// Convert the moduleData to JSON
	jsonData, err := json.Marshal(tofuJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal tofu JSON: %w", err)
	}

	query := `
		INSERT INTO module_submodules (module_namespace, module_name, module_target, version, submodule_name, path, tofu_json, index_md5_checksum, readme_md5_checksum)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (module_namespace, module_name, module_target, version, submodule_name)
		DO UPDATE SET
			path = EXCLUDED.path,
			tofu_json = EXCLUDED.tofu_json,
			index_md5_checksum = EXCLUDED.index_md5_checksum,
			readme_md5_checksum = EXCLUDED.readme_md5_checksum`

	_, err = tx.Exec(ctx, query, namespace, name, target, version, submoduleName, submodulePath, jsonData, indexChecksum, readmeChecksum)
	if err != nil {
		return fmt.Errorf("failed to store module submodule: %w", err)
	}

	return nil
}

// StoreModuleExample stores example information in the database
func StoreModuleExample(ctx context.Context, tx pgx.Tx, namespace, name, target, version, exampleName, examplePath string, tofuJSON any, indexChecksum, readmeChecksum string) error {
	// Convert the moduleData to JSON
	jsonData, err := json.Marshal(tofuJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal tofu JSON: %w", err)
	}

	query := `
		INSERT INTO module_examples (module_namespace, module_name, module_target, version, example_name, path, tofu_json, index_md5_checksum, readme_md5_checksum)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (module_namespace, module_name, module_target, version, example_name)
		DO UPDATE SET
			path = EXCLUDED.path,
			tofu_json = EXCLUDED.tofu_json,
			index_md5_checksum = EXCLUDED.index_md5_checksum,
			readme_md5_checksum = EXCLUDED.readme_md5_checksum`

	_, err = tx.Exec(ctx, query, namespace, name, target, version, exampleName, examplePath, jsonData, indexChecksum, readmeChecksum)
	if err != nil {
		return fmt.Errorf("failed to store module example: %w", err)
	}

	return nil
}

// StoreModuleVersionLicenses stores detailed license information for a module version
func StoreModuleVersionLicenses(ctx context.Context, tx pgx.Tx, namespace, name, target, version string, licenses license.List) error {
	// First ensure all licenses exist in the licenses table
	for _, lic := range licenses {
		licenseQuery := `
			INSERT INTO licenses (spdx_id, name, category, redistributable, url)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (spdx_id) DO NOTHING`

		// Determine if license is redistributable (compatible)
		redistributable := lic.IsCompatible

		// Use SPDX ID as name if no specific name available
		licenseName := lic.SPDX
		if licenseName == "" {
			licenseName = "Unknown"
		}

		_, err := tx.Exec(ctx, licenseQuery, lic.SPDX, licenseName, "detected", redistributable, "")
		if err != nil {
			return fmt.Errorf("failed to ensure license %s exists: %w", lic.SPDX, err)
		}
	}

	// Delete existing licenses for this module version
	deleteQuery := `
		DELETE FROM module_version_licenses 
		WHERE module_namespace = $1 AND module_name = $2 AND module_target = $3 AND version = $4`

	_, err := tx.Exec(ctx, deleteQuery, namespace, name, target, version)
	if err != nil {
		return fmt.Errorf("failed to delete existing module licenses: %w", err)
	}

	// Insert new licenses
	insertQuery := `
		INSERT INTO module_version_licenses (
			module_namespace, module_name, module_target, version, 
			license_spdx_id, confidence_score, file_path, match_type
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	for _, lic := range licenses {
		matchType := "detected"
		if lic.IsCompatible {
			matchType = "compatible"
		}

		_, err = tx.Exec(ctx, insertQuery,
			namespace, name, target, version,
			lic.SPDX, float64(lic.Confidence), lic.File, matchType)
		if err != nil {
			return fmt.Errorf("failed to insert module license %s: %w", lic.SPDX, err)
		}
	}

	return nil
}

// GetExistingModuleVersions returns all versions that already exist in the database for a given module
// Includes 'failed' status to avoid retrying failed versions on every run
func GetExistingModuleVersions(ctx context.Context, db Queryable, namespace, name, target string) ([]string, error) {
	query := `
		SELECT version
		FROM module_versions
		WHERE module_namespace = $1
		  AND module_name = $2
		  AND module_target = $3
		  AND scrape_status IN ('completed', 'skipped', 'failed')`

	rows, err := db.Query(ctx, query, namespace, name, target)
	if err != nil {
		return nil, fmt.Errorf("failed to query existing module versions: %w", err)
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan module version: %w", err)
		}
		versions = append(versions, version)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over module versions: %w", err)
	}

	return versions, nil
}

// UpdateModuleVersionStatus updates the scraping status of a module version
func UpdateModuleVersionStatus(ctx context.Context, db Queryable, namespace, name, target, version, status, skipReason, errorMessage string) error {
	query := `
		UPDATE module_versions
		SET scrape_status = $5,
		    skip_reason = $6,
		    error_message = $7,
		    last_attempt_at = NOW(),
		    updated_at = NOW()
		WHERE module_namespace = $1
		  AND module_name = $2
		  AND module_target = $3
		  AND version = $4`

	result, err := db.Exec(ctx, query, namespace, name, target, version, status, skipReason, errorMessage)
	if err != nil {
		return fmt.Errorf("failed to update module version status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("module version %s/%s/%s@%s not found", namespace, name, target, version)
	}

	return nil
}
