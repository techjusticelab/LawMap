package parse

import (
	"fmt"
	"regexp"
	"strings"
)

// CodeKey maps common code abbreviations to canonical uppercase keys.
var CodeKey = map[string]string{
	// California
	"bpc":  "BPC",  // Business and Professions Code
	"civ":  "CIV",  // Civil Code
	"ccp":  "CCP",  // Code of Civil Procedure
	"com":  "COM",  // Commercial Code
	"corp": "CORP", // Corporations Code
	"edc":  "EDC",  // Education Code
	"elec": "ELEC", // Elections Code
	"evid": "EVID", // Evidence Code
	"fam":  "FAM",  // Family Code
	"fin":  "FIN",  // Financial Code
	"fgc":  "FGC",  // Fish and Game Code
	"fac":  "FAC",  // Food and Agricultural Code
	"gov":  "GOV",  // Government Code
	"hnc":  "HNC",  // Harbors and Navigation Code
	"hsc":  "HSC",  // Health and Safety Code
	"ins":  "INS",  // Insurance Code
	"lab":  "LAB",  // Labor Code
	"mvc":  "MVC",  // Military and Veterans Code
	"pen":  "PEN",  // Penal Code
	"prob": "PROB", // Probate Code
	"pcc":  "PCC",  // Public Contract Code
	"prc":  "PRC",  // Public Resources Code
	"puc":  "PUC",  // Public Utilities Code
	"rtc":  "RTC",  // Revenue and Taxation Code
	"shc":  "SHC",  // Streets and Highways Code
	"uic":  "UIC",  // Unemployment Insurance Code
	"veh":  "VEH",  // Vehicle Code
	"wat":  "WAT",  // Water Code
	"wic":  "WIC",  // Welfare and Institutions Code
	"cons": "CONS", // Constitution
	"crc":  "CRC",  // California Rules of Court
	"ccr":  "CCR",  // California Code of Regulations

	// Full names
	"penal code":                          "PEN",
	"civil code":                          "CIV",
	"code of civil procedure":             "CCP",
	"evidence code":                       "EVID",
	"vehicle code":                        "VEH",
	"health and safety code":              "HSC",
	"welfare and institutions code":       "WIC",
	"business and professions code":       "BPC",

	// Federal
	"usc":   "USC",   // United States Code
	"cfr":   "CFR",   // Code of Federal Regulations
	"fr":    "FR",    // Federal Register
	"const": "CONST", // Constitution
	"frcp":  "FRCP",  // Federal Rules of Civil Procedure
	"fre":   "FRE",   // Federal Rules of Evidence
	"frcrp": "FRCRP", // Federal Rules of Criminal Procedure
	"frap":  "FRAP",  // Federal Rules of Appellate Procedure
	"ussg":  "USSG",  // U.S. Sentencing Guidelines
	"opn":   "OPN",   // Opinions (general)
}

// NormalizeCode normalizes a code abbreviation to its canonical uppercase form.
func NormalizeCode(code string) string {
	c := strings.ToLower(strings.TrimSpace(code))
	if canonical, ok := CodeKey[c]; ok {
		return canonical
	}
	// If not in map, just uppercase it
	return strings.ToUpper(c)
}

// NormalizeSection normalizes section numbers by removing extraneous punctuation.
// Examples: "§ 3342" -> "3342", "3342(a)" -> "3342(a)", "§§ 100-200" -> "100-200"
func NormalizeSection(section string) string {
	s := strings.TrimSpace(section)
	// Remove section symbol and leading/trailing spaces
	s = strings.ReplaceAll(s, "§", "")
	s = strings.ReplaceAll(s, "§§", "")
	s = strings.TrimSpace(s)
	return s
}

// CanonicalID builds a canonical ID from jurisdiction, code, and hierarchical components.
// Format: jurisdiction:code[:title][:chapter][:section]
// Example: CA:CIV:T02:CH02:§3342
func CanonicalID(jurisdiction, code string, parts ...string) string {
	jur := strings.ToUpper(strings.TrimSpace(jurisdiction))
	c := NormalizeCode(code)

	var components []string
	components = append(components, jur, c)
	for _, p := range parts {
		if p != "" {
			components = append(components, strings.TrimSpace(p))
		}
	}

	return strings.Join(components, ":")
}

// ParseCitation attempts to parse a natural language citation into structured components.
// Examples:
//   - "CIV § 3342" -> {Code: "CIV", Section: "3342"}
//   - "18 USC § 924(e)" -> {Code: "USC", Title: "18", Section: "924(e)"}
//   - "Cal. Const. art. I, § 13" -> {Code: "CONS", Article: "I", Section: "13"}
type Citation struct {
	Jurisdiction string
	Code         string
	Title        string
	Division     string
	Part         string
	Chapter      string
	Article      string
	Section      string
	Subsection   string
}

