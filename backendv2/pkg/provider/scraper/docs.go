package scraper

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/errgroup"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/license"
	"github.com/opentofu/registry-ui/pkg/provider/storage"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

// maxFileSize limits the amount of data read from the docs to 10MB to prevent memory-based DoS.
const maxFileSize = 10 * 1024 * 1024

// ProviderVersion represents a provider version with documentation
type ProviderVersion struct {
	ID                  string                  `json:"id"`
	Namespace           string                  `json:"namespace"`
	Name                string                  `json:"name"`
	Version             string                  `json:"version"`
	Published           time.Time               `json:"published"`
	Docs                ProviderDocs            `json:"docs"`
	CDKTFDocs           map[string]ProviderDocs `json:"cdktf_docs"`
	License             license.List            `json:"license,omitempty"`
	IncompatibleLicense bool                    `json:"incompatible_license"`
	Link                string                  `json:"link"`
}

// ProviderDocs represents provider documentation structure
type ProviderDocs struct {
	Guides        []DocItem `json:"guides,omitempty"`
	Resources     []DocItem `json:"resources,omitempty"`
	DataSources   []DocItem `json:"datasources,omitempty"`
	Functions     []DocItem `json:"functions,omitempty"`
	Ephemeral     []DocItem `json:"ephemeral,omitempty"`
	Actions       []DocItem `json:"actions,omitempty"`
	ListResources []DocItem `json:"listresources,omitempty"`
	Index         *DocItem  `json:"index,omitempty"`
}

// DocItem represents a documentation item
type DocItem struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Path        string `json:"path,omitempty"`
	Title       string `json:"title" yaml:"page_title"`
	Subcategory string `json:"subcategory,omitempty" yaml:"subcategory"`
	Description string `json:"description,omitempty" yaml:"description"`
	EditLink    string `json:"edit_link,omitempty" yaml:"edit_link"`
	contents    []byte
	md5Checksum string // MD5 checksum of the document contents
	isError     bool
}


type Scraper struct {
	config   *config.BackendConfig
	s3Client *s3.Client
	uploader *manager.Uploader
	pool     *pgxpool.Pool
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

	return s.StoreDocs(ctx, namespace, name, version, docs, licenses, tx)
}

// StoreDocs uploads already-scraped documentation to S3 and stores metadata in the database.
func (s *Scraper) StoreDocs(ctx context.Context, namespace, name, version string, docs map[string]*DocItem, licenses license.List, tx pgx.Tx) error {
	if err := s.saveToBucket(ctx, namespace, name, version, docs); err != nil {
		return err
	}
	// Convert docs to storage format
	storageDocs := make(map[string]*storage.DocItem)
	for path, doc := range docs {
		storageDocs[path] = &storage.DocItem{
			Name:        doc.Name,
			Title:       doc.Title,
			Subcategory: doc.Subcategory,
			Description: doc.Description,
			EditLink:    doc.EditLink,
			MD5Checksum: doc.md5Checksum,
		}
	}

	if err := storage.StoreProviderDocuments(ctx, tx, namespace, name, version, storageDocs); err != nil {
		return fmt.Errorf("failed to store documents in database: %w", err)
	}

	return s.GenerateAndStoreIndex(ctx, namespace, name, version, docs, licenses)
}

func (s *Scraper) ScrapeDocumentation(ctx context.Context, namespace, name, version, directory string) (map[string]*DocItem, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "docscraper.scrape")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
		attribute.String("provider.version", version),
		attribute.String("scrape.directory", directory),
	)

	slog.DebugContext(ctx, "Starting documentation scraping",
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
	slog.DebugContext(ctx, "Successfully scraped documentation",
		"namespace", namespace, "name", name, "version", version, "docs_count", len(docs))

	return docs, nil
}

