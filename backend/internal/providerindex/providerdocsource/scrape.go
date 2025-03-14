package providerdocsource

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/opentofu/registry-ui/internal/license"
	"github.com/opentofu/registry-ui/internal/providerindex/providerindexstorage"
	"github.com/opentofu/registry-ui/internal/providerindex/providertypes"
	"github.com/opentofu/registry-ui/internal/registry/provider"
)

// maxFileSize limits the amount of data read from the docs to 1MB to prevent memory-based DoS.
const maxFileSize = 1024 * 1024

var tplTooLarge = template.Must(template.New("").Parse(string(errorTooLarge)))

func Process(ctx context.Context, raw provider.Provider, workingCopy string, version providertypes.ProviderVersionDescriptor, licenseDetector license.Detector, logger *slog.Logger, storage providerindexstorage.API) (providertypes.ProviderVersion, error) {
	licenses, err := licenseDetector(workingCopy, raw.RepositoryURL()+"/blob/"+version.ID)
	if err != nil {
		return providertypes.ProviderVersion{}, fmt.Errorf("failed to detect licenses")
	}

	doc := providertypes.ProviderVersion{
		ProviderVersionDescriptor: version,
		CDKTFDocs:                 make(map[providertypes.CDKTFLanguage]providertypes.ProviderDocs),
		Licenses:                  licenses,
		IncompatibleLicense:       !licenses.IsRedistributable(),
		Link:                      raw.RepositoryURL() + "/tree/" + version.ID,
	}

	if !licenses.IsRedistributable() {
		// Store error in the index
		err := storage.StoreProviderDocItem(
			ctx,
			providertypes.Addr(raw),
			version.ID,
			providertypes.DocItemKindRoot,
			"index",
			errorIncompatibleLicense,
		)
		return doc, err
	}

	s := source{
		ctx:         ctx,
		raw:         raw,
		version:     version.ID,
		logger:      logger,
		workingCopy: workingCopy,
		storage:     storage,
	}
	for _, dir := range []string{
		path.Join("website", "docs"),
		"docs",
	} {
		if _, err := os.Stat(path.Join(s.workingCopy, dir)); err != nil {
			if os.IsNotExist(err) {
				// Note: we prefer website/docs over docs and not include docs because some providers
				// (such as AWS) use the docs directory for internal documentation.
				continue
			}
			return doc, err
		}

		doc.Docs, err = s.scrapeDocumentation(dir, "")
		if err != nil {
			return doc, fmt.Errorf("failed to scrape documentation directory (%w)", err)
		}

		doc.CDKTFDocs, err = s.scrapeCdktf(dir)
		if err != nil {
			return doc, fmt.Errorf("failed to scrape documentation directory (%w)", err)
		}
		break
	}

	if err := storage.StoreProviderVersion(ctx, providertypes.Addr(raw), doc); err != nil {
		return doc, err
	}

	return doc, nil
}

var normalizeRe = regexp.MustCompile("[^a-zA-Z0-9-_.]")

func normalizeName(name string) string {
	return normalizeRe.ReplaceAllString(name, "")
}

type source struct {
	ctx         context.Context
	raw         provider.Provider
	version     string
	logger      *slog.Logger
	workingCopy string
	storage     providerindexstorage.API
}

func (s source) scrapeCdktf(dir string) (map[providertypes.CDKTFLanguage]providertypes.ProviderDocs, error) {
	cdktf := make(map[providertypes.CDKTFLanguage]providertypes.ProviderDocs)

	cdktfDir := path.Join(dir, "cdktf")
	cdktfItems, err := os.ReadDir(path.Join(s.workingCopy, cdktfDir))
	if err != nil {
		if os.IsNotExist(err) {
			return cdktf, nil
		}
		return cdktf, fmt.Errorf("failed to list CDKTF documentation directory (%w)", err)
	}
	for _, item := range cdktfItems {
		if !item.IsDir() {
			continue
		}
		lang := providertypes.CDKTFLanguage(item.Name())
		if err := lang.Validate(); err != nil {
			s.logger.DebugContext(s.ctx, "%s is not a valid CDKTF language, skipping... (%v)", item.Name(), err)
			continue
		}
		cdktfDoc, err := s.scrapeDocumentation(path.Join(cdktfDir, string(lang)), lang)
		if err != nil {
			return cdktf, fmt.Errorf("failed to scrape CDKTF documentation for %s (%w)", item.Name(), err)
		}
		cdktf[lang] = cdktfDoc
	}
	return cdktf, nil
}

