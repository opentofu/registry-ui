package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel/attribute"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

func getDBConfig(cmd *cli.Command) config.DBConfig {
	return config.FromCLI(cmd).DB
}

/*
MIGRATION DEVELOPER GUIDE

## Adding a New Migration
- Use the next sequential ID number
- Use descriptive snake_case names: create_X_table, add_X_to_Y, rename_X_to_Y, remove_X_from_Y
- Write a clear Description field explaining the business purpose

## Required Documentation (Critical!)
- Add COMMENT ON TABLE for every table you create
- Add COMMENT ON COLUMN for every column you create or modify
- Document all indices with inline SQL comments
- Write descriptive migration names and descriptions
- Explain WHY the change is needed, not just WHAT changed

## UP/DOWN Symmetry Requirements
- DOWN should reverse ALL changes in UP
- Drop objects in reverse order of creation: triggers → indices → columns → tables
- Always use IF EXISTS guards in DROP statements
- Always use IF NOT EXISTS guards in CREATE statements
- Test both directions before submitting

## Best Practices
- Use ON DELETE CASCADE where child records should be removed with parent
- Include created_at and updated_at TIMESTAMP WITH TIME ZONE on tables
- Preserve data when restructuring (use INSERT ... SELECT patterns)
- Keep migrations focused on a single logical change
- Add inline comments for complex SQL operations
*/

// Migration represents a single database migration
type Migration struct {
	ID          int
	Name        string
	Description string
	Up          string
	Down        string
}

