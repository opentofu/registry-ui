package docscraper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/errgroup"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/license"
	"github.com/opentofu/registry-ui/pkg/providers"
)

// maxFileSize limits the amount of data read from the docs to 10MB to prevent memory-based DoS.
const maxFileSize = 10 * 1024 * 1024

// docTypes defines the mapping between source directories and target categorization paths
var docTypes = map[string]struct {
	SourceDirs []string
	TargetPath string
}{
	"resources": {
		SourceDirs: []string{"r", "resources"},
		TargetPath: "resources",
	},
	"datasources": {
		SourceDirs: []string{"d", "data-sources", "datasources"},
		TargetPath: "datasources",
	},
	"functions": {
		SourceDirs: []string{"f", "functions"},
		TargetPath: "functions",
	},
	"guides": {
		SourceDirs: []string{"guides"},
		TargetPath: "guides",
	},
	"ephemeral": {
		SourceDirs: []string{"ephemeral", "ephemeral-resources"},
		TargetPath: "ephemeral",
	},
}

type Scraper struct {
	config   *config.BackendConfig
	s3Client *s3.Client
	uploader *manager.Uploader
	pool     *pgxpool.Pool
}

type DocItem struct {
	Name        string `yaml:"page_title"`
	Title       string `yaml:"page_title"`
	Subcategory string `yaml:"subcategory"`
	Description string `yaml:"description"`
	EditLink    string
	contents    []byte
	isError     bool
}

func New(cfg *config.BackendConfig, s3Client *s3.Client, pool *pgxpool.Pool) *Scraper {
	return &Scraper{
		config:   cfg,
		s3Client: s3Client,
		uploader: manager.NewUploader(s3Client),
		pool:     pool,
	}
}

func (s *Scraper) ScrapeAndStore(ctx context.Context, namespace, name, version, directory string, licenses license.List, tx pgx.Tx) error {
	docs, err := s.ScrapeDocumentation(ctx, namespace, name, version, directory)
	if err != nil {
		return err
	}

	if err := s.saveToBucket(ctx, namespace, name, version, docs); err != nil {
		return err
	}
	if err := s.saveToDB(ctx, tx, namespace, name, version, docs); err != nil {
		return fmt.Errorf("failed to store documents in database: %w", err)
	}

	return s.GenerateAndStoreIndex(ctx, namespace, name, version, docs, licenses)
}

func (s *Scraper) ScrapeDocumentation(ctx context.Context, namespace, name, version, directory string) (map[string]*DocItem, error) {
	tracer := otel.Tracer("opentofu-registry-backend")
	ctx, span := tracer.Start(ctx, "docscraper.scrape")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
		attribute.String("provider.version", version),
		attribute.String("scrape.directory", directory),
	)

	slog.InfoContext(ctx, "Starting documentation scraping",
		"namespace", namespace, "name", name, "version", version, "directory", directory)

	fsys := os.DirFS(directory)
	readDirFS, ok := fsys.(fs.ReadDirFS)
	if !ok {
		span.RecordError(fmt.Errorf("filesystem does not implement ReadDirFS"))
		return nil, fmt.Errorf("filesystem does not implement ReadDirFS")
	}

	docs := make(map[string]*DocItem)
	repoURL := fmt.Sprintf("https://github.com/%s/terraform-provider-%s", namespace, name)

	/*
		Right now in the existing registry implementations, provider documentation can come in a couple different formats:

		Latest format:
			- all inside /docs
			- /docs/guides/<guide>.md for guides
			- /docs/resources/<resource>.md for resources
			- /docs/data-sources/<data-source>.md for resources.
			- /docs/functions/<function>.md for functions
		Legacy format:
			- all inside /website/docs
			- all inside /website/docs
			- /website/docs/guides/<guide>.[html.markdown|html.md]
			- /website/docs/r/<resource>.[html.markdown|html.md]
			- /website/docs/d/<data-source>.[html.markdown|html.md]
			- /website/docs/functions/<function>.[html.markdown|html.md]

		And both formats support the new cdktf docs structure, which is as follows

		Latest format:
			/docs/cdktf/[python|typescript]/resources/<data-source>.md
			/docs/cdktf/[python|typescript]/data-sources/<data-source>.md
			/docs/cdktf/[python|typescript]/functions/<function>.md
		Legacy format:
			/website/docs/cdktf/[python|typescript]/r/<resource>.[html.markdown|html.md]
			/website/docs/cdktf/[python|typescript]/d/<data-source>.[html.markdown|html.md]
			/website/docs/cdktf/[python|typescript]/f/<function>.[html.markdown|html.md]

		We want to move all the legacy docs into the new format, so that we can have a single source of truth for the docs.

		Once the docs have been downloaded, we need to rejig them into the new format, the directory should already be normalized (ie, files are not inside ./website/docs/guides, but rather ./guides)
	*/

	for _, dir := range []string{
		path.Join("website", "docs"),
		"docs",
	} {
		// Note: we prefer website/docs over docs and not include docs because some providers
		// (such as AWS) use the docs directory for internal documentation.
		foundDocs, err := s.scrapeDir(ctx, dir, readDirFS, docs, repoURL, version)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scrape directory %s: %w", dir, err)
		}
		if foundDocs {
			break
		}
	}

	span.SetAttributes(attribute.Int("docs.scraped_count", len(docs)))
	slog.InfoContext(ctx, "Successfully scraped documentation",
		"namespace", namespace, "name", name, "version", version, "docs_count", len(docs))

	return docs, nil
}