func (s source) scrapeDocumentation(dir string, language providertypes.CDKTFLanguage) (providertypes.ProviderDocs, error) {
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

	docs := providertypes.ProviderDocs{
		// Initialize for non-null api output
		Resources:   make([]providertypes.ProviderDocItem, 0),
		DataSources: make([]providertypes.ProviderDocItem, 0),
		Functions:   make([]providertypes.ProviderDocItem, 0),
		Guides:      make([]providertypes.ProviderDocItem, 0),
	}
	var err error

	docs.Root, err = s.extractRootDoc(dir, language)
	if err != nil {
		return docs, err
	}

	for _, scrape := range []struct {
		dir         string
		destination *[]providertypes.ProviderDocItem
		kind        providertypes.DocItemKind
	}{
		{"r", &docs.Resources, providertypes.DocItemKindResource},
		{"resources", &docs.Resources, providertypes.DocItemKindResource},
		{"d", &docs.DataSources, providertypes.DocItemKindDataSource},
		{"data-sources", &docs.DataSources, providertypes.DocItemKindDataSource},
		{"f", &docs.Functions, providertypes.DocItemKindFunction},
		{"functions", &docs.Functions, providertypes.DocItemKindFunction},
		{"guides", &docs.Guides, providertypes.DocItemKindGuide},
	} {
		sourceDir := path.Join(dir, scrape.dir)
		found, err := s.scrapeType(sourceDir, language, scrape.kind)
		if err != nil {
			return docs, fmt.Errorf("failed to scrape %s, %w", sourceDir, err)
		}
		(*scrape.destination) = append(*scrape.destination, found...)
	}
	return docs, nil
}

func (s source) extractRootDoc(dir string, language providertypes.CDKTFLanguage) (*providertypes.ProviderDocItem, error) {
	items, err := os.ReadDir(filepath.Join(s.workingCopy, dir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read directory %s (%w)", dir, err)
	}
	for _, item := range items {
		if item.IsDir() {
			continue
		}
		if strings.HasPrefix(item.Name(), "index.") {
			root, err := s.readDocFile(dir, item, language, providertypes.DocItemKindRoot)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %s (%w)", item.Name(), err)
			}
			if root != nil {
				return root, nil
			}
		}
	}
	return nil, nil
}

var suffixes = []string{
	".html.md",
	".html.markdown",
	".md.html",
	".markdown.html",
	".md",
	".markdown",
}

func (s source) scrapeType(dir string, language providertypes.CDKTFLanguage, kind providertypes.DocItemKind) ([]providertypes.ProviderDocItem, error) {
	items, err := os.ReadDir(filepath.Join(s.workingCopy, dir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read directory %s (%w)", dir, err)
	}
	var destination []providertypes.ProviderDocItem
	for _, item := range items {
		if item.IsDir() {
			continue
		}

		doc, err := s.readDocFile(dir, item, language, kind)
		if err != nil {
			return nil, err
		}
		if doc != nil {
			destination = append(destination, *doc)
		}

		select {
		case <-s.ctx.Done():
			return nil, s.ctx.Err()
		default:
		}
	}
	return destination, nil
}

func (s source) readDocFile(dir string, item fs.DirEntry, language providertypes.CDKTFLanguage, itemKind providertypes.DocItemKind) (*providertypes.ProviderDocItem, error) {
	fn := path.Join(dir, item.Name())
	name := normalizeName(path.Base(fn))
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
	if err := providertypes.DocItemName(name).Validate(); err != nil {
		s.logger.WarnContext(s.ctx, "Invalid document item name %s, skipping... (%v)", name, err)
		return nil, nil
	}

	stat, err := item.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to stat %s (%w)", fn, err)
	}
	var contents []byte
	if stat.Size() > maxFileSize {
		contents, err = s.getFileTooLargeError(fn)
		if err != nil {
			return nil, err
		}
	} else {
		contents, err = s.readFile(fn)
		if err != nil {
			return nil, err
		}
	}

	doc, err := s.ExtractFrontmatterPermissively(contents)
	if err != nil {
		s.logger.InfoContext(s.ctx, "Failed to extract frontmatter from %s (%v)", fn, err)
	}

	doc.Name = providertypes.DocItemName(name)
	doc.EditLink = fmt.Sprintf("%s/blob/%s/%s", s.raw.RepositoryURL(), s.version, fn)

	if doc.Title == "" {
		doc.Title = string(doc.Name)
	}

	if language == "" {
		if err := s.storage.StoreProviderDocItem(
			s.ctx,
			providertypes.Addr(s.raw),
			s.version,
			itemKind,
			providertypes.DocItemName(name),
			contents,
		); err != nil {
			return nil, fmt.Errorf("failed to store documentation item %s (%w)", name, err)
		}

	} else {
		if err := s.storage.StoreProviderCDKTFDocItem(
			s.ctx,
			providertypes.Addr(s.raw),
			s.version,
			language,
			itemKind,
			providertypes.DocItemName(name),
			contents,
		); err != nil {
			return nil, fmt.Errorf("failed to store CDKTF documentation item %s for language %s (%w)", name, language, err)
		}
	}

	// TODO cam72cam writefile!
	//

	return doc, nil
}