// All migrations in order
var migrations = []Migration{
	{
		ID:          1,
		Name:        "create_repositories_table",
		Description: "Create the repositories table with organisation, name, and fork information",
		Up: `
CREATE TABLE repositories (
    organisation VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    fork_count INTEGER DEFAULT 0,
    is_fork BOOLEAN DEFAULT false,
    parent_organisation VARCHAR(255),
    parent_name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    PRIMARY KEY (organisation, name),
    FOREIGN KEY (parent_organisation, parent_name)
        REFERENCES repositories(organisation, name)
);`,
		Down: `DROP TABLE repositories;`,
	},
	{
		ID:          2,
		Name:        "create_providers_table",
		Description: "Create the providers table with namespace, name, and repository references",
		Up: `
CREATE TABLE providers (
    namespace VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    alias VARCHAR(255),
    warnings TEXT[] DEFAULT '{}',
    repo_organisation VARCHAR(255) NOT NULL,
    repo_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    PRIMARY KEY (namespace, name),
    FOREIGN KEY (repo_organisation, repo_name)
        REFERENCES repositories(organisation, name)
);`,
		Down: `DROP TABLE providers;`,
	},
	{
		ID:          3,
		Name:        "create_provider_versions_table",
		Description: "Create provider_versions table to store version information for providers",
		Up: `
CREATE TABLE provider_versions (
    id SERIAL PRIMARY KEY,
    provider_namespace VARCHAR(255) NOT NULL,
    provider_name VARCHAR(255) NOT NULL,
    version VARCHAR(255) NOT NULL,
    tag_created_at TIMESTAMP WITH TIME ZONE,
    discovered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (provider_namespace, provider_name) 
        REFERENCES providers(namespace, name) ON DELETE CASCADE,
    UNIQUE(provider_namespace, provider_name, version)
);

CREATE INDEX idx_provider_versions_tag_created_at ON provider_versions(tag_created_at);
CREATE INDEX idx_provider_versions_discovered_at ON provider_versions(discovered_at);`,
		Down: `
DROP INDEX IF EXISTS idx_provider_versions_discovered_at;
DROP INDEX IF EXISTS idx_provider_versions_tag_created_at;
DROP TABLE provider_versions;`,
	},
	{
		ID:          4,
		Name:        "add_license_accepted_to_provider_versions",
		Description: "Add license_accepted field to track whether provider version licenses are acceptable",
		Up: `
ALTER TABLE provider_versions ADD COLUMN license_accepted BOOLEAN DEFAULT false;
CREATE INDEX idx_provider_versions_license_accepted ON provider_versions(license_accepted);`,
		Down: `
DROP INDEX IF EXISTS idx_provider_versions_license_accepted;
ALTER TABLE provider_versions DROP COLUMN license_accepted;`,
	},
	{
		ID:          5,
		Name:        "create_licenses_table",
		Description: "Create the licenses table to store license information with SPDX identifiers",
		Up: `
CREATE TABLE licenses (
    spdx_id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(50),
    redistributable BOOLEAN DEFAULT false,
    url VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_licenses_category ON licenses(category);
CREATE INDEX idx_licenses_redistributable ON licenses(redistributable);`,
		Down: `
DROP INDEX IF EXISTS idx_licenses_redistributable;
DROP INDEX IF EXISTS idx_licenses_category;
DROP TABLE licenses;`,
	},
	{
		ID:          6,
		Name:        "create_provider_version_licenses_table",
		Description: "Create the junction table to link provider versions with their detected licenses",
		Up: `
CREATE TABLE provider_version_licenses (
    id SERIAL PRIMARY KEY,
    provider_namespace VARCHAR(255) NOT NULL,
    provider_name VARCHAR(255) NOT NULL,
    version VARCHAR(255) NOT NULL,
    license_spdx_id VARCHAR(255) NOT NULL,
    confidence_score DECIMAL(4,3),
    file_path TEXT,
    match_type VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (provider_namespace, provider_name, version) 
        REFERENCES provider_versions(provider_namespace, provider_name, version) ON DELETE CASCADE,
    FOREIGN KEY (license_spdx_id) REFERENCES licenses(spdx_id) ON DELETE CASCADE,
    UNIQUE(provider_namespace, provider_name, version, license_spdx_id, file_path)
);

CREATE INDEX idx_provider_version_licenses_provider ON provider_version_licenses(provider_namespace, provider_name, version);
CREATE INDEX idx_provider_version_licenses_license ON provider_version_licenses(license_spdx_id);
CREATE INDEX idx_provider_version_licenses_confidence ON provider_version_licenses(confidence_score);`,
		Down: `
DROP INDEX IF EXISTS idx_provider_version_licenses_confidence;
DROP INDEX IF EXISTS idx_provider_version_licenses_license;
DROP INDEX IF EXISTS idx_provider_version_licenses_provider;
DROP TABLE provider_version_licenses;`,
	},
	{
		ID:          7,
		Name:        "create_provider_aliases_table",
		Description: "Create the provider_aliases table to map original provider addresses to target addresses",
		Up: `
CREATE TABLE provider_aliases (
    id SERIAL PRIMARY KEY,
    original_namespace VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    target_namespace VARCHAR(255) NOT NULL,
    target_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(original_namespace, original_name)
);

CREATE INDEX idx_provider_aliases_original ON provider_aliases(original_namespace, original_name);
CREATE INDEX idx_provider_aliases_target ON provider_aliases(target_namespace, target_name);`,
		Down: `
DROP INDEX IF EXISTS idx_provider_aliases_target;
DROP INDEX IF EXISTS idx_provider_aliases_original;
DROP TABLE provider_aliases;`,
	},
	{
		ID:          8,
		Name:        "populate_provider_aliases",
		Description: "Populate provider_aliases table with data from libregistry",
		Up: `
INSERT INTO provider_aliases (original_namespace, original_name, target_namespace, target_name) VALUES
('opentofu', 'aci', 'CiscoDevNet', 'aci'),
('opentofu', 'ad', 'hashicorp', 'ad'),
('opentofu', 'akamai', 'akamai', 'akamai'),
('opentofu', 'alicloud', 'aliyun', 'alicloud'),
('opentofu', 'archive', 'hashicorp', 'archive'),
('opentofu', 'assert', 'hashicorp', 'assert'),
('opentofu', 'auth0', 'auth0', 'auth0'),
('opentofu', 'aws', 'hashicorp', 'aws'),
('opentofu', 'awscc', 'hashicorp', 'awscc'),
('opentofu', 'azuread', 'hashicorp', 'azuread'),
('opentofu', 'azurerm', 'hashicorp', 'azurerm'),
('opentofu', 'azurestack', 'hashicorp', 'azurestack'),
('opentofu', 'boundary', 'hashicorp', 'boundary'),
('opentofu', 'cloudinit', 'hashicorp', 'cloudinit'),
('opentofu', 'consul', 'hashicorp', 'consul'),
('opentofu', 'dns', 'hashicorp', 'dns'),
('opentofu', 'external', 'hashicorp', 'external'),
('opentofu', 'github', 'integrations', 'github'),
('opentofu', 'gitlab', 'gitlabhq', 'gitlab'),
('opentofu', 'google', 'hashicorp', 'google'),
('opentofu', 'google-beta', 'hashicorp', 'google-beta'),
('opentofu', 'googleworkspace', 'hashicorp', 'googleworkspace'),
('opentofu', 'hcp', 'hashicorp', 'hcp'),
('opentofu', 'hcs', 'hashicorp', 'hcs'),
('opentofu', 'helm', 'hashicorp', 'helm'),
('opentofu', 'http', 'hashicorp', 'http'),
('opentofu', 'kubernetes', 'hashicorp', 'kubernetes'),
('opentofu', 'local', 'hashicorp', 'local'),
('opentofu', 'nomad', 'hashicorp', 'nomad'),
('opentofu', 'null', 'hashicorp', 'null'),
('opentofu', 'oci', 'oracle', 'oci'),
('opentofu', 'opc', 'hashicorp', 'opc'),
('opentofu', 'oraclepaas', 'hashicorp', 'oraclepaas'),
('opentofu', 'random', 'hashicorp', 'random'),
('opentofu', 'salesforce', 'hashicorp', 'salesforce'),
('opentofu', 'template', 'hashicorp', 'template'),
('opentofu', 'tfe', 'hashicorp', 'tfe'),
('opentofu', 'time', 'hashicorp', 'time'),
('opentofu', 'tls', 'hashicorp', 'tls'),
('opentofu', 'vault', 'hashicorp', 'vault'),
('opentofu', 'vsphere', 'hashicorp', 'vsphere')
ON CONFLICT (original_namespace, original_name) DO NOTHING;`,
		Down: `DELETE FROM provider_aliases WHERE original_namespace = 'opentofu';`,
	},
	{
		ID:          9,
		Name:        "create_provider_documents_table",
		Description: "Create provider_documents table to store document metadata for search functionality",
		Up: `
CREATE TABLE provider_documents (
    id SERIAL PRIMARY KEY,
    provider_namespace VARCHAR(255) NOT NULL,
    provider_name VARCHAR(255) NOT NULL,
    version VARCHAR(255) NOT NULL,
    document_type VARCHAR(50) NOT NULL CHECK (document_type IN ('resources', 'datasources', 'functions', 'guides', 'ephemeral', 'index')),
    document_name VARCHAR(255) NOT NULL,
    title VARCHAR(500),
    subcategory VARCHAR(255),
    description TEXT,
    edit_link VARCHAR(1000),
    s3_key TEXT,
    language VARCHAR(50) NOT NULL DEFAULT 'default',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (provider_namespace, provider_name, version) 
        REFERENCES provider_versions(provider_namespace, provider_name, version) ON DELETE CASCADE,
    UNIQUE(provider_namespace, provider_name, version, document_type, document_name, language)
);

CREATE INDEX idx_provider_documents_provider ON provider_documents(provider_namespace, provider_name);
CREATE INDEX idx_provider_documents_type ON provider_documents(document_type);
CREATE INDEX idx_provider_documents_name ON provider_documents(document_name);
CREATE INDEX idx_provider_documents_subcategory ON provider_documents(subcategory);
CREATE INDEX idx_provider_documents_language ON provider_documents(language);
CREATE INDEX idx_provider_documents_fulltext ON provider_documents USING GIN(to_tsvector('simple', COALESCE(title, '') || ' ' || COALESCE(description, '')));

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_repositories_updated_at 
    BEFORE UPDATE ON repositories 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_providers_updated_at 
    BEFORE UPDATE ON providers 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_provider_versions_updated_at 
    BEFORE UPDATE ON provider_versions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_licenses_updated_at 
    BEFORE UPDATE ON licenses 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_provider_aliases_updated_at 
    BEFORE UPDATE ON provider_aliases 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_provider_documents_updated_at 
    BEFORE UPDATE ON provider_documents 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE provider_documents IS 'Stores metadata for provider documentation with multi-language support';
COMMENT ON COLUMN provider_documents.document_type IS 'Type: resources, datasources, functions, guides, ephemeral, or index';
COMMENT ON COLUMN provider_documents.language IS 'Language code (typescript, python, java, csharp, go) or default for standard docs';`,
		Down: `
DROP TRIGGER IF EXISTS update_provider_documents_updated_at ON provider_documents;
DROP TRIGGER IF EXISTS update_provider_aliases_updated_at ON provider_aliases;
DROP TRIGGER IF EXISTS update_licenses_updated_at ON licenses;
DROP TRIGGER IF EXISTS update_provider_versions_updated_at ON provider_versions;
DROP TRIGGER IF EXISTS update_providers_updated_at ON providers;
DROP TRIGGER IF EXISTS update_repositories_updated_at ON repositories;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP INDEX IF EXISTS idx_provider_documents_fulltext;
DROP INDEX IF EXISTS idx_provider_documents_language;
DROP INDEX IF EXISTS idx_provider_documents_subcategory;
DROP INDEX IF EXISTS idx_provider_documents_name;
DROP INDEX IF EXISTS idx_provider_documents_type;
DROP INDEX IF EXISTS idx_provider_documents_provider;
DROP TABLE provider_documents;`,
	},
	{
		ID:          10,
		Name:        "add_repository_metadata_fields",
		Description: "Add popularity and description fields to repositories table",
		Up: `
ALTER TABLE repositories 
ADD COLUMN popularity INTEGER DEFAULT 0,
ADD COLUMN description TEXT;

CREATE INDEX idx_repositories_popularity ON repositories(popularity DESC);`,
		Down: `
DROP INDEX IF EXISTS idx_repositories_popularity;
ALTER TABLE repositories 
DROP COLUMN description,
DROP COLUMN popularity;`,
	},
	{
		ID:          11,
		Name:        "rename_popularity_to_stars",
		Description: "Rename popularity column to stars in repositories table",
		Up: `
ALTER TABLE repositories RENAME COLUMN popularity TO stars;
DROP INDEX IF EXISTS idx_repositories_popularity;
CREATE INDEX idx_repositories_stars ON repositories(stars DESC);`,
		Down: `
DROP INDEX IF EXISTS idx_repositories_stars;
ALTER TABLE repositories RENAME COLUMN stars TO popularity;
CREATE INDEX idx_repositories_popularity ON repositories(popularity DESC);`,
	},
	{
		ID:          12,
		Name:        "create_repository_stats_table",
		Description: "Create repository_stats table to track repository metrics over time",
		Up: `
CREATE TABLE repository_stats (
    id SERIAL PRIMARY KEY,
    repo_organisation VARCHAR(255) NOT NULL,
    repo_name VARCHAR(255) NOT NULL,
    stars INTEGER DEFAULT 0,
    forks INTEGER DEFAULT 0,
    watchers INTEGER DEFAULT 0,
    open_issues INTEGER DEFAULT 0,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (repo_organisation, repo_name) 
        REFERENCES repositories(organisation, name) ON DELETE CASCADE
);

CREATE INDEX idx_repository_stats_repo ON repository_stats(repo_organisation, repo_name);
CREATE INDEX idx_repository_stats_recorded_at ON repository_stats(recorded_at);
CREATE INDEX idx_repository_stats_stars ON repository_stats(stars DESC);
CREATE INDEX idx_repository_stats_repo_time ON repository_stats(repo_organisation, repo_name, recorded_at DESC);

-- Migrate existing star data to repository_stats
INSERT INTO repository_stats (repo_organisation, repo_name, stars, recorded_at)
SELECT organisation, name, COALESCE(stars, 0), NOW()
FROM repositories
WHERE stars IS NOT NULL AND stars > 0;


COMMENT ON TABLE repository_stats IS 'Time-series data for repository metrics (stars, forks, etc.)';
COMMENT ON COLUMN repository_stats.recorded_at IS 'When these stats were recorded from GitHub API';`,
		Down: `
DROP INDEX IF EXISTS idx_repository_stats_repo_time;
DROP INDEX IF EXISTS idx_repository_stats_stars;
DROP INDEX IF EXISTS idx_repository_stats_recorded_at;
DROP INDEX IF EXISTS idx_repository_stats_repo;
DROP TABLE repository_stats;`,
	},
	{
		ID:          13,
		Name:        "remove_stats_from_repositories",
		Description: "Remove stats fields from repositories table since they're now tracked in repository_stats",
		Up: `
-- Remove stats columns from repositories table
ALTER TABLE repositories 
DROP COLUMN IF EXISTS stars,
DROP COLUMN IF EXISTS fork_count;

-- Remove related indexes
DROP INDEX IF EXISTS idx_repositories_stars;`,
		Down: `
-- Re-add stats columns to repositories table
ALTER TABLE repositories 
ADD COLUMN stars INTEGER DEFAULT 0,
ADD COLUMN fork_count INTEGER DEFAULT 0;

-- Populate with latest stats from repository_stats
UPDATE repositories SET 
    stars = COALESCE((
        SELECT stars 
        FROM repository_stats 
        WHERE repo_organisation = repositories.organisation 
        AND repo_name = repositories.name 
        ORDER BY recorded_at DESC 
        LIMIT 1
    ), 0),
    fork_count = COALESCE((
        SELECT forks 
        FROM repository_stats 
        WHERE repo_organisation = repositories.organisation 
        AND repo_name = repositories.name 
        ORDER BY recorded_at DESC 
        LIMIT 1
    ), 0);

CREATE INDEX idx_repositories_stars ON repositories(stars DESC);`,
	},
	{
		ID:          14,
		Name:        "add_redirect_tracking_to_repositories",
		Description: "Add fields to track GitHub repository redirects in repositories table",
		Up: `
-- Add redirect tracking fields for GitHub redirects (not forks)
ALTER TABLE repositories 
ADD COLUMN original_organisation VARCHAR(255),
ADD COLUMN original_name VARCHAR(255),
ADD COLUMN is_redirect BOOLEAN DEFAULT false;

-- Add index for redirect lookups
CREATE INDEX idx_repositories_redirect ON repositories(is_redirect, original_organisation, original_name);

-- Add comments to explain the new fields
COMMENT ON COLUMN repositories.original_organisation IS 'Original requested GitHub organisation if this repository was accessed via redirect (e.g. terraform-providers)';
COMMENT ON COLUMN repositories.original_name IS 'Original requested GitHub repository name if this repository was accessed via redirect';
COMMENT ON COLUMN repositories.is_redirect IS 'True if this repository was discovered via GitHub redirect (not fork)';`,
		Down: `
-- Remove redirect tracking fields
DROP INDEX IF EXISTS idx_repositories_redirect;
ALTER TABLE repositories 
DROP COLUMN IF EXISTS is_redirect,
DROP COLUMN IF EXISTS original_name,
DROP COLUMN IF EXISTS original_organisation;`,
	},
	{
		ID:          15,
		Name:        "create_repository_redirects_table",
		Description: "Create repository_redirects table for GitHub repository redirects",
		Up: `
CREATE TABLE repository_redirects (
    from_organisation VARCHAR(255) NOT NULL,
    from_name VARCHAR(255) NOT NULL,
    to_organisation VARCHAR(255) NOT NULL,
    to_name VARCHAR(255) NOT NULL,
    reason VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    PRIMARY KEY (from_organisation, from_name),
    UNIQUE(from_organisation, from_name)
);

CREATE INDEX idx_repository_redirects_to ON repository_redirects(to_organisation, to_name);
CREATE INDEX idx_repository_redirects_reason ON repository_redirects(reason);

-- Add trigger for updated_at
CREATE TRIGGER update_repository_redirects_updated_at 
    BEFORE UPDATE ON repository_redirects 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE repository_redirects IS 'Tracks GitHub repository redirects (e.g. terraform-providers -> hashicorp)';
COMMENT ON COLUMN repository_redirects.reason IS 'Reason for redirect: github_redirect, moved, etc.';`,
		Down: `
DROP TRIGGER IF EXISTS update_repository_redirects_updated_at ON repository_redirects;
DROP INDEX IF EXISTS idx_repository_redirects_reason;
DROP INDEX IF EXISTS idx_repository_redirects_to;
DROP TABLE repository_redirects;`,
	},
	{
		ID:          16,
		Name:        "migrate_redirect_data_to_repository_redirects",
		Description: "Migrate existing redirect data from repositories table to repository_redirects table",
		Up: `
-- Copy redirect data from repositories table to repository_redirects table
INSERT INTO repository_redirects (from_organisation, from_name, to_organisation, to_name, reason, created_at)
SELECT 
    original_organisation,
    original_name,
    organisation,
    name,
    'github_redirect' as reason,
    created_at
FROM repositories 
WHERE is_redirect = true 
  AND original_organisation IS NOT NULL 
  AND original_name IS NOT NULL
ON CONFLICT (from_organisation, from_name) DO NOTHING;`,
		Down: `
-- Remove migrated redirect data (we can't reliably restore it to repositories table)
DELETE FROM repository_redirects WHERE reason = 'github_redirect';`,
	},
	{
		ID:          17,
		Name:        "remove_redirect_fields_from_repositories",
		Description: "Remove redirect fields from repositories table now that they're in repository_redirects",
		Up: `
-- Remove redirect tracking fields from repositories table
DROP INDEX IF EXISTS idx_repositories_redirect;
ALTER TABLE repositories 
DROP COLUMN IF EXISTS is_redirect,
DROP COLUMN IF EXISTS original_name,
DROP COLUMN IF EXISTS original_organisation;`,
		Down: `
-- Re-add redirect fields to repositories table
ALTER TABLE repositories 
ADD COLUMN original_organisation VARCHAR(255),
ADD COLUMN original_name VARCHAR(255),
ADD COLUMN is_redirect BOOLEAN DEFAULT false;

CREATE INDEX idx_repositories_redirect ON repositories(is_redirect, original_organisation, original_name);

-- Restore redirect data from repository_redirects table
UPDATE repositories SET 
    is_redirect = true,
    original_organisation = rr.from_organisation,
    original_name = rr.from_name
FROM repository_redirects rr
WHERE repositories.organisation = rr.to_organisation 
  AND repositories.name = rr.to_name 
  AND rr.reason = 'github_redirect';`,
	},
	{
		ID:          18,
		Name:        "add_additional_stats_to_repository_stats",
		Description: "Add subscribers and topics fields to repository_stats table",
		Up: `
-- Add new fields to repository_stats table
ALTER TABLE repository_stats 
ADD COLUMN subscribers INTEGER DEFAULT 0,
ADD COLUMN topics TEXT[] DEFAULT '{}';

-- Add indexes for new fields
CREATE INDEX idx_repository_stats_subscribers ON repository_stats(subscribers DESC);
CREATE INDEX idx_repository_stats_topics ON repository_stats USING GIN(topics);

-- Add comments to explain the new fields
COMMENT ON COLUMN repository_stats.subscribers IS 'Number of repository subscribers/watchers from GitHub API';
COMMENT ON COLUMN repository_stats.topics IS 'Array of repository topics/tags from GitHub API';`,
		Down: `
-- Remove new fields and indexes
DROP INDEX IF EXISTS idx_repository_stats_topics;
DROP INDEX IF EXISTS idx_repository_stats_subscribers;
ALTER TABLE repository_stats 
DROP COLUMN IF EXISTS topics,
DROP COLUMN IF EXISTS subscribers;`,
	},
	{
		ID:          19,
		Name:        "add_repository_metadata_fields",
		Description: "Add homepage, language, archived status and GitHub timestamps to repositories table",
		Up: `
-- Add new metadata fields to repositories table
ALTER TABLE repositories 
ADD COLUMN homepage TEXT,
ADD COLUMN language VARCHAR(50),
ADD COLUMN archived BOOLEAN DEFAULT false,
ADD COLUMN default_branch VARCHAR(255),
ADD COLUMN created_at_github TIMESTAMP WITH TIME ZONE,
ADD COLUMN pushed_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN updated_at_github TIMESTAMP WITH TIME ZONE;

-- Add indexes for commonly queried fields
CREATE INDEX idx_repositories_archived ON repositories(archived);
CREATE INDEX idx_repositories_language ON repositories(language);
CREATE INDEX idx_repositories_pushed_at ON repositories(pushed_at DESC);

-- Add comments to explain the new fields
COMMENT ON COLUMN repositories.homepage IS 'Repository homepage URL from GitHub API';
COMMENT ON COLUMN repositories.language IS 'Primary programming language from GitHub API';
COMMENT ON COLUMN repositories.archived IS 'Whether the repository is archived (no longer maintained)';
COMMENT ON COLUMN repositories.default_branch IS 'Default branch name from GitHub API';
COMMENT ON COLUMN repositories.created_at_github IS 'Repository creation timestamp from GitHub API';
COMMENT ON COLUMN repositories.pushed_at IS 'Last push timestamp from GitHub API (indicates recent activity)';
COMMENT ON COLUMN repositories.updated_at_github IS 'Last update timestamp from GitHub API';`,
		Down: `
-- Remove indexes and new fields
DROP INDEX IF EXISTS idx_repositories_pushed_at;
DROP INDEX IF EXISTS idx_repositories_language;
DROP INDEX IF EXISTS idx_repositories_archived;
ALTER TABLE repositories 
DROP COLUMN IF EXISTS updated_at_github,
DROP COLUMN IF EXISTS pushed_at,
DROP COLUMN IF EXISTS created_at_github,
DROP COLUMN IF EXISTS default_branch,
DROP COLUMN IF EXISTS archived,
DROP COLUMN IF EXISTS language,
DROP COLUMN IF EXISTS homepage;`,
	},
	{
		ID:          20,
		Name:        "add_module_table",
		Description: "Adds the modules table to track info about modules",
		Up: `
CREATE TABLE modules (
	namespace VARCHAR(255) NOT NULL,
	name VARCHAR(255) NOT NULL,
	target VARCHAR(255) NOT NULL,
	
	repo_organisation VARCHAR(255) NOT NULL,
	repo_name VARCHAR(255) NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

	PRIMARY KEY (namespace, name, target),
	FOREIGN KEY (repo_organisation, repo_name)
		  REFERENCES repositories(organisation, name)
);
CREATE INDEX idx_modules_namespace ON modules(namespace);
CREATE INDEX idx_modules_name ON modules(name);
CREATE INDEX idx_modules_target ON modules(target);

-- add trigger for updated_at
CREATE TRIGGER update_modules_updated_at 
		BEFORE UPDATE ON modules 
		FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
COMMENT ON TABLE modules IS 'Stores metadata for Terraform modules';
COMMENT ON COLUMN modules.namespace IS 'Module namespace (usually GitHub organisation)';
COMMENT ON COLUMN modules.name IS 'Module name (the name of the item being managed by the repository)';
COMMENT ON COLUMN modules.target IS 'Module target (the platform or technology the module is targetting ie aws, azurerm, gcp)';
		`,
		Down: `
DROP TRIGGER IF EXISTS update_modules_updated_at ON modules;
DROP INDEX IF EXISTS idx_modules_target;
DROP INDEX IF EXISTS idx_modules_name;
DROP INDEX IF EXISTS idx_modules_namespace;
DROP TABLE modules;`,
	},
	{
		ID:          21,
		Name:        "create_module_versions_table",
		Description: "Create module_versions table to store version information for modules",
		Up: `
CREATE TABLE module_versions (
    id SERIAL PRIMARY KEY,
    module_namespace VARCHAR(255) NOT NULL,
    module_name VARCHAR(255) NOT NULL,
    module_target VARCHAR(255) NOT NULL,
    version VARCHAR(255) NOT NULL,
    tag_created_at TIMESTAMP WITH TIME ZONE,
    discovered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (module_namespace, module_name, module_target) 
        REFERENCES modules(namespace, name, target) ON DELETE CASCADE,
    UNIQUE(module_namespace, module_name, module_target, version)
);

CREATE INDEX idx_module_versions_module ON module_versions(module_namespace, module_name, module_target);
CREATE INDEX idx_module_versions_version ON module_versions(version);
CREATE INDEX idx_module_versions_tag_created_at ON module_versions(tag_created_at);
CREATE INDEX idx_module_versions_discovered_at ON module_versions(discovered_at);

-- Add trigger for updated_at
CREATE TRIGGER update_module_versions_updated_at 
    BEFORE UPDATE ON module_versions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE module_versions IS 'Stores version information for Terraform modules';
COMMENT ON COLUMN module_versions.version IS 'Semantic version string (e.g., v1.0.0)';
COMMENT ON COLUMN module_versions.tag_created_at IS 'When the git tag was created';
COMMENT ON COLUMN module_versions.discovered_at IS 'When this version was discovered by the registry';`,
		Down: `
DROP TRIGGER IF EXISTS update_module_versions_updated_at ON module_versions;
DROP INDEX IF EXISTS idx_module_versions_discovered_at;
DROP INDEX IF EXISTS idx_module_versions_tag_created_at;
DROP INDEX IF EXISTS idx_module_versions_version;
DROP INDEX IF EXISTS idx_module_versions_module;
DROP TABLE module_versions;`,
	},
	{
		ID:          22,
		Name:        "add_module_scraping_tables",
		Description: "Add tables for comprehensive module scraping including submodules, examples, and licenses",
		Up: `
-- Add processed_at and tofu_json to module_versions
ALTER TABLE module_versions 
ADD COLUMN processed_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN tofu_json JSONB;

-- Create module_version_licenses table (like provider_version_licenses)
CREATE TABLE module_version_licenses (
    id SERIAL PRIMARY KEY,
    module_namespace VARCHAR(255) NOT NULL,
    module_name VARCHAR(255) NOT NULL,
    module_target VARCHAR(255) NOT NULL,
    version VARCHAR(255) NOT NULL,
    license_spdx_id VARCHAR(255) NOT NULL,
    confidence_score DECIMAL(4,3),
    file_path TEXT,
    match_type VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (module_namespace, module_name, module_target, version) 
        REFERENCES module_versions(module_namespace, module_name, module_target, version) ON DELETE CASCADE,
    FOREIGN KEY (license_spdx_id) REFERENCES licenses(spdx_id) ON DELETE CASCADE,
    UNIQUE(module_namespace, module_name, module_target, version, license_spdx_id, file_path)
);

-- Create module_submodules table
CREATE TABLE module_submodules (
    id SERIAL PRIMARY KEY,
    module_namespace VARCHAR(255) NOT NULL,
    module_name VARCHAR(255) NOT NULL,
    module_target VARCHAR(255) NOT NULL,
    version VARCHAR(255) NOT NULL,
    submodule_name VARCHAR(255) NOT NULL,
    path VARCHAR(500) NOT NULL,
    tofu_json JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (module_namespace, module_name, module_target, version) 
        REFERENCES module_versions(module_namespace, module_name, module_target, version) ON DELETE CASCADE,
    UNIQUE(module_namespace, module_name, module_target, version, submodule_name)
);

-- Create module_examples table  
CREATE TABLE module_examples (
    id SERIAL PRIMARY KEY,
    module_namespace VARCHAR(255) NOT NULL,
    module_name VARCHAR(255) NOT NULL,
    module_target VARCHAR(255) NOT NULL,
    version VARCHAR(255) NOT NULL,
    example_name VARCHAR(255) NOT NULL,
    path VARCHAR(500) NOT NULL,
    tofu_json JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (module_namespace, module_name, module_target, version) 
        REFERENCES module_versions(module_namespace, module_name, module_target, version) ON DELETE CASCADE,
    UNIQUE(module_namespace, module_name, module_target, version, example_name)
);

-- Create indexes for new tables
CREATE INDEX idx_module_version_licenses_module ON module_version_licenses(module_namespace, module_name, module_target, version);
CREATE INDEX idx_module_version_licenses_license ON module_version_licenses(license_spdx_id);
CREATE INDEX idx_module_version_licenses_confidence ON module_version_licenses(confidence_score);

CREATE INDEX idx_module_submodules_module ON module_submodules(module_namespace, module_name, module_target, version);
CREATE INDEX idx_module_submodules_name ON module_submodules(submodule_name);

CREATE INDEX idx_module_examples_module ON module_examples(module_namespace, module_name, module_target, version);
CREATE INDEX idx_module_examples_name ON module_examples(example_name);

-- Add indexes for JSONB fields for better query performance
CREATE INDEX idx_module_versions_tofu_json ON module_versions USING GIN(tofu_json);
CREATE INDEX idx_module_submodules_tofu_json ON module_submodules USING GIN(tofu_json);
CREATE INDEX idx_module_examples_tofu_json ON module_examples USING GIN(tofu_json);

-- Add comments
COMMENT ON TABLE module_version_licenses IS 'Stores detailed license information for module versions';
COMMENT ON TABLE module_submodules IS 'Stores information about submodules within modules';
COMMENT ON TABLE module_examples IS 'Stores information about examples within modules';
COMMENT ON COLUMN module_versions.processed_at IS 'When this version was processed and indexed';
COMMENT ON COLUMN module_versions.tofu_json IS 'Raw JSON output from tofu show command for the root module';`,
		Down: `
-- Drop indexes
DROP INDEX IF EXISTS idx_module_examples_tofu_json;
DROP INDEX IF EXISTS idx_module_submodules_tofu_json;
DROP INDEX IF EXISTS idx_module_versions_tofu_json;

DROP INDEX IF EXISTS idx_module_examples_name;
DROP INDEX IF EXISTS idx_module_examples_module;
DROP INDEX IF EXISTS idx_module_submodules_name;
DROP INDEX IF EXISTS idx_module_submodules_module;
DROP INDEX IF EXISTS idx_module_version_licenses_confidence;
DROP INDEX IF EXISTS idx_module_version_licenses_license;
DROP INDEX IF EXISTS idx_module_version_licenses_module;

-- Drop tables
DROP TABLE IF EXISTS module_examples;
DROP TABLE IF EXISTS module_submodules;
DROP TABLE IF EXISTS module_version_licenses;

-- Remove columns from module_versions
ALTER TABLE module_versions 
DROP COLUMN IF EXISTS tofu_json,
DROP COLUMN IF EXISTS processed_at;`,
	},
	{
		ID:          23,
		Name:        "add_missing_comments",
		Description: "Add missing table and column comments for documentation",
		Up: `
-- repositories table comments
COMMENT ON TABLE repositories IS 'Github repositories table that serve up providers and modules';
COMMENT ON COLUMN repositories.organisation IS 'GitHub org';
COMMENT ON COLUMN repositories.name IS 'GitHub repo name';
COMMENT ON COLUMN repositories.is_fork IS 'If the repo is a fork of another or not';
COMMENT ON COLUMN repositories.parent_organisation IS 'The organization of the forked project';
COMMENT ON COLUMN repositories.parent_name IS 'The repo name of the forked proejct';
COMMENT ON COLUMN repositories.created_at IS ' when this item was created';
COMMENT ON COLUMN repositories.updated_at IS 'when this item was updated';
COMMENT ON COLUMN repositories.description IS 'The github description of this project.';

-- providers table comments
COMMENT ON TABLE providers IS 'Stores information about providers known to the OpenTofu registry';
COMMENT ON COLUMN providers.namespace IS 'the namespace of the provider, similar to GitHub org today';
COMMENT ON COLUMN providers.name IS 'the name of the provider, taken from the github org name, ie terraform-provider-myname becomes myname';
COMMENT ON COLUMN providers.alias IS 'the namespace/name of the provider which we are aliasing here';
COMMENT ON COLUMN providers.warnings IS 'Any human written warnings about this provider. These warnings will be displayed when someone calls tofu init and are usually to tell people about deprecation or incorrect namespacing/aliasing';
COMMENT ON COLUMN providers.repo_organisation IS 'the github org the provider is hosted on';
COMMENT ON COLUMN providers.repo_name IS 'the github repo name the provider is hosted from';
COMMENT ON COLUMN providers.created_at IS 'when this item was created';
COMMENT ON COLUMN providers.updated_at IS 'when this item was updated last';

-- provider_versions table comments
COMMENT ON TABLE provider_versions IS 'Stores information about individual provider versions';
COMMENT ON COLUMN provider_versions.id IS 'id';
COMMENT ON COLUMN provider_versions.provider_namespace IS 'the namespace of the provider, similar to GitHub org today';
COMMENT ON COLUMN provider_versions.provider_name IS 'the name of the provider, taken from the github org name, ie terraform-provider-myname becomes myname';
COMMENT ON COLUMN provider_versions.version IS 'the version of the provider, this normally corrolates to a github release AND tag';
COMMENT ON COLUMN provider_versions.tag_created_at IS 'when the github tag was created';
COMMENT ON COLUMN provider_versions.discovered_at IS 'when we discovered this provider version';
COMMENT ON COLUMN provider_versions.created_at IS 'wen this item was created';
COMMENT ON COLUMN provider_versions.updated_at IS 'when this item was last updated';
COMMENT ON COLUMN provider_versions.license_accepted IS 'if we accept the license or not, for more info you should see the provider_version_licenses table -- TODO: Maybe remove this for a provider_version_licenses only lookup';

-- licenses table comments
COMMENT ON TABLE licenses IS 'Stores information about different licenses and their properties';
COMMENT ON COLUMN licenses.spdx_id IS 'spdx-identifier';
COMMENT ON COLUMN licenses.name IS 'human-readable-name';
COMMENT ON COLUMN licenses.redistributable IS 'If we deem this license to be redistributable or not';
COMMENT ON COLUMN licenses.created_at IS 'when this item was created';
COMMENT ON COLUMN licenses.updated_at IS 'when this item was updated';

-- provider_version_licenses table comments
COMMENT ON TABLE provider_version_licenses IS 'Stores information about which licenses a specific provider version uses';
COMMENT ON COLUMN provider_version_licenses.id IS 'id';
COMMENT ON COLUMN provider_version_licenses.provider_namespace IS 'the namespace of the provider';
COMMENT ON COLUMN provider_version_licenses.provider_name IS 'the name of the provider';
COMMENT ON COLUMN provider_version_licenses.version IS 'the version of the provider';
COMMENT ON COLUMN provider_version_licenses.license_spdx_id IS 'the spdx-identifier of the license';
COMMENT ON COLUMN provider_version_licenses.confidence_score IS 'The confidence score of the license detection';
COMMENT ON COLUMN provider_version_licenses.file_path IS 'path to the file in the git repo';
COMMENT ON COLUMN provider_version_licenses.match_type IS 'If this was detected using license detection or GitHubs API';
COMMENT ON COLUMN provider_version_licenses.created_at IS 'When this item was created';

-- provider_aliases table comments
COMMENT ON TABLE provider_aliases IS 'Stores information about aliases for providers';
COMMENT ON COLUMN provider_aliases.id IS 'id';
COMMENT ON COLUMN provider_aliases.original_namespace IS 'original namespace of the provider before alias';
COMMENT ON COLUMN provider_aliases.original_name IS 'original name of the provider before alias';
COMMENT ON COLUMN provider_aliases.target_namespace IS 'new namespace of the provider after alias';
COMMENT ON COLUMN provider_aliases.target_name IS 'new name of the provider after alias';
COMMENT ON COLUMN provider_aliases.created_at IS 'when this item was created';
COMMENT ON COLUMN provider_aliases.updated_at IS 'when this item was last updated';

-- repository_stats table comments
COMMENT ON TABLE repository_stats IS 'Stores info about the github repo behind the provider/module';
COMMENT ON COLUMN repository_stats.repo_organisation IS 'id';
COMMENT ON COLUMN repository_stats.repo_name IS 'name of the github repo';
COMMENT ON COLUMN repository_stats.stars IS 'count of GitHub stars';
COMMENT ON COLUMN repository_stats.forks IS 'count of GitHub forks of this repository';
COMMENT ON COLUMN repository_stats.watchers IS 'count of GitHub watchers';
COMMENT ON COLUMN repository_stats.open_issues IS 'count of GitHub open issues';

-- repository_redirects table comments
COMMENT ON TABLE repository_redirects IS 'Stores info about the github repo redirect, for situations where repos are renamed, or orgs redirect to new org names';
COMMENT ON COLUMN repository_redirects.from_organisation IS 'the org we are redirecting from';
COMMENT ON COLUMN repository_redirects.from_name IS 'the repo name we are redirecting from';
COMMENT ON COLUMN repository_redirects.to_organisation IS 'the org we are redirecting to';
COMMENT ON COLUMN repository_redirects.to_name IS 'the repo name we are redirecting to';
COMMENT ON COLUMN repository_redirects.created_at IS 'when this item was created';
COMMENT ON COLUMN repository_redirects.updated_at IS 'when this item was last updated';

-- modules table comments
COMMENT ON TABLE modules IS 'Stores information about modules known in the OpenTofu registry';
COMMENT ON COLUMN modules.namespace IS 'namespace of the module, similar to github org';
COMMENT ON COLUMN modules.name IS 'name of the module';
COMMENT ON COLUMN modules.target IS 'target system for the module';
COMMENT ON COLUMN modules.created_at IS 'when this item was created';
COMMENT ON COLUMN modules.updated_at IS 'when this item was last updated';
COMMENT ON COLUMN modules.repo_organisation IS 'organization of the github repo for this module';
COMMENT ON COLUMN modules.repo_name IS 'the name of the github repo for this module';

-- module_versions table comments
COMMENT ON COLUMN module_versions.id IS '';
COMMENT ON COLUMN module_versions.module_namespace IS 'the namespace of the module, similar to github org';
COMMENT ON COLUMN module_versions.module_name IS ' the name of the module';
COMMENT ON COLUMN module_versions.module_target IS 'the target system for the module';
COMMENT ON COLUMN module_versions.version IS 'the version of the module';
COMMENT ON COLUMN module_versions.discovered_at IS 'when this item was discovered by this repository';
COMMENT ON COLUMN module_versions.tofu_json IS 'Raw JSON output from tofu show command for the root module';

-- module_version_licenses table comments
COMMENT ON TABLE module_version_licenses IS 'Stores information about licenses that have been found for specific module versions';
COMMENT ON COLUMN module_version_licenses.id IS 'id';
COMMENT ON COLUMN module_version_licenses.module_namespace IS 'namespace of the module';
COMMENT ON COLUMN module_version_licenses.module_name IS 'name of the module';
COMMENT ON COLUMN module_version_licenses.module_target IS 'target system of the module';
COMMENT ON COLUMN module_version_licenses.version IS 'version of the module';
COMMENT ON COLUMN module_version_licenses.license_spdx_id IS 'the spdx-identifiier we determined for this module version';
COMMENT ON COLUMN module_version_licenses.confidence_score IS 'The confidence score of the license detection';
COMMENT ON COLUMN module_version_licenses.file_path IS 'the file path to the license in the git repo';
COMMENT ON COLUMN module_version_licenses.match_type IS 'if this license was detected by us, or if we fell back to the GitHub API';
COMMENT ON COLUMN module_version_licenses.created_at IS 'when this item was created in our db (Not when the license was created)';

-- module_submodules table comments
COMMENT ON TABLE module_submodules IS 'stores info about submodules from modules';
COMMENT ON COLUMN module_submodules.id IS 'id';
COMMENT ON COLUMN module_submodules.module_namespace IS 'the namespace of the module, similar to github org';
COMMENT ON COLUMN module_submodules.module_name IS ' the name of the module';
COMMENT ON COLUMN module_submodules.module_target IS 'the target system for the module';
COMMENT ON COLUMN module_submodules.version IS 'the version of the module';
COMMENT ON COLUMN module_submodules.submodule_name IS 'the name of the submodule';
COMMENT ON COLUMN module_submodules.path IS 'the path to the submodule in the git repo';
COMMENT ON COLUMN module_submodules.tofu_json IS 'The raw output of tofu config show run in this submodule directory';
COMMENT ON COLUMN module_submodules.created_at IS 'when this item was created in our db (Not when the submodule was created)';

-- module_examples table comments
COMMENT ON TABLE module_examples IS 'Stores information about module examples';
COMMENT ON COLUMN module_examples.id IS 'id';
COMMENT ON COLUMN module_examples.module_namespace IS 'the namespace of the module, similar to github org';
COMMENT ON COLUMN module_examples.module_name IS ' the name of the module';
COMMENT ON COLUMN module_examples.module_target IS 'the target system for the module';
COMMENT ON COLUMN module_examples.version IS 'the version of the module';
COMMENT ON COLUMN module_examples.example_name IS 'the name of the example';
COMMENT ON COLUMN module_examples.path IS 'the path to the example in the git repo';
COMMENT ON COLUMN module_examples.tofu_json IS 'the raw output of tofu config show run in this example directory';
COMMENT ON COLUMN module_examples.created_at IS 'when this item was created in our db (Not when the example was created)';`,
		Down: `-- Note: Removing comments doesn't affect functionality, no need to include it all here`,
	},
	{
		ID:          24,
		Name:        "add_missing_indices",
		Description: "Add missing foreign key and performance indices",
		Up: `
-- Foreign key indices for providers table
CREATE INDEX idx_providers_repository ON providers(repo_organisation, repo_name);
-- Foreign key indices for provider_versions table
CREATE INDEX idx_provider_versions_provider ON provider_versions(provider_namespace, provider_name);
-- Foreign key indices for modules table
CREATE INDEX idx_modules_repository ON modules(repo_organisation, repo_name);
-- Foreign key indices for module_versions table
CREATE INDEX idx_module_versions_composite ON module_versions(module_namespace, module_name, module_target);
-- Additional performance indices for repositories
CREATE INDEX idx_repositories_created_at ON repositories(created_at);

-- Performance indices for version lookups
CREATE INDEX idx_provider_versions_composite ON provider_versions(provider_namespace, provider_name, version);
CREATE INDEX idx_module_versions_full_composite ON module_versions(module_namespace, module_name, module_target, version);

-- Index for repository stats time-series queries about repo stats -- TODO: review, not sure about this being useful yet
CREATE INDEX idx_repository_stats_composite ON repository_stats(repo_organisation, repo_name, stars DESC, recorded_at DESC);`,
		Down: `
-- Remove performance indices
DROP INDEX IF EXISTS idx_repository_stats_composite;
DROP INDEX IF EXISTS idx_module_versions_full_composite;
DROP INDEX IF EXISTS idx_provider_versions_composite;
DROP INDEX IF EXISTS idx_repositories_created_at;

-- Remove foreign key indices
DROP INDEX IF EXISTS idx_module_versions_composite;
DROP INDEX IF EXISTS idx_modules_repository;
DROP INDEX IF EXISTS idx_provider_versions_provider;
DROP INDEX IF EXISTS idx_providers_repository;`,
	},
	{
		ID:          25,
		Name:        "remove_unused_columns",
		Description: "Remove category and url from licenses table, remove tag_created_at from module_versions table",
		Up: `

ALTER TABLE licenses
DROP COLUMN IF EXISTS category,
DROP COLUMN IF EXISTS url;

-- Remove unused column from module_versions table
ALTER TABLE module_versions
DROP COLUMN IF EXISTS tag_created_at;`,
		Down: `
-- Restore columns to licenses table
ALTER TABLE licenses
ADD COLUMN category VARCHAR(50),
ADD COLUMN url VARCHAR(500);

-- Restore index for licenses.category
CREATE INDEX idx_licenses_category ON licenses(category);

-- Restore column to module_versions table
ALTER TABLE module_versions
ADD COLUMN tag_created_at TIMESTAMP WITH TIME ZONE;

-- Restore index for module_versions.tag_created_at
CREATE INDEX idx_module_versions_tag_created_at ON module_versions(tag_created_at);`,
	},
	{
		ID:          26,
		Name:        "add_document_types_for_actions_and_list_resources",
		Description: "Add document types for provider actions and list-resources to support docs for these new provider features",
		Up: `
	ALTER TABLE provider_documents
	  DROP CONSTRAINT IF EXISTS provider_documents_document_type_check;

	ALTER TABLE provider_documents
	  ADD CONSTRAINT provider_documents_document_type_check
	  CHECK (document_type IN ('resources', 'datasources', 'functions', 'guides', 'ephemeral', 'index', 'actions', 'listresources'));
	`,
		Down: `
	ALTER TABLE provider_documents
	  DROP CONSTRAINT IF EXISTS provider_documents_document_type_check;

	ALTER TABLE provider_documents
	  ADD CONSTRAINT provider_documents_document_type_check
	  CHECK (document_type IN ('resources', 'datasources', 'functions', 'guides', 'ephemeral', 'index'));
	`,
	},
	{
		ID:          27,
		Name:        "restore_tag_created_at_to_module_versions",
		Description: "Restore tag_created_at column to module_versions table - needed by indexer to track when module version git tags were created for accurate version history display",
		Up: `
-- Restore tag_created_at column that was removed in migration 25
-- This column is actively used by the module indexer to store git tag creation dates
ALTER TABLE module_versions
ADD COLUMN IF NOT EXISTS tag_created_at TIMESTAMP WITH TIME ZONE;

-- Restore index for efficient querying by tag creation date
CREATE INDEX IF NOT EXISTS idx_module_versions_tag_created_at ON module_versions(tag_created_at);

COMMENT ON COLUMN module_versions.tag_created_at IS 'When the git tag for this module version was created in the source repository';`,
		Down: `
-- Remove index before dropping column
DROP INDEX IF EXISTS idx_module_versions_tag_created_at;

-- Remove tag_created_at column
ALTER TABLE module_versions
DROP COLUMN IF EXISTS tag_created_at;`,
	},
	{
		ID:          28,
		Name:        "restore_category_and_url_to_licenses",
		Description: "Restore category and url columns to licenses table - needed by module storage to categorize detected licenses and store reference URLs",
		Up: `
-- Restore category and url columns that were removed in migration 25
-- These columns are actively used by module license storage to track license metadata
ALTER TABLE licenses
ADD COLUMN IF NOT EXISTS category VARCHAR(50),
ADD COLUMN IF NOT EXISTS url VARCHAR(500);

-- Restore index for efficient querying by category
CREATE INDEX IF NOT EXISTS idx_licenses_category ON licenses(category);

COMMENT ON COLUMN licenses.category IS 'Category of the license (e.g., detected, manual, inherited)';
COMMENT ON COLUMN licenses.url IS 'Reference URL for more information about the license';`,
		Down: `
-- Remove index before dropping columns
DROP INDEX IF EXISTS idx_licenses_category;

-- Remove category and url columns
ALTER TABLE licenses
DROP COLUMN IF EXISTS category,
DROP COLUMN IF EXISTS url;`,
	},
	{
		ID:          29,
		Name:        "add_version_scrape_status_tracking",
		Description: "Add scraping status tracking fields to provider_versions and module_versions tables to support skip functionality, error tracking, and manual version management for both license issues and processing failures",
		Up: `
-- Add scraping status tracking to provider_versions
ALTER TABLE provider_versions
ADD COLUMN IF NOT EXISTS scrape_status VARCHAR(50) DEFAULT 'completed' NOT NULL,
ADD COLUMN IF NOT EXISTS skip_reason VARCHAR(100),
ADD COLUMN IF NOT EXISTS error_message TEXT,
ADD COLUMN IF NOT EXISTS last_attempt_at TIMESTAMP WITH TIME ZONE;

-- Add scraping status tracking to module_versions
ALTER TABLE module_versions
ADD COLUMN IF NOT EXISTS scrape_status VARCHAR(50) DEFAULT 'completed' NOT NULL,
ADD COLUMN IF NOT EXISTS skip_reason VARCHAR(100),
ADD COLUMN IF NOT EXISTS error_message TEXT,
ADD COLUMN IF NOT EXISTS last_attempt_at TIMESTAMP WITH TIME ZONE;

-- Create indexes for efficient status filtering
CREATE INDEX IF NOT EXISTS idx_provider_versions_status ON provider_versions(scrape_status);
CREATE INDEX IF NOT EXISTS idx_module_versions_status ON module_versions(scrape_status);

-- Migrate existing provider versions with incompatible licenses to skipped status
UPDATE provider_versions
SET scrape_status = 'skipped',
    skip_reason = 'incompatible_license'
WHERE license_accepted = false;

-- Document the new columns
COMMENT ON COLUMN provider_versions.scrape_status IS 'Scraping status: completed (successful), skipped (intentionally excluded), failed (processing error)';
COMMENT ON COLUMN provider_versions.skip_reason IS 'Reason for skip: incompatible_license, no_license, processing_error, manual_skip, malformed_data';
COMMENT ON COLUMN provider_versions.error_message IS 'Detailed error message if scraping failed';
COMMENT ON COLUMN provider_versions.last_attempt_at IS 'Timestamp of the last scraping attempt for this version';

COMMENT ON COLUMN module_versions.scrape_status IS 'Scraping status: completed (successful), skipped (intentionally excluded), failed (processing error)';
COMMENT ON COLUMN module_versions.skip_reason IS 'Reason for skip: incompatible_license, no_license, processing_error, manual_skip, malformed_data';
COMMENT ON COLUMN module_versions.error_message IS 'Detailed error message if scraping failed';
COMMENT ON COLUMN module_versions.last_attempt_at IS 'Timestamp of the last scraping attempt for this version';`,
		Down: `
-- Remove indexes
DROP INDEX IF EXISTS idx_module_versions_status;
DROP INDEX IF EXISTS idx_provider_versions_status;

-- Remove columns from module_versions
ALTER TABLE module_versions
DROP COLUMN IF EXISTS last_attempt_at,
DROP COLUMN IF EXISTS error_message,
DROP COLUMN IF EXISTS skip_reason,
DROP COLUMN IF EXISTS scrape_status;

-- Remove columns from provider_versions
ALTER TABLE provider_versions
DROP COLUMN IF EXISTS last_attempt_at,
DROP COLUMN IF EXISTS error_message,
DROP COLUMN IF EXISTS skip_reason,
DROP COLUMN IF EXISTS scrape_status;`,
	},
	{
		ID:          30,
		Name:        "add_md5_checksums_to_s3_documents",
		Description: "Add MD5 checksum columns to track S3 file integrity and enable content-based deduplication",
		Up: `
-- Add MD5 checksum to provider documents
ALTER TABLE provider_documents
ADD COLUMN IF NOT EXISTS md5_checksum VARCHAR(32);

CREATE INDEX IF NOT EXISTS idx_provider_documents_checksum ON provider_documents(md5_checksum);

-- Add MD5 checksums to module versions (for index.json and README.md)
ALTER TABLE module_versions
ADD COLUMN IF NOT EXISTS index_md5_checksum VARCHAR(32),
ADD COLUMN IF NOT EXISTS readme_md5_checksum VARCHAR(32);

-- Add MD5 checksums to module submodules (for index.json and README.md)
ALTER TABLE module_submodules
ADD COLUMN IF NOT EXISTS index_md5_checksum VARCHAR(32),
ADD COLUMN IF NOT EXISTS readme_md5_checksum VARCHAR(32);

-- Add MD5 checksums to module examples (for index.json and README.md)
ALTER TABLE module_examples
ADD COLUMN IF NOT EXISTS index_md5_checksum VARCHAR(32),
ADD COLUMN IF NOT EXISTS readme_md5_checksum VARCHAR(32);

-- Document the new columns
COMMENT ON COLUMN provider_documents.md5_checksum IS 'MD5 checksum of the document content stored in S3 for integrity verification and deduplication';
COMMENT ON COLUMN module_versions.index_md5_checksum IS 'MD5 checksum of the module index.json file stored in S3';
COMMENT ON COLUMN module_versions.readme_md5_checksum IS 'MD5 checksum of the module README.md file stored in S3';
COMMENT ON COLUMN module_submodules.index_md5_checksum IS 'MD5 checksum of the submodule index.json file stored in S3';
COMMENT ON COLUMN module_submodules.readme_md5_checksum IS 'MD5 checksum of the submodule README.md file stored in S3';
COMMENT ON COLUMN module_examples.index_md5_checksum IS 'MD5 checksum of the example index.json file stored in S3';
COMMENT ON COLUMN module_examples.readme_md5_checksum IS 'MD5 checksum of the example README.md file stored in S3';`,
		Down: `
-- Remove index
DROP INDEX IF EXISTS idx_provider_documents_checksum;

-- Remove checksum columns from provider_documents
ALTER TABLE provider_documents
DROP COLUMN IF EXISTS md5_checksum;

-- Remove checksum columns from module_versions
ALTER TABLE module_versions
DROP COLUMN IF EXISTS readme_md5_checksum,
DROP COLUMN IF EXISTS index_md5_checksum;

-- Remove checksum columns from module_submodules
ALTER TABLE module_submodules
DROP COLUMN IF EXISTS readme_md5_checksum,
DROP COLUMN IF EXISTS index_md5_checksum;

-- Remove checksum columns from module_examples
ALTER TABLE module_examples
DROP COLUMN IF EXISTS readme_md5_checksum,
DROP COLUMN IF EXISTS index_md5_checksum;`,
	},
	{
		ID:          31,
		Name:        "enable_semver_extension",
		Description: "Enable PostgreSQL semver extension for proper semantic version sorting in global indexes",
		Up: `
-- Enable semver extension (supported on Neon)
CREATE EXTENSION IF NOT EXISTS semver;

COMMENT ON EXTENSION semver IS 'Enables semantic version data type for proper version sorting';`,
		Down: `
-- Remove semver extension
DROP EXTENSION IF EXISTS semver;`,
	},
	{
		ID:          32,
		Name:        "add_safe_to_semver_function",
		Description: "Add helper function for normalizing version strings before semver sorting",
		Up: `
-- Helper function to normalize version strings for semver sorting
-- Strips 'v' prefix, handles partial versions, and rejects invalid versions
CREATE OR REPLACE FUNCTION safe_to_semver(v TEXT) RETURNS semver AS $$
BEGIN
    -- Strip leading 'v' if present
    v := regexp_replace(v, '^v', '');

    -- Reject versions with numbers too large for semver (max 2147483647)
    -- Check for any number with 10+ digits
    IF v ~ '\d{10,}' THEN
        RETURN '0.0.0'::semver;
    END IF;

    -- Pad partial versions: "1" -> "1.0.0", "1.0" -> "1.0.0"
    IF v ~ '^\d+$' THEN
        v := v || '.0.0';
    ELSIF v ~ '^\d+\.\d+$' THEN
        v := v || '.0';
    END IF;

    RETURN v::semver;
EXCEPTION
    WHEN OTHERS THEN
        -- Invalid versions sort to the bottom
        RETURN '0.0.0'::semver;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

COMMENT ON FUNCTION safe_to_semver(TEXT) IS 'Normalizes version strings (strips v prefix, pads partial versions) and safely casts to semver';`,
		Down: `
DROP FUNCTION IF EXISTS safe_to_semver(TEXT);`,
	},
}

