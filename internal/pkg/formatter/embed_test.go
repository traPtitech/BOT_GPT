package formatter

import "testing"

func TestFormatEmbeds(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No embeds",
			input:    "Hello, world!",
			expected: "Hello, world!",
		},
		{
			name:     "Single embed",
			input:    `Hello, !{"type":"channel","raw":"#general","id":"04ad2c18-fdcb-4c43-beef-82e8ba26ac98"}, world`,
			expected: "Hello, #general, world",
		},
		{
			name:     "Multiple embeds",
			input:    `!{"type":"user","raw":"@cp20","id":"be77174f-13c5-4464-8b15-7f45b96d5b18"}!{"type":"channel","raw":"#general","id":"04ad2c18-fdcb-4c43-beef-82e8ba26ac98"}`,
			expected: "@cp20#general",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatEmbeds(tt.input)
			if result != tt.expected {
				t.Errorf("FormatEmbeds(%q) = %q; expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
