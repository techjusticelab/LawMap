package parse

import (
	"testing"
)

func TestNormalizeCode(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"civ", "CIV"},
		{"CIV", "CIV"},
		{"  pen  ", "PEN"},
		{"USC", "USC"},
		{"cfr", "CFR"},
		{"unknown", "UNKNOWN"},
	}

	for _, tt := range tests {
		got := NormalizeCode(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeCode(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeSection(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"§ 3342", "3342"},
		{"§§ 100-200", "100-200"},
		{"3342(a)", "3342(a)"},
		{"  924(e)  ", "924(e)"},
		{"§600.4", "600.4"},
	}

	for _, tt := range tests {
		got := NormalizeSection(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeSection(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCanonicalID(t *testing.T) {
	tests := []struct {
		name         string
		jurisdiction string
		code         string
		parts        []string
		want         string
	}{
		{
			name:         "simple jurisdiction and code",
			jurisdiction: "CA",
			code:         "CIV",
			parts:        nil,
			want:         "CA:CIV",
		},
		{
			name:         "full hierarchy",
			jurisdiction: "CA",
			code:         "CIV",
			parts:        []string{"T02", "CH02", "§3342"},
			want:         "CA:CIV:T02:CH02:§3342"},
		{
			name:         "federal USC",
			jurisdiction: "US",
			code:         "USC",
			parts:        []string{"T18", "§924(e)"},
			want:         "US:USC:T18:§924(e)"},
		{
			name:         "normalize code",
			jurisdiction: "ca",
			code:         "pen",
			parts:        []string{"§1538.5"},
			want:         "CA:PEN:§1538.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanonicalID(tt.jurisdiction, tt.code, tt.parts...)
			if got != tt.want {
				t.Errorf("CanonicalID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseCitationString(t *testing.T) {
	tests := []struct {
		input string
		want  *Citation
	}{
		{
			input: "CIV § 3342",
			want: &Citation{
				Jurisdiction: "CA",
				Code:         "CIV",
				Section:      "3342",
			},
		},
		{
			input: "Penal Code § 1538.5",
			want: &Citation{
				Jurisdiction: "CA",
				Code:         "PEN",
				Section:      "1538.5",
			},
		},
		{
			input: "18 USC § 924(e)",
			want: &Citation{
				Jurisdiction: "US",
				Code:         "USC",
				Title:        "18",
				Section:      "924(e)",
			},
		},
		{
			input: "18 U.S.C. § 924(e)",
			want: &Citation{
				Jurisdiction: "US",
				Code:         "USC",
				Title:        "18",
				Section:      "924(e)",
			},
		},
		{
			input: "28 CFR § 600.4",
			want: &Citation{
				Jurisdiction: "US",
				Code:         "CFR",
				Title:        "28",
				Section:      "600.4",
			},
		},
		{
			input: "Cal. Const. art. I, § 13",
			want: &Citation{
				Jurisdiction: "CA",
				Code:         "CONS",
				Article:      "I",
				Section:      "13",
			},
		},
		{
			input: "U.S. Const. amend. IV",
			want: &Citation{
				Jurisdiction: "US",
				Code:         "CONST",
				Section:      "IV",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseCitationString(tt.input)
			if err != nil {
				t.Fatalf("ParseCitationString(%q) error: %v", tt.input, err)
			}
			if got.Jurisdiction != tt.want.Jurisdiction {
				t.Errorf("Jurisdiction = %q, want %q", got.Jurisdiction, tt.want.Jurisdiction)
			}
			if got.Code != tt.want.Code {
				t.Errorf("Code = %q, want %q", got.Code, tt.want.Code)
			}
			if got.Title != tt.want.Title {
				t.Errorf("Title = %q, want %q", got.Title, tt.want.Title)
			}
			if got.Section != tt.want.Section {
				t.Errorf("Section = %q, want %q", got.Section, tt.want.Section)
			}
			if got.Article != tt.want.Article {
				t.Errorf("Article = %q, want %q", got.Article, tt.want.Article)
			}
		})
	}
}

func TestCitationToCanonicalID(t *testing.T) {
	tests := []struct {
		name     string
		citation Citation
		want     string
	}{
		{
			name: "CA Civil Code section",
			citation: Citation{
				Jurisdiction: "CA",
				Code:         "CIV",
				Section:      "3342",
			},
			want: "CA:CIV:§3342",
		},
		{
			name: "US Code with title",
			citation: Citation{
				Jurisdiction: "US",
				Code:         "USC",
				Title:        "18",
				Section:      "924(e)",
			},
			want: "US:USC:T18:§924(e)",
		},
		{
			name: "CA Constitution with article",
			citation: Citation{
				Jurisdiction: "CA",
				Code:         "CONS",
				Article:      "I",
				Section:      "13",
			},
			want: "CA:CONS:ArtI:§13",
		},
		{
			name: "CFR with title",
			citation: Citation{
				Jurisdiction: "US",
				Code:         "CFR",
				Title:        "28",
				Section:      "600.4",
			},
			want: "US:CFR:T28:§600.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.citation.ToCanonicalID()
			if got != tt.want {
				t.Errorf("ToCanonicalID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	inputs := []string{
		"CIV § 3342",
		"18 USC § 924(e)",
		"28 CFR § 600.4",
		"Cal. Const. art. I, § 13",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			cit, err := ParseCitationString(input)
			if err != nil {
				t.Fatalf("ParseCitationString error: %v", err)
			}
			id := cit.ToCanonicalID()
			if id == "" {
				t.Errorf("ToCanonicalID() returned empty string for %q", input)
			}
			t.Logf("%s -> %s", input, id)
		})
	}
}
