package providerdocsource

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"

	"github.com/opentofu/libregistry/logger"
	"github.com/opentofu/libregistry/vcs"
	"gopkg.in/yaml.v3"

	"github.com/opentofu/registry-ui/internal/license"
	"github.com/opentofu/registry-ui/internal/license/vcslinkfetcher"
	"github.com/opentofu/registry-ui/internal/providerindex/providertypes"
)

// maxFileSize limits the amount of data read from the docs to 1MB to prevent memory-based DoS.
const maxFileSize = 1024 * 1024

//go:embed err_file_too_large.md.tpl
var tooLargeError []byte

var tplTooLarge = template.Must(template.New("").Parse(string(tooLargeError)))

func New(licenseDetector license.Detector, logger logger.Logger) (API, error) {
	return &source{licenseDetector: licenseDetector, logger: logger.WithName("Provider doc source")}, nil
}

type source struct {
	licenseDetector license.Detector
	logger          logger.Logger
}

func (s source) Describe(ctx context.Context, workingCopy vcs.WorkingCopy) (ProviderDocumentation, error) {
	licenses, err := s.licenseDetector.Detect(
		ctx,
		workingCopy,
		license.WithLinkFetcher(vcslinkfetcher.Fetcher(
			ctx,
			workingCopy.Repository(),
			workingCopy.Version(),
			workingCopy.Client(),
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to detect licenses")
	}

	link, err := workingCopy.Client().GetVersionBrowseURL(ctx, workingCopy.Repository(), workingCopy.Version())
	if err != nil {
		s.logger.Warn(ctx, "Cannot determine repository link for provider %s version %s (%v)", workingCopy.Repository(), workingCopy.Version(), err)
	}

	doc := &documentation{
		docs:     newProviderDoc(),
		link:     link,
		cdktf:    map[string]Documentation{},
		licenses: licenses,
	}
	if !doc.licenses.IsRedistributable() {
		return doc, nil
	}
	for _, dir := range []string{
		path.Join("website", "docs"),
		"docs",
	} {
		// Note: we prefer website/docs over docs and not include docs because some providers
		// (such as AWS) use the docs directory for internal documentation.
		foundDocs, err := s.scrapeDir(ctx, dir, workingCopy, doc)
		if err != nil {
			return nil, err
		}
		if foundDocs {
			break
		}
	}
	return doc, nil
}

var normalizeRe = regexp.MustCompile("[^a-zA-Z0-9-_.]")

func normalizeName(name string) string {
	return normalizeRe.ReplaceAllString(name, "")
}

func (s source) scrapeDir(ctx context.Context, dir string, workingCopy vcs.WorkingCopy, doc *documentation) (bool, error) {
	if doc.docs == nil {
		doc.docs = newProviderDoc()
	}
	_, err := workingCopy.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if err := s.scrapeDocumentation(ctx, workingCopy, dir, doc.docs); err != nil {
		return true, fmt.Errorf("failed to scrape documentation directory (%w)", err)
	}
	cdktfDir := path.Join(dir, "cdktf")
	cdktfItems, err := workingCopy.ReadDir(cdktfDir)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return true, fmt.Errorf("failed to list CDKTF documentation directory (%w)", err)
	}
	for _, item := range cdktfItems {
		if !item.IsDir() {
			continue
		}
		lang := providertypes.CDKTFLanguage(item.Name())
		if err := lang.Validate(); err != nil {
			s.logger.Debug(ctx, "%s is not a valid CDKTF language, skipping... (%v)", item.Name(), err)
			continue
		}
		cdktfDoc := newProviderDoc()
		doc.cdktf[string(lang)] = cdktfDoc
		if err := s.scrapeDocumentation(ctx, workingCopy, path.Join(cdktfDir, string(lang)), cdktfDoc); err != nil {
			return true, fmt.Errorf("failed to scrape CDKTF documentation for %s (%w)", item.Name(), err)
		}
	}
	return true, nil
}

func (s source) scrapeDocumentation(ctx context.Context, workingCopy vcs.WorkingCopy, dir string, doc *providerDoc) error {
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

	if err := s.extractRootDoc(ctx, workingCopy, dir, doc); err != nil {
		return err
	}

	for _, scrapeType := range []struct {
		dir         string
		destination *map[string]DocumentationItem
	}{
		{"r", &doc.resources},
		{"resources", &doc.resources},
		{"d", &doc.datasources},
		{"data-sources", &doc.datasources},
		{"f", &doc.functions},
		{"functions", &doc.functions},
		{"guides", &doc.guides},
	} {
		sourceDir := path.Join(dir, scrapeType.dir)
		if err := s.scrapeType(ctx, workingCopy, sourceDir, scrapeType.destination); err != nil {
			return fmt.Errorf("failed to scrape %s", sourceDir)
		}
	}
	return nil
}

func (s source) extractRootDoc(ctx context.Context, workingCopy vcs.WorkingCopy, dir string, doc *providerDoc) error {
	items, err := workingCopy.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read directory %s (%w)", dir, err)
	}
	for _, item := range items {
		if item.IsDir() {
			continue
		}
		if strings.HasPrefix(item.Name(), "index.") {
			root, err := s.readDocFile(ctx, dir, item, workingCopy)
			if err != nil {
				return fmt.Errorf("failed to parse %s (%w)", item.Name(), err)
			}
			if root != nil {
				doc.root = root
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

func (s source) scrapeType(ctx context.Context, workingCopy vcs.WorkingCopy, dir string, destination *map[string]DocumentationItem) error {
	items, err := workingCopy.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read directory %s (%w)", dir, err)
	}
	for _, item := range items {
		if item.IsDir() {
			continue
		}

		doc, err := s.readDocFile(ctx, dir, item, workingCopy)
		if err != nil {
			return err
		}
		if doc != nil {
			(*destination)[doc.Name] = doc
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	return nil
}

func (s source) readDocFile(ctx context.Context, dir string, item fs.DirEntry, workingCopy vcs.WorkingCopy) (*docItem, error) {
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
		s.logger.Warn(ctx, "Invalid document item name %s, skipping... (%v)", name, err)
		return nil, nil
	}

	stat, err := item.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to stat %s (%w)", fn, err)
	}
	var contents []byte
	if stat.Size() > maxFileSize {
		contents, err = s.getFileTooLargeError(ctx, workingCopy, fn)
		if err != nil {
			return nil, err
		}
	} else {
		contents, err = s.readFile(ctx, workingCopy, fn)
		if err != nil {
			return nil, err
		}
	}

	doc := &docItem{
		Name:        name,
		Title:       "",
		Subcategory: "",
		Description: "",
		EditLink:    "",
		contents:    contents,
	}
	if err := s.ExtractFrontmatterPermissively(ctx, contents, doc); err != nil {
		s.logger.Info(ctx, "Failed to extract frontmatter from %s (%v)", fn, err)
	}
	doc.EditLink, err = workingCopy.Client().GetFileViewURL(ctx, workingCopy.Repository(), workingCopy.Version(), fn)
	if err != nil {
		s.logger.Warn(ctx, "Cannot determine edit link for %s (%v)", fn, err)
	}
	return doc, nil
}

func (s source) readFile(_ context.Context, workingCopy vcs.WorkingCopy, fn string) ([]byte, error) {
	fh, err := workingCopy.Open(fn)
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

func (s source) getFileTooLargeError(ctx context.Context, workingCopy vcs.WorkingCopy, fn string) ([]byte, error) {
	wr := &bytes.Buffer{}
	vcsClient := workingCopy.Client()
	viewURL, err := vcsClient.GetFileViewURL(ctx, workingCopy.Repository(), workingCopy.Version(), fn)
	if err != nil {
		return nil, fmt.Errorf("failed to get view URL for %s (%w)", fn, err)
	}
	if err := tplTooLarge.Execute(wr, viewURL); err != nil {
		return nil, fmt.Errorf("failed to render template (%w)", err)
	}
	return wr.Bytes(), nil
}

func (s source) extractFrontmatter(_ context.Context, contents []byte, doc *docItem) error {
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
		return fmt.Errorf("failed to read front matter (%w)", err)
	}

	frontMatterString := strings.Join(frontMatterLines, "\n")
	if err := yaml.Unmarshal([]byte(frontMatterString), &doc); err != nil {
		return fmt.Errorf("failed to unmarshal front matter (%w)", err)
	}
	return nil
}

func (s source) ExtractFrontmatterPermissively(ctx context.Context, contents []byte, doc *docItem) error {
	// Attempt to extract it as YAML first
	if err := s.extractFrontmatter(ctx, contents, doc); err == nil {
		return nil
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
		return fmt.Errorf("failed to read front matter (%w)", err)
	}

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

	return nil
}