// scrapeDocTypes scrapes documentation for all configured doc types in the given base directory
func (s *Scraper) scrapeDocTypes(ctx context.Context, fsys fs.ReadDirFS, baseDir string, docs map[string]*DocItem, repoURL, version, pathPrefix string) error {
	for _, docType := range docTypes {
		for _, sourceDir := range docType.SourceDirs {
			sourceDir := path.Join(baseDir, sourceDir)
			targetPrefix := docType.TargetPath
			if pathPrefix != "" {
				targetPrefix = path.Join(pathPrefix, docType.TargetPath)
			}

			// this is inefficient but it works for now
			if err := s.scrapeType(ctx, fsys, sourceDir, targetPrefix, docs, repoURL, version); err != nil {
				// Log debug message but don't fail - some directories may not exist
				slog.DebugContext(ctx, "Failed to scrape directory", "directory", sourceDir, "error", err)
			}
		}
	}
	return nil
}

func (s *Scraper) saveToBucket(ctx context.Context, namespace, name, version string, docs map[string]*DocItem) error {
	tracer := otel.Tracer("opentofu-registry-backend")
	ctx, span := tracer.Start(ctx, "docscraper.saveToBucket")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
		attribute.String("provider.version", version),
		attribute.Int("docs.count", len(docs)),
	)

	uploadCount := len(docs)
	if uploadCount > 0 {
		if err := s.bulkUploadDocs(ctx, namespace, name, version, docs); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to bulk upload docs to S3: %w", err)
		}
	}

	span.SetAttributes(attribute.Int("docs.uploaded_count", uploadCount))
	slog.InfoContext(ctx, "Successfully uploaded documentation",
		"namespace", namespace, "name", name, "version", version, "docs_count", uploadCount)

	return nil
}

func (s *Scraper) GenerateAndStoreIndex(ctx context.Context, namespace, name, version string, docs map[string]*DocItem, licenses license.List) error {
	tracer := otel.Tracer("opentofu-registry-backend")
	ctx, span := tracer.Start(ctx, "docscraper.generate_index")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
		attribute.String("provider.version", version),
		attribute.Int("docs.count", len(docs)),
	)

	providerVersion := s.buildProviderVersionJSON(namespace, name, version, docs, licenses)

	jsonData, err := json.MarshalIndent(providerVersion, "", "  ")
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to marshal provider version JSON: %w", err)
	}

	key := fmt.Sprintf("providers/%s/%s/%s/index.json", namespace, name, version)
	_, err = s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.config.Bucket.BucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(jsonData),
		ContentType: aws.String("application/json"),
	})

	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to upload index.json to S3: %w", err)
	}

	span.SetAttributes(attribute.String("index.s3_key", key))
	slog.InfoContext(ctx, "Successfully generated and uploaded provider index.json",
		"namespace", namespace, "name", name, "version", version, "s3_key", key, "json_size", len(jsonData))

	return nil
}

var normalizeRe = regexp.MustCompile("[^a-zA-Z0-9-_.]")

