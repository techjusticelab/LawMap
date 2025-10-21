package leginfo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"lawmap/internal/pkg/log"
)

// Fetcher retrieves content from the California LegInfo website.
type Fetcher struct {
	baseURL    string
	httpClient *http.Client
	rateLimit  time.Duration
	logger     *log.Logger
	lastFetch  time.Time
	code       string // Code to fetch (e.g., "CIV", "PEN")
	maxSections int   // Maximum sections to fetch (0 = all)
}

// FetcherConfig holds configuration for the fetcher.
type FetcherConfig struct {
	BaseURL           string
	RateLimitPerSec   float64
	TimeoutSeconds    int
	UserAgent         string
	Code              string // Which code to fetch
	MaxSections       int    // Maximum sections (0 = all, for testing use small number)
}

// NewFetcher creates a new LegInfo fetcher.
func NewFetcher(cfg FetcherConfig) *Fetcher {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://leginfo.legislature.ca.gov"
	}
	if cfg.TimeoutSeconds == 0 {
		cfg.TimeoutSeconds = 30
	}
	if cfg.RateLimitPerSec == 0 {
		cfg.RateLimitPerSec = 2.0
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "LawMap/0.1 (+https://github.com/yourorg/lawmap)"
	}
	if cfg.Code == "" {
		cfg.Code = "CIV" // Default to Civil Code
	}

	rateLimit := time.Duration(float64(time.Second) / cfg.RateLimitPerSec)

	return &Fetcher{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.TimeoutSeconds) * time.Second,
		},
		rateLimit:   rateLimit,
		logger:      log.Default().WithField("component", "leginfo-fetcher"),
		code:        cfg.Code,
		maxSections: cfg.MaxSections,
	}
}

// Name returns the fetcher name.
func (f *Fetcher) Name() string {
	return "leginfo-fetcher"
}

// Extract fetches a complete code from LegInfo.
// It fetches the table of contents, parses it to get all sections, then fetches each section.
func (f *Fetcher) Extract(ctx context.Context) ([]byte, error) {
	f.logger.Info("Starting code extraction", map[string]any{
		"code":         f.code,
		"max_sections": f.maxSections,
	})

	// Step 1: Fetch table of contents
	tocHTML, err := f.FetchCode(ctx, f.code)
	if err != nil {
		return nil, fmt.Errorf("fetch code TOC: %w", err)
	}

	// Step 2: Parse TOC to get section list
	tocParser := NewTOCParser()
	sections, err := tocParser.ExtractSectionList(string(tocHTML), f.code)
	if err != nil {
		return nil, fmt.Errorf("parse code TOC: %w", err)
	}

	f.logger.Info("Extracted section list from TOC", map[string]any{
		"total_sections": len(sections),
	})

	// If TOC parsing returned no sections (likely JavaScript-rendered page),
	// fall back to range-based fetching
	if len(sections) == 0 {
		f.logger.Warn("TOC returned no sections (likely JavaScript-rendered page), using range-based fetching")

		start, end := GetCommonSectionRanges(f.code)
		rangeFetcher := NewSectionRangeFetcher(FetcherConfig{
			BaseURL:         f.baseURL,
			RateLimitPerSec: 2.0,
			TimeoutSeconds:  30,
			UserAgent:       "LawMap/0.1",
			Code:            f.code,
			MaxSections:     f.maxSections,
		}, start, end)

		return rangeFetcher.ExtractByRange(ctx)
	}

	// Apply max sections limit if set (useful for testing)
	if f.maxSections > 0 && len(sections) > f.maxSections {
		f.logger.Info("Limiting sections for testing", map[string]any{
			"limit": f.maxSections,
		})
		sections = sections[:f.maxSections]
	}

	// Step 3: Fetch each section
	var sectionData []SectionData
	for i, section := range sections {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		f.logger.Debug("Fetching section", map[string]any{
			"section": section.Number,
			"progress": fmt.Sprintf("%d/%d", i+1, len(sections)),
		})

		html, err := f.FetchSection(ctx, f.code, section.Number)
		if err != nil {
			f.logger.Error("Failed to fetch section", map[string]any{
				"section": section.Number,
				"error":   err.Error(),
			})
			// Continue with next section instead of failing completely
			continue
		}

		sectionData = append(sectionData, SectionData{
			Number:       section.Number,
			HTML:         string(html),
			URL:          section.URL,
			FetchedAt:    time.Now().Format(time.RFC3339),
			HierarchyCtx: section.HierarchyCtx,
		})

		// Progress logging every 10 sections
		if (i+1)%10 == 0 {
			f.logger.Info("Fetch progress", map[string]any{
				"fetched": i + 1,
				"total":   len(sections),
			})
		}
	}

	// Step 4: Package data into ExtractedData structure
	extracted := ExtractedData{
		Code:         f.code,
		Jurisdiction: "CA",
		TOCHTML:      string(tocHTML),
		Sections:     sectionData,
		Metadata: map[string]any{
			"total_sections":   len(sections),
			"fetched_sections": len(sectionData),
			"timestamp":        time.Now().Format(time.RFC3339),
		},
	}

	// Step 5: Serialize to JSON
	data, err := json.Marshal(extracted)
	if err != nil {
		return nil, fmt.Errorf("marshal extracted data: %w", err)
	}

	f.logger.Info("Extraction complete", map[string]any{
		"code":            f.code,
		"sections_total":  len(sections),
		"sections_fetched": len(sectionData),
		"data_size":       len(data),
	})

	return data, nil
}

// FetchSection retrieves a specific code section from LegInfo.
// Example URL: https://leginfo.legislature.ca.gov/faces/codes_displaySection.xhtml?lawCode=CIV&sectionNum=3342
func (f *Fetcher) FetchSection(ctx context.Context, code, section string) ([]byte, error) {
	// Respect rate limiting
	if !f.lastFetch.IsZero() {
		elapsed := time.Since(f.lastFetch)
		if elapsed < f.rateLimit {
			time.Sleep(f.rateLimit - elapsed)
		}
	}

	url := fmt.Sprintf("%s/faces/codes_displaySection.xhtml?lawCode=%s&sectionNum=%s",
		f.baseURL, code, section)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "LawMap/0.1")

	f.logger.Debug("Fetching section", map[string]any{"url": url})

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	f.lastFetch = time.Now()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %d for %s", resp.StatusCode, url)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	f.logger.Info("Fetched section", map[string]any{
		"code":    code,
		"section": section,
		"bytes":   len(data),
	})

	return data, nil
}

// FetchCode retrieves information about a specific code (e.g., all sections).
// This would typically parse the code's table of contents.
func (f *Fetcher) FetchCode(ctx context.Context, code string) ([]byte, error) {
	url := fmt.Sprintf("%s/faces/codes_displayexpandedbranch.xhtml?lawCode=%s&division=&title=&part=&chapter=&article=",
		f.baseURL, code)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "LawMap/0.1")

	f.logger.Debug("Fetching code TOC", map[string]any{"url": url, "code": code})

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	f.lastFetch = time.Now()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %d for %s", resp.StatusCode, url)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	f.logger.Info("Fetched code TOC", map[string]any{"code": code, "bytes": len(data)})
	return data, nil
}
