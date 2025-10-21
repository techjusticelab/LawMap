package leginfo

// CodeStructure represents the complete hierarchy of a California code.
type CodeStructure struct {
	Code         string             // e.g., "CIV", "PEN"
	CodeName     string             // e.g., "Civil Code"
	Jurisdiction string             // "CA"
	Titles       []TitleStructure   // Top-level divisions
	Sections     []SectionReference // Flat list of all sections for fetching
}

// TitleStructure represents a Title within a code.
type TitleStructure struct {
	Number   int                 // e.g., 2 for "Title 2"
	Name     string              // e.g., "Real or Immovable Property"
	Divisions []DivisionStructure // Divisions within this title (if any)
	Chapters []ChapterStructure  // Chapters within this title (if no divisions)
}

// DivisionStructure represents a Division within a title.
type DivisionStructure struct {
	Number   int                // e.g., 1 for "Division 1"
	Name     string             // e.g., "Persons"
	Parts    []PartStructure    // Parts within this division (if any)
	Chapters []ChapterStructure // Chapters within this division (if no parts)
}

// PartStructure represents a Part within a division.
type PartStructure struct {
	Number   int                // e.g., 1 for "Part 1"
	Name     string             // e.g., "Persons With Unsound Mind"
	Chapters []ChapterStructure // Chapters within this part
}

// ChapterStructure represents a Chapter.
type ChapterStructure struct {
	Number   int                 // e.g., 2 for "Chapter 2"
	Name     string              // e.g., "Dogs"
	Articles []ArticleStructure  // Articles within this chapter (if any)
	Sections []SectionReference  // Sections directly in this chapter
}

// ArticleStructure represents an Article within a chapter.
type ArticleStructure struct {
	Number   int                // e.g., 1 for "Article 1"
	Name     string             // e.g., "General Provisions"
	Sections []SectionReference // Sections within this article
}

// SectionReference represents a reference to a code section for fetching.
type SectionReference struct {
	Number       string           // e.g., "3342", "3342.5", "924(e)"
	URL          string           // Full URL to fetch this section
	HierarchyCtx HierarchyContext // Context for building canonical ID
}

// HierarchyContext contains the hierarchical position of a section.
type HierarchyContext struct {
	Code     string // e.g., "CIV"
	Title    int    // Title number (0 if not applicable)
	Division int    // Division number (0 if not applicable)
	Part     int    // Part number (0 if not applicable)
	Chapter  int    // Chapter number (0 if not applicable)
	Article  int    // Article number (0 if not applicable)
}

// ExtractedData represents the data extracted by the fetcher.
type ExtractedData struct {
	Code         string             // e.g., "CIV"
	Jurisdiction string             // "CA"
	TOCHTML      string             // Raw HTML of table of contents
	Sections     []SectionData      // All fetched sections
	Metadata     map[string]any     // Additional metadata
}

// SectionData represents a fetched section's raw data.
type SectionData struct {
	Number       string           // e.g., "3342"
	HTML         string           // Raw HTML content
	URL          string           // URL fetched from
	FetchedAt    string           // RFC3339 timestamp
	HierarchyCtx HierarchyContext // Position in code hierarchy
}

// ParsedSection represents a fully parsed section ready for graph insertion.
type ParsedSection struct {
	Number        string           // e.g., "3342"
	Title         string           // e.g., "Dog bite liability"
	Text          string           // Full section text
	Citation      string           // e.g., "CIV § 3342"
	EffectiveDate string           // Effective date if available
	History       string           // Legislative history/derivation
	HierarchyCtx  HierarchyContext // Hierarchical context
}

// HierarchyLayer represents a single layer in the code hierarchy.
type HierarchyLayer struct {
	ID     string         // Canonical ID (e.g., "CA:CIV:T02")
	Label  string         // Node label (CODE, TITLE, DIVISION, CHAPTER, etc.)
	Title  string         // Display name
	Props  map[string]any // Additional properties
	Order  int            // Order within parent
}