func normalizeName(name string) string {
	return normalizeRe.ReplaceAllString(name, "")
}

func (s *Scraper) scrapeDir(ctx context.Context, dir string, fsys fs.ReadDirFS, docs map[string]*DocItem, repoURL, version string) (bool, error) {
	_, err := fsys.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if err := s.scrapeDocumentation(ctx, fsys, dir, docs, repoURL, version); err != nil {
		return true, fmt.Errorf("failed to scrape documentation directory: %w", err)
	}

	// Handle CDKTF documentation separately
	cdktfDir := path.Join(dir, "cdktf")
	if err := s.scrapeCDKTFDocumentation(ctx, fsys, cdktfDir, docs, repoURL, version); err != nil {
		return true, fmt.Errorf("failed to scrape CDKTF documentation: %w", err)
	}

	return true, nil
}

func (s *Scraper) scrapeCDKTFDocumentation(ctx context.Context, fsys fs.ReadDirFS, cdktfDir string, docs map[string]*DocItem, repoURL, version string) error {
	cdktfItems, err := fsys.ReadDir(cdktfDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to list CDKTF documentation directory: %w", err)
	}

	for _, item := range cdktfItems {
		if !item.IsDir() {
			continue
		}

		lang := item.Name()
		if lang != "python" && lang != "typescript" && lang != "csharp" && lang != "java" && lang != "go" {
			slog.DebugContext(ctx, "Skipping unknown CDKTF language", "language", lang)
			continue
		}

		langDir := path.Join(cdktfDir, lang)
		indexPrefix := fmt.Sprintf("cdktf/%s", lang)
		if err := s.extractRootDoc(ctx, fsys, langDir, docs, repoURL, version, indexPrefix); err != nil {
			slog.WarnContext(ctx, "Failed to extract CDKTF index", "language", lang, "error", err)
		}

		// Scrape CDKTF documentation using unified approach
		pathPrefix := fmt.Sprintf("cdktf/%s", lang)
		if err := s.scrapeDocTypes(ctx, fsys, langDir, docs, repoURL, version, pathPrefix); err != nil {
			slog.DebugContext(ctx, "Failed to scrape CDKTF documentation", "language", lang, "error", err)
		}
	}

	return nil
}

func (s *Scraper) scrapeDocumentation(ctx context.Context, fsys fs.ReadDirFS, dir string, docs map[string]*DocItem, repoURL, version string) error {
	// Extract root document (index file)
	if err := s.extractRootDoc(ctx, fsys, dir, docs, repoURL, version); err != nil {
		return err
	}

	// Scrape different types of documentation using unified approach
	if err := s.scrapeDocTypes(ctx, fsys, dir, docs, repoURL, version, ""); err != nil {
		return err
	}

	return nil
}

func (s *Scraper) extractRootDoc(ctx context.Context, fsys fs.ReadDirFS, dir string, docs map[string]*DocItem, repoURL, version string, pathPrefixes ...string) error {
	pathPrefix := ""
	if len(pathPrefixes) > 0 {
		pathPrefix = pathPrefixes[0]
	}
	items, err := fsys.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	for _, item := range items {
		if item.IsDir() {
			continue
		}
		if strings.HasPrefix(item.Name(), "index.") {
			doc, err := s.readDocFile(ctx, dir, item, fsys, repoURL, version)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", item.Name(), err)
			}
			if doc != nil {
				indexKey := "index"
				if pathPrefix != "" {
					indexKey = path.Join(pathPrefix, "index")
				}
				docs[indexKey] = doc
			}
		}
	}

	return nil
}

var suffixes = []string{
	".html.md",
	".html.markdown",
	".md.html",
	".markdown.html",
	".md",
	".markdown",
}