// scrapeDocTypes scrapes documentation for all configured doc types in the given base directory
func (s *Scraper) scrapeDocTypes(ctx context.Context, fsys fs.ReadDirFS, baseDir string, docs map[string]*DocItem, repoURL, version, pathPrefix string) error {
	for _, docType := range storage.DocTypes {
		for _, sourceDir := range docType.SourceDirs {
			fullSourceDir := path.Join(baseDir, sourceDir)
			targetPrefix := docType.TargetPath
			if pathPrefix != "" {
				targetPrefix = path.Join(pathPrefix, docType.TargetPath)
			}

			// this is inefficient but it works for now
			if err := s.scrapeType(ctx, fsys, fullSourceDir, targetPrefix, docs, repoURL, version); err != nil {
				// Log debug message but don't fail - some directories may not exist
				slog.DebugContext(ctx, "Failed to scrape directory", "directory", fullSourceDir, "error", err)
			}
		}
	}
	return nil
}

func (s *Scraper) saveToBucket(ctx context.Context, namespace, name, version string, docs map[string]*DocItem) error {
	ctx, span := telemetry.Tracer().Start(ctx, "docscraper.saveToBucket")
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
	slog.DebugContext(ctx, "Successfully uploaded documentation",
		"namespace", namespace, "name", name, "version", version, "docs_count", uploadCount)

	return nil
}

func (s *Scraper) GenerateAndStoreIndex(ctx context.Context, namespace, name, version string, docs map[string]*DocItem, licenses license.List) error {
	ctx, span := telemetry.Tracer().Start(ctx, "docscraper.generate_index")
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
	slog.DebugContext(ctx, "Successfully generated and uploaded provider index.json",
		"provider.namespace", namespace, "provider.name", name, "provider.version", version, "s3_key", key, "json_size", len(jsonData))

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
		contents = fmt.Appendf(nil, "# File Too Large\n\nThis file is too large to display. View it directly in the repository: %s/blob/%s/%s",
			repoURL, version, fn)
	} else {
		contents, err = s.readFile(fsys, fn)
		if err != nil {
			return nil, err
		}
	}

	hash := md5.Sum(contents)
	checksum := hex.EncodeToString(hash[:])

	doc := &DocItem{
		Name:        name,
		contents:    contents,
		md5Checksum: checksum,
	}

	// Extract frontmatter
	if err := s.extractFrontmatterPermissively(contents, doc); err != nil {
		slog.WarnContext(ctx, "Failed to extract frontmatter", "file", fn, "error", err)
		// Don't fail completely, but record this as an error document
		doc.isError = true
	}

	// Generate edit link
	doc.EditLink = fmt.Sprintf("%s/blob/v%s/%s", repoURL, version, fn)

	return doc, nil
}