func (s source) readFile(fn string) ([]byte, error) {
	fh, err := os.Open(filepath.Join(s.workingCopy, fn))
	if err != nil {
		return nil, fmt.Errorf("failed to read %s (%w)", fn, err)
	}
	defer func() {
		// Read-only file, ignore errors
		_ = fh.Close()
	}()
	contents, err := io.ReadAll(fh)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s (%w)", fn, err)
	}
	return contents, nil
}

func (s source) getFileTooLargeError(fn string) ([]byte, error) {
	wr := &bytes.Buffer{}
	viewURL := fmt.Sprintf("%s/blob/%s/%s/%s", s.raw.RepositoryURL(), s.version, fn)
	if err := tplTooLarge.Execute(wr, viewURL); err != nil {
		return nil, fmt.Errorf("failed to render template (%w)", err)
	}
	return wr.Bytes(), nil
}

func (s source) extractFrontmatter(contents []byte) (*providertypes.ProviderDocItem, error) {
	scanner := bufio.NewScanner(bytes.NewReader(contents))
	var frontMatterLines []string
	capturing := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == "---" {
			if capturing {
				break
			}
			capturing = true
			continue
		}

		if capturing {
			frontMatterLines = append(frontMatterLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read front matter (%w)", err)
	}

	frontMatterString := strings.Join(frontMatterLines, "\n")
	var yamlTarget struct {
		Title       string `yaml:"page_title"`  // The page title taken from the frontmatter
		Subcategory string `yaml:"subcategory"` // The subcategory of the resource, data source, or function, taken from the frontmatter
		Description string `yaml:"description"` // The description of the resource.
	}
	if err := yaml.Unmarshal([]byte(frontMatterString), &yamlTarget); err != nil {
		return nil, fmt.Errorf("failed to unmarshal front matter (%w)", err)
	}
	return &providertypes.ProviderDocItem{
		Title:       yamlTarget.Title,
		Subcategory: yamlTarget.Subcategory,
		Description: yamlTarget.Description,
	}, nil
}

func (s source) ExtractFrontmatterPermissively(contents []byte) (*providertypes.ProviderDocItem, error) {
	// Attempt to extract it as YAML first
	if doc, err := s.extractFrontmatter(contents); err == nil {
		return doc, nil
	}

	// otherwise, go again and this time be as loose as we can, parsing it manually :(
	scanner := bufio.NewScanner(bytes.NewReader(contents))
	isCapturing := false
	currentKey := ""
	multilineBuffers := make(map[string]*strings.Builder)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == "---" {
			if isCapturing {
				// end of frontmatter
				break
			}
			isCapturing = true
			continue
		}

		if isCapturing {
			trimmedLine := strings.TrimSpace(line)

			// Check if the line contains a key-value pair or is part of a multiline value
			if strings.Contains(trimmedLine, ":") && !strings.HasPrefix(trimmedLine, "-") {
				// Finalize the current key's value if we're switching to a new key
				if currentKey != "" {
					multilineBuffers[currentKey].WriteString("\n")
				}

				parts := strings.SplitN(trimmedLine, ":", 2)
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Handle single-line key-value pairs or start of a multiline value
				currentKey = key
				multilineBuffers[currentKey] = &strings.Builder{}
				if value == "|-" || value == "|" {
					continue
				}
				multilineBuffers[currentKey].WriteString(value)
			} else if currentKey != "" && multilineBuffers[currentKey] != nil {
				// Handle continuation of a multiline value
				if trimmedLine == "" {
					continue
				}
				multilineBuffers[currentKey].WriteString("\n" + trimmedLine)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read front matter (%w)", err)
	}

	var doc providertypes.ProviderDocItem

	// Assign the accumulated values to the docItem fields
	for key, buffer := range multilineBuffers {
		value := strings.Trim(buffer.String(), "\n ")
		switch key {
		case "page_title":
			doc.Title = strings.Trim(value, `"'`)
		case "subcategory":
			doc.Subcategory = strings.Trim(value, `"'`)
		case "description":
			doc.Description = strings.Trim(value, `"'`)
		}
	}

	return &doc, nil
}
