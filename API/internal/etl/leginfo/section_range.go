package leginfo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// SectionRangeFetcher fetches sections by trying a range of section numbers.
// This is a workaround for JavaScript-rendered TOC pages.
type SectionRangeFetcher struct {
	*Fetcher
	startSection int
	endSection   int
}

// NewSectionRangeFetcher creates a fetcher that tries a range of section numbers.
func NewSectionRangeFetcher(cfg FetcherConfig, startSection, endSection int) *SectionRangeFetcher {
	return &SectionRangeFetcher{
		Fetcher:      NewFetcher(cfg),
		startSection: startSection,
		endSection:   endSection,
	}
}

// ExtractByRange fetches sections by trying each number in the range.
// It skips sections that return 404 or errors.
func (f *SectionRangeFetcher) ExtractByRange(ctx context.Context) ([]byte, error) {
	f.logger.Info("Starting range-based extraction", map[string]any{
		"code":  f.code,
		"start": f.startSection,
		"end":   f.endSection,
	})

	var sections []SectionData
	attempted := 0
	successful := 0

	for sectionNum := f.startSection; sectionNum <= f.endSection; sectionNum++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Apply max sections limit if set
		if f.maxSections > 0 && successful >= f.maxSections {
			f.logger.Info("Reached max sections limit", map[string]any{
				"limit": f.maxSections,
			})
			break
		}

		attempted++
		sectionStr := fmt.Sprintf("%d", sectionNum)

		html, err := f.FetchSection(ctx, f.code, sectionStr)
		if err != nil {
			// Check if it's a 404 (section doesn't exist)
			if isNotFound(err) {
				f.logger.Debug("Section not found (skipping)", map[string]any{
					"section": sectionStr,
				})
				continue
			}

			f.logger.Warn("Failed to fetch section", map[string]any{
				"section": sectionStr,
				"error":   err.Error(),
			})
			continue
		}

		sections = append(sections, SectionData{
			Number:    sectionStr,
			HTML:      string(html),
			URL:       fmt.Sprintf("https://leginfo.legislature.ca.gov/faces/codes_displaySection.xhtml?lawCode=%s&sectionNum=%s", f.code, sectionStr),
			FetchedAt: time.Now().Format(time.RFC3339),
			HierarchyCtx: HierarchyContext{
				Code: f.code,
			},
		})
		successful++

		// Progress logging
		if successful%10 == 0 {
			f.logger.Info("Range fetch progress", map[string]any{
				"attempted":  attempted,
				"successful": successful,
				"current":    sectionNum,
			})
		}
	}

	// Package into ExtractedData
	extracted := ExtractedData{
		Code:         f.code,
		Jurisdiction: "CA",
		Sections:     sections,
		Metadata: map[string]any{
			"fetch_method":     "section_range",
			"start_section":    f.startSection,
			"end_section":      f.endSection,
			"attempted":        attempted,
			"successful":       successful,
			"timestamp":        time.Now().Format(time.RFC3339),
		},
	}

	data, err := json.Marshal(extracted)
	if err != nil {
		return nil, fmt.Errorf("marshal extracted data: %w", err)
	}

	f.logger.Info("Range extraction complete", map[string]any{
		"code":       f.code,
		"attempted":  attempted,
		"successful": successful,
		"data_size":  len(data),
	})

	return data, nil
}

// isNotFound checks if an error is due to a section not being found.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	// Check for HTTP 404 errors
	return containsString(err.Error(), "404") || containsString(err.Error(), "not found")
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetCommonSectionRanges returns common section number ranges for different codes.
func GetCommonSectionRanges(code string) (start, end int) {
	ranges := map[string][2]int{
		"BPC":  {100, 30000},   // Business & Professions Code
		"CIV":  {1, 10000},     // Civil Code
		"CCP":  {1, 2500},      // Code of Civil Procedure
		"EVID": {1, 2000},      // Evidence Code
		"PEN":  {1, 13000},     // Penal Code
		"VEH":  {1, 45000},     // Vehicle Code (very large)
		"FAM":  {1, 10000},     // Family Code
		"GOV":  {1, 100000},    // Government Code (huge)
		"HSC":  {1, 150000},    // Health & Safety Code (huge)
		"LAB":  {1, 8000},      // Labor Code
		"PROB": {1, 22000},     // Probate Code
		"WIC":  {1, 20000},     // Welfare & Institutions Code
	}

	if r, ok := ranges[code]; ok {
		return r[0], r[1]
	}

	// Default range
	return 1, 10000
}
