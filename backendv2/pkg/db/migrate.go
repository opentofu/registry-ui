package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/opentofu/registry-ui/pkg/config"
)

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
}

func NewMigrateCommand(config config.DBConfig) *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Run database migrations",
		Commands: []*cli.Command{
			{
				Name:  "up",
				Usage: "Run all pending migrations",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					pool, err := config.GetPool(ctx)
					if err != nil {
						return fmt.Errorf("failed to connect to database: %w", err)
					}
					return migrateUp(ctx, pool)
				},
			},
			{
				Name:  "down",
				Usage: "Rollback all migrations",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					pool, err := config.GetPool(ctx)
					if err != nil {
						return fmt.Errorf("failed to connect to database: %w", err)
					}
					return migrateDown(ctx, pool)
				},
			},
			{
				Name:  "reset",
				Usage: "Rollback all migrations and run them again",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					pool, err := config.GetPool(ctx)
					if err != nil {
						return fmt.Errorf("failed to connect to database: %w", err)
					}
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
					pool, err := config.GetPool(ctx)
					if err != nil {
						return fmt.Errorf("failed to connect to database: %w", err)
					}
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
	tracer := otel.Tracer("opentofu-registry-backend")

	applied := 0
	for _, migration := range migrations {
		if migration.ID > currentVersion {
			ctx, span := tracer.Start(ctx, fmt.Sprintf("migration.up.%03d_%s", migration.ID, migration.Name))

			slog.InfoContext(ctx, "Applying migration", "id", migration.ID, "name", migration.Name, "description", migration.Description)
			span.SetAttributes(
				attribute.Int("migration.id", migration.ID),
				attribute.String("migration.name", migration.Name),
				attribute.String("migration.description", migration.Description),
			)

			tx, err := pool.Begin(ctx)
			if err != nil {
				span.RecordError(err)
				span.End()
				return fmt.Errorf("failed to begin transaction: %w", err)
			}

			// Set dirty flag
			if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version, dirty) VALUES ($1, true) ON CONFLICT (version) DO UPDATE SET dirty = true", migration.ID); err != nil {
				tx.Rollback(ctx)
				span.RecordError(err)
				span.End()
				return fmt.Errorf("failed to set dirty flag: %w", err)
			}

			// Run migration
			if _, err := tx.Exec(ctx, migration.Up); err != nil {
				tx.Rollback(ctx)
				span.RecordError(err)
				span.End()
				return fmt.Errorf("migration %d failed: %w", migration.ID, err)
			}

			// Clear dirty flag
			if _, err := tx.Exec(ctx, "UPDATE schema_migrations SET dirty = false WHERE version = $1", migration.ID); err != nil {
				tx.Rollback(ctx)
				span.RecordError(err)
				span.End()
				return fmt.Errorf("failed to clear dirty flag: %w", err)
			}

			if err := tx.Commit(ctx); err != nil {
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
	tracer := otel.Tracer("opentofu-registry-backend")

	rolled := 0
	// Run migrations in reverse order
	for i := len(migrations) - 1; i >= 0; i-- {
		migration := migrations[i]
		if migration.ID <= currentVersion {
			ctx, span := tracer.Start(ctx, fmt.Sprintf("migration.down.%03d_%s", migration.ID, migration.Name))

			slog.InfoContext(ctx, "Rolling back migration", "id", migration.ID, "name", migration.Name, "description", migration.Description)
			span.SetAttributes(
				attribute.Int("migration.id", migration.ID),
				attribute.String("migration.name", migration.Name),
				attribute.String("migration.description", migration.Description),
			)

			tx, err := pool.Begin(ctx)
			if err != nil {
				span.RecordError(err)
				span.End()
				return fmt.Errorf("failed to begin transaction: %w", err)
			}

			// Set dirty flag
			if _, err := tx.Exec(ctx, "UPDATE schema_migrations SET dirty = true WHERE version = $1", migration.ID); err != nil {
				tx.Rollback(ctx)
				span.RecordError(err)
				span.End()
				return fmt.Errorf("failed to set dirty flag: %w", err)
			}

			// Run rollback
			if _, err := tx.Exec(ctx, migration.Down); err != nil {
				tx.Rollback(ctx)
				span.RecordError(err)
				span.End()
				return fmt.Errorf("rollback %d failed: %w", migration.ID, err)
			}

			// Remove migration record
			if _, err := tx.Exec(ctx, "DELETE FROM schema_migrations WHERE version = $1", migration.ID); err != nil {
				tx.Rollback(ctx)
				span.RecordError(err)
				span.End()
				return fmt.Errorf("failed to remove migration record: %w", err)
			}

			if err := tx.Commit(ctx); err != nil {
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