func (s *Scraper) readFile(fsys fs.ReadDirFS, fn string) ([]byte, error) {
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

// bulkUploadDocs uploads multiple docs concurrently using errgroup
func (s *Scraper) bulkUploadDocs(ctx context.Context, namespace, name, version string, docs map[string]*DocItem) error {
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(s.config.Concurrency.Upload)

	slog.DebugContext(ctx, "Starting bulk upload of docs",
		"namespace", namespace, "name", name, "version", version, "count", len(docs), "concurrency", s.config.Concurrency.Upload)

	// Fetch all existing checksums in one query (instead of per-document)
	existingChecksums, err := storage.GetAllDocumentChecksums(ctx, s.pool, namespace, name, version)
	if err != nil {
		slog.WarnContext(ctx, "Failed to fetch existing checksums, will upload all docs",
			"namespace", namespace, "name", name, "version", version, "error", err)
		existingChecksums = make(map[string]string) // Continue without skip optimization
	}

	for filePath, doc := range docs {
		filePath, doc := filePath, doc
		g.Go(func() error {
			return s.uploadDocToS3(gctx, namespace, name, version, filePath, doc, existingChecksums)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("bulk upload failed: %w", err)
	}

	slog.DebugContext(ctx, "Completed bulk upload of docs",
		"namespace", namespace, "name", name, "version", version, "count", len(docs))
	return nil
}

func (s *Scraper) uploadDocToS3(ctx context.Context, namespace, name, version, filePath string, doc *DocItem, existingChecksums map[string]string) error {
	// filePath should already be normalized during scraping (e.g., "resources/aws_instance")
	// the prefix of providers is fine here because we dont use this method for anything else

	key := fmt.Sprintf("providers/%s/%s/%s/%s.md", namespace, name, version, filePath)

	// Check if a document with the same checksum already exists (content-based deduplication)
	docType := getDocCategory(filePath)
	language := extractLanguage(filePath)

	if docType != "" && doc.md5Checksum != "" {
		// Use pre-fetched checksums map instead of per-document DB query
		checksumKey := fmt.Sprintf("%s:%s:%s", docType, doc.Name, language)
		if existingChecksum, ok := existingChecksums[checksumKey]; ok && existingChecksum == doc.md5Checksum {
			slog.DebugContext(ctx, "Skipping S3 upload - document content unchanged (checksum match)",
				"key", key, "checksum", doc.md5Checksum)
			return nil // Skip upload, content is identical
		}
	}

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

	slog.DebugContext(ctx, "Uploaded document to S3",
		"key", key, "checksum", doc.md5Checksum)

	return nil
}

func (s *Scraper) buildProviderVersionJSON(namespace, name, version string, docs map[string]*DocItem, licenses license.List) *ProviderVersion {
	providerDocs := s.buildProviderDocs(docs)
	cdktfDocs := s.buildCDKTFDocs(docs)

	return &ProviderVersion{
		ID:                  version,
		Namespace:           namespace,
		Name:                name,
		Version:             version,
		Published:           time.Now(),
		Docs:                providerDocs,
		CDKTFDocs:           cdktfDocs,
		License:             licenses,
		IncompatibleLicense: !licenses.IsRedistributable(),
		Link:                fmt.Sprintf("https://github.com/%s/terraform-provider-%s", namespace, name),
	}
}

func (s *Scraper) buildProviderDocs(docs map[string]*DocItem) ProviderDocs {
	result := ProviderDocs{}
	resultVal := reflect.ValueOf(&result).Elem()

	for filePath, doc := range docs {
		// Skip CDKTF docs - they'll be handled separately
		if strings.Contains(filePath, "cdktf/") {
			continue
		}

		docItem := DocItem{
			Name:        doc.Name,
			EditLink:    doc.EditLink,
			Title:       doc.Title,
			Subcategory: doc.Subcategory,
			Description: doc.Description,
		}

		category := getDocCategory(filePath)
		if category == "index" {
			result.Index = &docItem
			continue
		}

		appendDocToField(resultVal, category, docItem)
	}

	return result
}

// appendDocToField uses reflection to append a DocItem to the appropriate field in ProviderDocs
// based on the category. The field name is looked up from the storage.DocTypes map.
func appendDocToField(resultVal reflect.Value, category string, docItem DocItem) {
	for _, cfg := range storage.DocTypes {
		if cfg.TargetPath == category {
			field := resultVal.FieldByName(cfg.FieldName)
			if field.IsValid() && field.CanSet() {
				field.Set(reflect.Append(field, reflect.ValueOf(docItem)))
			}
			return
		}
	}
}

func (s *Scraper) buildCDKTFDocs(docs map[string]*DocItem) map[string]ProviderDocs {
	result := map[string]ProviderDocs{}

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
			result[language] = ProviderDocs{}
		}

		langDocs := result[language]
		langDocsVal := reflect.ValueOf(&langDocs).Elem()

		// Convert our DocItem to DocItem
		docItem := DocItem{
			Name:        doc.Name,
			EditLink:    doc.EditLink,
			Title:       doc.Title,
			Subcategory: doc.Subcategory,
			Description: doc.Description,
		}

		// Categorize based on remaining path (everything after language) using unified categorization logic
		remainingPath := strings.Join(parts[cdktfIndex+2:], "/")
		category := getDocCategory(remainingPath)
		if category == "index" {
			langDocs.Index = &docItem
		} else {
			appendDocToField(langDocsVal, category, docItem)
		}

		result[language] = langDocs
	}

	return result
}