func (s *Scraper) scrapeType(ctx context.Context, fsys fs.ReadDirFS, dir, pathPrefix string, docs map[string]*DocItem, repoURL, version string) error {
	items, err := fsys.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	for _, item := range items {
		if item.IsDir() {
			continue
		}

		doc, err := s.readDocFile(ctx, dir, item, fsys, repoURL, version)
		if err != nil {
			return err
		}
		if doc != nil {
			// Normalize the path for consistent storage
			normalizedPath := path.Join(pathPrefix, doc.Name)
			docs[normalizedPath] = doc
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	return nil
}

func (s *Scraper) readDocFile(ctx context.Context, dir string, item fs.DirEntry, fsys fs.ReadDirFS, repoURL, version string) (*DocItem, error) {
	fn := path.Join(dir, item.Name())
	name := normalizeName(path.Base(fn))

	slog.DebugContext(ctx, "Processing documentation file", "file", fn, "normalized_name", name)

	// Check for valid suffix
	suffixFound := false
	for _, suffix := range suffixes {
		if strings.HasSuffix(name, suffix) {
			suffixFound = true
			name = strings.TrimSuffix(name, suffix)
			break
		}
	}
	if !suffixFound {
		return nil, nil
	}

	// Validate document name (basic validation)
	if name == "" || strings.Contains(name, "..") {
		slog.WarnContext(ctx, "Invalid document name, skipping", "name", name)
		return nil, nil
	}

	stat, err := item.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to stat %s: %w", fn, err)
	}

	var contents []byte
	if stat.Size() > maxFileSize {
		contents = []byte(fmt.Sprintf("# File Too Large\n\nThis file is too large to display. View it directly in the repository: %s/blob/%s/%s",
			repoURL, version, fn))
	} else {
		contents, err = s.readFile(ctx, fsys, fn)
		if err != nil {
			return nil, err
		}
	}

	doc := &DocItem{
		Name:     name,
		contents: contents,
	}

	// Extract frontmatter
	if err := s.extractFrontmatterPermissively(ctx, contents, doc); err != nil {
		slog.WarnContext(ctx, "Failed to extract frontmatter", "file", fn, "error", err)
		// Don't fail completely, but record this as an error document
		doc.isError = true
	}

	// Generate edit link
	doc.EditLink = fmt.Sprintf("%s/blob/v%s/%s", repoURL, version, fn)

	return doc, nil
}

func (s *Scraper) readFile(ctx context.Context, fsys fs.ReadDirFS, fn string) ([]byte, error) {
	fh, err := fsys.Open(fn)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", fn, err)
	}
	defer fh.Close()

	contents, err := io.ReadAll(fh)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", fn, err)
	}

	return contents, nil
}