func NewMigrateCommand() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Run database migrations",
		Commands: []*cli.Command{
			{
				Name:  "up",
				Usage: "Run all pending migrations",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					dbConfig := getDBConfig(cmd)
					pool, err := dbConfig.GetPool(ctx)
					if err != nil {
						return fmt.Errorf("failed to connect to database: %w", err)
					}
					defer pool.Close()
					return migrateUp(ctx, pool)
				},
			},
			{
				Name:  "down",
				Usage: "Rollback all migrations",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					dbConfig := getDBConfig(cmd)
					pool, err := dbConfig.GetPool(ctx)
					if err != nil {
						return fmt.Errorf("failed to connect to database: %w", err)
					}
					defer pool.Close()
					return migrateDown(ctx, pool)
				},
			},
			{
				Name:  "reset",
				Usage: "Rollback all migrations and run them again",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					dbConfig := getDBConfig(cmd)
					pool, err := dbConfig.GetPool(ctx)
					if err != nil {
						return fmt.Errorf("failed to connect to database: %w", err)
					}
					defer pool.Close()
					if err := migrateDown(ctx, pool); err != nil {
						return err
					}

					return migrateUp(ctx, pool)
				},
			},
			{
				Name:  "version",
				Usage: "Show current migration version",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					dbConfig := getDBConfig(cmd)
					pool, err := dbConfig.GetPool(ctx)
					if err != nil {
						return fmt.Errorf("failed to connect to database: %w", err)
					}
					defer pool.Close()
					return showVersion(ctx, pool)
				},
			},
		},
	}
}

func createMigrationsTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version bigint NOT NULL,
			dirty boolean NOT NULL,
			PRIMARY KEY (version)
		)
	`)
	return err
}

func getCurrentVersion(ctx context.Context, pool *pgxpool.Pool) (int, bool, error) {
	var version int
	var dirty bool

	err := pool.QueryRow(ctx, "SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&version, &dirty)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}

	slog.InfoContext(ctx, "Latest migration ID identified", "version", version)

	return version, dirty, nil
}

func migrateUp(ctx context.Context, pool *pgxpool.Pool) error {
	slog.InfoContext(ctx, "Running migrations up")

	if err := createMigrationsTable(ctx, pool); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	currentVersion, dirty, err := getCurrentVersion(ctx, pool)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in dirty state, manual intervention required")
	}

	return runMigrationsUp(ctx, pool, currentVersion)
}

func runMigrationsUp(ctx context.Context, pool *pgxpool.Pool, currentVersion int) error {
	applied := 0
	for _, migration := range migrations {
		if migration.ID > currentVersion {
			migrationCtx, span := telemetry.Tracer().Start(ctx, fmt.Sprintf("migration.up.%03d_%s", migration.ID, migration.Name))

			slog.InfoContext(migrationCtx, "Applying migration", "id", migration.ID, "name", migration.Name, "description", migration.Description)
			span.SetAttributes(
				attribute.Int("migration.id", migration.ID),
				attribute.String("migration.name", migration.Name),
				attribute.String("migration.description", migration.Description),
			)

			tx, err := pool.Begin(migrationCtx)
			if err != nil {
				span.RecordError(err)
				span.End()
				return fmt.Errorf("failed to begin transaction: %w", err)
			}

			// Set dirty flag
			if _, err := tx.Exec(migrationCtx, "INSERT INTO schema_migrations (version, dirty) VALUES ($1, true) ON CONFLICT (version) DO UPDATE SET dirty = true", migration.ID); err != nil {
				tx.Rollback(ctx)
				span.RecordError(err)
				span.End()
				return fmt.Errorf("failed to set dirty flag: %w", err)
			}

			// Run migration
			if _, err := tx.Exec(migrationCtx, migration.Up); err != nil {
				tx.Rollback(ctx)
				span.RecordError(err)
				span.End()
				return fmt.Errorf("migration %d failed: %w", migration.ID, err)
			}

			// Clear dirty flag
			if _, err := tx.Exec(migrationCtx, "UPDATE schema_migrations SET dirty = false WHERE version = $1", migration.ID); err != nil {
				tx.Rollback(ctx)
				span.RecordError(err)
				span.End()
				return fmt.Errorf("failed to clear dirty flag: %w", err)
			}

			if err := tx.Commit(migrationCtx); err != nil {
				span.RecordError(err)
				span.End()
				return fmt.Errorf("failed to commit transaction: %w", err)
			}

			span.End()
			applied++
		}
	}

	if applied == 0 {
		slog.Info("No migrations to apply")
	} else {
		slog.Info("Successfully applied migrations", "count", applied)
	}

	return nil
}

func migrateDown(ctx context.Context, pool *pgxpool.Pool) error {
	slog.InfoContext(ctx, "Running migrations down")

	if err := createMigrationsTable(ctx, pool); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	currentVersion, dirty, err := getCurrentVersion(ctx, pool)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in dirty state, manual intervention required")
	}

	return runMigrationsDown(ctx, pool, currentVersion)
}

func runMigrationsDown(ctx context.Context, pool *pgxpool.Pool, currentVersion int) error {
	rolled := 0
	// Run migrations in reverse order
	for i := len(migrations) - 1; i >= 0; i-- {
		migration := migrations[i]
		if migration.ID <= currentVersion {
			migrationCtx, span := telemetry.Tracer().Start(ctx, fmt.Sprintf("migration.down.%03d_%s", migration.ID, migration.Name))

			slog.InfoContext(migrationCtx, "Rolling back migration", "id", migration.ID, "name", migration.Name, "description", migration.Description)
			span.SetAttributes(
				attribute.Int("migration.id", migration.ID),
				attribute.String("migration.name", migration.Name),
				attribute.String("migration.description", migration.Description),
			)

			tx, err := pool.Begin(migrationCtx)
			if err != nil {
				span.RecordError(err)
				span.End()
				return fmt.Errorf("failed to begin transaction: %w", err)
			}

			// Set dirty flag
			if _, err := tx.Exec(migrationCtx, "UPDATE schema_migrations SET dirty = true WHERE version = $1", migration.ID); err != nil {
				tx.Rollback(ctx)
				span.RecordError(err)
				span.End()
				return fmt.Errorf("failed to set dirty flag: %w", err)
			}

			// Run rollback
			if _, err := tx.Exec(migrationCtx, migration.Down); err != nil {
				tx.Rollback(ctx)
				span.RecordError(err)
				span.End()
				return fmt.Errorf("rollback %d failed: %w", migration.ID, err)
			}

			// Remove migration record
			if _, err := tx.Exec(migrationCtx, "DELETE FROM schema_migrations WHERE version = $1", migration.ID); err != nil {
				tx.Rollback(ctx)
				span.RecordError(err)
				span.End()
				return fmt.Errorf("failed to remove migration record: %w", err)
			}

			if err := tx.Commit(migrationCtx); err != nil {
				span.RecordError(err)
				span.End()
				return fmt.Errorf("failed to commit transaction: %w", err)
			}

			span.End()
			rolled++
		}
	}

	if rolled == 0 {
		slog.Info("No migrations to rollback")
	} else {
		slog.Info("Successfully rolled back migrations", "count", rolled)
	}

	return nil
}

func showVersion(ctx context.Context, pool *pgxpool.Pool) error {
	if err := createMigrationsTable(ctx, pool); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	version, dirty, err := getCurrentVersion(ctx, pool)
	if err != nil {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	status := "clean"
	if dirty {
		status = "dirty"
	}

	slog.InfoContext(ctx, "Migration version", "version", version, "status", status)
	fmt.Printf("Current migration version: %d (status: %s)\n", version, status)

	return nil
}
