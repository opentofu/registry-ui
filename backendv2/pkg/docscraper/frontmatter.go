package docscraper

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func (s *Scraper) extractFrontmatterPermissively(ctx context.Context, contents []byte, doc *DocItem) error {
	// Attempt to extract it as YAML first
	if err := s.extractFrontmatter(ctx, contents, doc); err == nil {
		return nil
	}

	// Otherwise, parse manually for problematic frontmatter
	scanner := bufio.NewScanner(bytes.NewReader(contents))
	isCapturing := false
	currentKey := ""
	multilineBuffers := make(map[string]*strings.Builder)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == "---" {
			if isCapturing {
				break
			}
			isCapturing = true
			continue
		}

		if isCapturing {
			trimmedLine := strings.TrimSpace(line)

			if strings.Contains(trimmedLine, ":") && !strings.HasPrefix(trimmedLine, "-") {
				if currentKey != "" {
					multilineBuffers[currentKey].WriteString("\n")
				}

				parts := strings.SplitN(trimmedLine, ":", 2)
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				currentKey = key
				multilineBuffers[currentKey] = &strings.Builder{}
				if value == "|-" || value == "|" {
					continue
				}
				multilineBuffers[currentKey].WriteString(value)
			} else if currentKey != "" && multilineBuffers[currentKey] != nil {
				if trimmedLine == "" {
					continue
				}
				multilineBuffers[currentKey].WriteString("\n" + trimmedLine)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read front matter: %w", err)
	}

	// Assign values to DocItem
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

func (s *Scraper) extractFrontmatter(ctx context.Context, contents []byte, doc *DocItem) (err error) {
	// Recover from YAML panics (like duplicate keys)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("YAML parsing panic: %v", r)
		}
	}()

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
		return fmt.Errorf("failed to read front matter: %w", err)
	}

	frontMatterString := strings.Join(frontMatterLines, "\n")
	if err := yaml.Unmarshal([]byte(frontMatterString), &doc); err != nil {
		return fmt.Errorf("failed to unmarshal front matter (content: %q): %w", frontMatterString, err)
	}

	return nil
}