// bulkUploadDocs uploads multiple docs concurrently using errgroup
func (s *Scraper) bulkUploadDocs(ctx context.Context, namespace, name, version string, docs map[string]*DocItem) error {
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(s.config.Concurrency.Upload)

	slog.InfoContext(ctx, "Starting bulk upload of docs",
		"namespace", namespace, "name", name, "version", version, "count", len(docs), "concurrency", s.config.Concurrency.Upload)

	for filePath, doc := range docs {
		filePath, doc := filePath, doc
		g.Go(func() error {
			return s.uploadDocToS3(gctx, namespace, name, version, filePath, doc)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("bulk upload failed: %w", err)
	}

	slog.InfoContext(ctx, "Completed bulk upload of docs",
		"namespace", namespace, "name", name, "version", version, "count", len(docs))
	return nil
}

func (s *Scraper) uploadDocToS3(ctx context.Context, namespace, name, version, filePath string, doc *DocItem) error {
	// filePath should already be normalized during scraping (e.g., "resources/aws_instance")
	// the prefix of providers is fine here because we dont use this method for anything else
	key := fmt.Sprintf("providers/%s/%s/%s/%s.md", namespace, name, version, filePath)

	_, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.config.Bucket.BucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(doc.contents),
		ContentType: aws.String("text/markdown"),
		Metadata: map[string]string{
			"title":       doc.Title,
			"subcategory": doc.Subcategory,
			"edit-link":   doc.EditLink,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to upload %s to S3: %w", key, err)
	}

	return nil
}

func (s *Scraper) buildProviderVersionJSON(namespace, name, version string, docs map[string]*DocItem, licenses license.List) *providers.ProviderVersion {
	providerDocs := s.buildProviderDocs(docs)
	cdktfDocs := s.buildCDKTFDocs(docs)

	return &providers.ProviderVersion{
		ID:                  version,
		Published:           time.Now(),
		Docs:                providerDocs,
		CDKTFDocs:           cdktfDocs,
		License:             licenses,
		IncompatibleLicense: !licenses.IsRedistributable(),
		Link:                fmt.Sprintf("https://github.com/%s/terraform-provider-%s", namespace, name),
	}
}

func (s *Scraper) buildProviderDocs(docs map[string]*DocItem) providers.ProviderDocs {
	result := providers.ProviderDocs{}

	for filePath, doc := range docs {
		// Skip CDKTF docs - they'll be handled separately
		if strings.Contains(filePath, "cdktf/") {
			continue
		}

		docItem := providers.DocItem{
			Name:        doc.Name,
			EditLink:    doc.EditLink,
			Title:       doc.Title,
			Subcategory: doc.Subcategory,
			Description: doc.Description,
		}

		switch getDocCategory(filePath) {
		case "resources":
			result.Resources = append(result.Resources, docItem)
		case "datasources":
			result.DataSources = append(result.DataSources, docItem)
		case "functions":
			result.Functions = append(result.Functions, docItem)
		case "guides":
			result.Guides = append(result.Guides, docItem)
		case "ephemeral":
			result.Ephemeral = append(result.Ephemeral, docItem)
		case "index":
			result.Index = &docItem
		}
	}

	return result
}

func (s *Scraper) buildCDKTFDocs(docs map[string]*DocItem) map[string]providers.ProviderDocs {
	result := map[string]providers.ProviderDocs{}

	for filePath, doc := range docs {
		// Only process CDKTF docs
		if !strings.Contains(filePath, "cdktf/") {
			continue
		}

		// Extract language from path - handle both `website/docs` and `docs` by iterating across all the parts and looking for `cdktf`
		parts := strings.Split(filePath, "/")

		cdktfIndex := -1
		for i, part := range parts {
			if part == "cdktf" {
				cdktfIndex = i
				break
			}
		}
		// check if there's more of the path to actually parse
		if cdktfIndex == -1 || len(parts) < cdktfIndex+2 {
			continue
		}

		language := parts[cdktfIndex+1] // Language should come after "cdktf"

		// Initialize language docs if not exists
		if _, exists := result[language]; !exists {
			result[language] = providers.ProviderDocs{}
		}

		langDocs := result[language]

		// Convert our DocItem to providers.DocItem
		docItem := providers.DocItem{
			Name:        doc.Name,
			EditLink:    doc.EditLink,
			Title:       doc.Title,
			Subcategory: doc.Subcategory,
			Description: doc.Description,
		}

		// Categorize based on remaining path (everything after language) using unified categorization logic
		remainingPath := strings.Join(parts[cdktfIndex+2:], "/")
		switch getDocCategory(remainingPath) {
		case "resources":
			langDocs.Resources = append(langDocs.Resources, docItem)
		case "datasources":
			langDocs.DataSources = append(langDocs.DataSources, docItem)
		case "functions":
			langDocs.Functions = append(langDocs.Functions, docItem)
		case "guides":
			langDocs.Guides = append(langDocs.Guides, docItem)
		case "ephemeral":
			langDocs.Ephemeral = append(langDocs.Ephemeral, docItem)
		case "index":
			langDocs.Index = &docItem
		}

		result[language] = langDocs
	}

	return result
}

// saveToDB stores all document metadata in the database using bulk inserts
func (s *Scraper) saveToDB(ctx context.Context, tx pgx.Tx, namespace, name, version string, docs map[string]*DocItem) error {
	tracer := otel.Tracer("opentofu-registry-backend")
	ctx, span := tracer.Start(ctx, "docscraper.saveToDB")
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

	slog.InfoContext(ctx, "Storing document metadata in database",
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
			 title, subcategory, description, edit_link, s3_key, language)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (provider_namespace, provider_name, version, document_type, document_name, language)
			DO UPDATE SET 
				title = EXCLUDED.title,
				subcategory = EXCLUDED.subcategory,
				description = EXCLUDED.description,
				edit_link = EXCLUDED.edit_link,
				s3_key = EXCLUDED.s3_key,
				updated_at = NOW()
		`, namespace, name, version, docType, doc.Name,
			doc.Title, doc.Subcategory, doc.Description, doc.EditLink, s3Key, language)
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
	slog.InfoContext(ctx, "Successfully stored document metadata in database",
		"namespace", namespace, "name", name, "version", version,
		"inserted_count", insertCount, "total_docs", len(docs))

	return nil
}
