package leginfo

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"lawmap/internal/pkg/log"
)

// BulkDownloader handles bulk downloads from LegInfo.
type BulkDownloader struct {
	baseURL    string
	httpClient *http.Client
	logger     *log.Logger
}

// NewBulkDownloader creates a new bulk downloader.
func NewBulkDownloader() *BulkDownloader {
	return &BulkDownloader{
		baseURL: "https://leginfo.legislature.ca.gov",
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Bulk downloads may be large
		},
		logger: log.Default().WithField("component", "leginfo-bulk"),
	}
}

// DownloadCode downloads a complete code as bulk data.
// LegInfo provides downloadable databases at https://leginfo.legislature.ca.gov/
func (b *BulkDownloader) DownloadCode(ctx context.Context, code string) (*ExtractedData, error) {
	b.logger.Info("Starting bulk download", map[string]any{"code": code})

	// For now, this is a placeholder that will use the section-by-section approach
	// The actual bulk download URL structure needs to be discovered from LegInfo
	// Typical formats:
	// - ZIP file with XML/HTML files for each section
	// - Single XML file with all sections
	// - Text files organized by hierarchy

	// TODO: Implement actual bulk download once we know the URL structure
	// For now, return an error to fall back to section fetching
	return nil, fmt.Errorf("bulk download not yet implemented - use section-by-section fetching")
}

// DownloadAndExtractZip downloads and extracts a ZIP file.
func (b *BulkDownloader) DownloadAndExtractZip(ctx context.Context, url string) (map[string][]byte, error) {
	b.logger.Info("Downloading ZIP", map[string]any{"url": url})

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "LawMap/0.1 (+https://github.com/yourorg/lawmap)")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Read entire response into memory
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	b.logger.Info("Downloaded ZIP", map[string]any{"bytes": len(data)})

	// Extract ZIP
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}

	files := make(map[string][]byte)
	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			b.logger.Warn("Failed to open file in zip", map[string]any{
				"file":  file.Name,
				"error": err.Error(),
			})
			continue
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			b.logger.Warn("Failed to read file in zip", map[string]any{
				"file":  file.Name,
				"error": err.Error(),
			})
			continue
		}

		files[file.Name] = content
	}

	b.logger.Info("Extracted ZIP", map[string]any{"files": len(files)})
	return files, nil
}

// ParseBulkXML parses bulk XML data into ExtractedData format.
func (b *BulkDownloader) ParseBulkXML(code string, files map[string][]byte) (*ExtractedData, error) {
	extracted := &ExtractedData{
		Code:         code,
		Jurisdiction: "CA",
		Sections:     []SectionData{},
		Metadata: map[string]any{
			"source":    "bulk_download",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	// Parse each file to extract section data
	for filename, content := range files {
		// Skip non-section files
		if !strings.Contains(filename, ".xml") && !strings.Contains(filename, ".html") {
			continue
		}

		// Extract section number from filename
		// Common patterns: "BPC_7512.xml", "section_7512.html", etc.
		sectionNum := b.extractSectionFromFilename(filename)
		if sectionNum == "" {
			b.logger.Debug("Could not extract section number", map[string]any{
				"filename": filename,
			})
			continue
		}

		extracted.Sections = append(extracted.Sections, SectionData{
			Number:    sectionNum,
			HTML:      string(content),
			URL:       fmt.Sprintf("bulk_download/%s", filename),
			FetchedAt: time.Now().Format(time.RFC3339),
			HierarchyCtx: HierarchyContext{
				Code: code,
				// Hierarchy will be extracted from content
			},
		})
	}

	b.logger.Info("Parsed bulk data", map[string]any{
		"code":     code,
		"sections": len(extracted.Sections),
	})

	return extracted, nil
}

// extractSectionFromFilename attempts to extract a section number from a filename.
func (b *BulkDownloader) extractSectionFromFilename(filename string) string {
	base := filepath.Base(filename)
	base = strings.TrimSuffix(base, filepath.Ext(base))

	// Try various patterns
	patterns := []string{
		`_(\d+[\w\.]*)$`,           // BPC_7512, section_7512.5
		`section[\s_-]?(\d+[\w\.]*)`, // section7512, section_7512
		`(\d+[\w\.]*)`,               // Just numbers
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if m := re.FindStringSubmatch(base); m != nil && len(m) > 1 {
			return m[1]
		}
	}

	return ""
}

// FetchCodeWithBulkFallback tries bulk download first, falls back to section-by-section.
func FetchCodeWithBulkFallback(ctx context.Context, cfg FetcherConfig) ([]byte, error) {
	logger := log.Default()

	// Try bulk download first
	downloader := NewBulkDownloader()
	bulkData, err := downloader.DownloadCode(ctx, cfg.Code)
	if err == nil {
		logger.Info("Using bulk download")
		return json.Marshal(bulkData)
	}

	logger.Info("Bulk download not available, falling back to section fetching", map[string]any{
		"error": err.Error(),
	})

	// Fall back to regular section-by-section fetching
	fetcher := NewFetcher(cfg)
	return fetcher.Extract(ctx)
}