var (
	// Pattern for California citations: "CIV § 3342" or "Penal Code § 1538.5"
	reCACitation = regexp.MustCompile(`(?i)(BPC|CIV|CCP|COM|CORP|EDC|ELEC|EVID|FAM|FIN|FGC|FAC|GOV|HNC|HSC|INS|LAB|MVC|PEN|PROB|PCC|PRC|PUC|RTC|SHC|UIC|VEH|WAT|WIC|Penal\s+Code|Civil\s+Code|Code\s+of\s+Civil\s+Procedure|Evidence\s+Code|Vehicle\s+Code|Health\s+and\s+Safety\s+Code|Welfare\s+and\s+Institutions\s+Code)\s*§*\s*(\d+[\.\w\(\)]*)`)

	// Pattern for US Code: "18 USC § 924(e)" or "18 U.S.C. § 924(e)"
	reUSCCitation = regexp.MustCompile(`(?i)(\d+)\s+(USC|U\.S\.C\.)\s*§*\s*(\d+[\.\w\(\)]*)`)

	// Pattern for CFR: "28 CFR § 600.4" or "28 C.F.R. § 600.4"
	reCFRCitation = regexp.MustCompile(`(?i)(\d+)\s+(CFR|C\.F\.R\.)\s*§*\s*(\d+[\.\w]*)`)

	// Pattern for California Constitution: "Cal. Const. art. I, § 13"
	reCAConst = regexp.MustCompile(`(?i)Cal\.\s*Const\.\s*art\.\s*([IVX]+),\s*§\s*(\d+)`)

	// Pattern for US Constitution: "U.S. Const. amend. IV"
	reUSConst = regexp.MustCompile(`(?i)U\.S\.\s*Const\.\s*amend\.\s*([IVX]+)`)
)

// ParseCitationString attempts to parse a citation string into a Citation struct.
func ParseCitationString(s string) (*Citation, error) {
	s = strings.TrimSpace(s)

	// Try CA code pattern
	if m := reCACitation.FindStringSubmatch(s); m != nil {
		code := NormalizeCode(m[1])
		section := ""
		if len(m) > 2 {
			section = NormalizeSection(m[2])
		}
		return &Citation{
			Jurisdiction: "CA",
			Code:         code,
			Section:      section,
		}, nil
	}

	// Try USC pattern
	if m := reUSCCitation.FindStringSubmatch(s); m != nil {
		title := m[1]
		section := NormalizeSection(m[3])
		return &Citation{
			Jurisdiction: "US",
			Code:         "USC",
			Title:        title,
			Section:      section,
		}, nil
	}

	// Try CFR pattern
	if m := reCFRCitation.FindStringSubmatch(s); m != nil {
		title := m[1]
		section := NormalizeSection(m[3])
		return &Citation{
			Jurisdiction: "US",
			Code:         "CFR",
			Title:        title,
			Section:      section,
		}, nil
	}

	// Try CA Constitution
	if m := reCAConst.FindStringSubmatch(s); m != nil {
		article := m[1]
		section := m[2]
		return &Citation{
			Jurisdiction: "CA",
			Code:         "CONS",
			Article:      article,
			Section:      section,
		}, nil
	}

	// Try US Constitution
	if m := reUSConst.FindStringSubmatch(s); m != nil {
		amendment := m[1]
		return &Citation{
			Jurisdiction: "US",
			Code:         "CONST",
			Section:      amendment,
		}, nil
	}

	return nil, fmt.Errorf("unable to parse citation: %s", s)
}

// ToCanonicalID converts a Citation to a canonical ID string.
func (c *Citation) ToCanonicalID() string {
	parts := []string{}

	if c.Title != "" {
		parts = append(parts, fmt.Sprintf("T%s", c.Title))
	}
	if c.Division != "" {
		parts = append(parts, fmt.Sprintf("D%s", c.Division))
	}
	if c.Part != "" {
		parts = append(parts, fmt.Sprintf("P%s", c.Part))
	}
	if c.Chapter != "" {
		parts = append(parts, fmt.Sprintf("CH%s", c.Chapter))
	}
	if c.Article != "" {
		parts = append(parts, fmt.Sprintf("Art%s", c.Article))
	}
	if c.Section != "" {
		// Preserve section symbol for canonical IDs
		parts = append(parts, fmt.Sprintf("§%s", c.Section))
	}
	if c.Subsection != "" {
		parts = append(parts, fmt.Sprintf("(%s)", c.Subsection))
	}

	return CanonicalID(c.Jurisdiction, c.Code, parts...)
}
